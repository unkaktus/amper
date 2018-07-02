// encoder.go - AMP HTML data codec.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package ampcodec

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

var (
	ampOldBoilerplate = "<style>body {opacity: 0}</style><noscript><style>body {opacity: 1}</style></noscript>"
	ampBoilerplate    = "<style amp-boilerplate>body{-webkit-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-moz-animation:-amp-start 8s steps(1,end) 0s 1 normal both;-ms-animation:-amp-start 8s steps(1,end) 0s 1 normal both;animation:-amp-start 8s steps(1,end) 0s 1 normal both}@-webkit-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-moz-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-ms-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@-o-keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}@keyframes -amp-start{from{visibility:hidden}to{visibility:visible}}</style><noscript><style amp-boilerplate>body{-webkit-animation:none;-moz-animation:none;-ms-animation:none;animation:none}</style></noscript>"
	ampHeaderFormat   = `<!doctype html>
<html amp>
  <head>
    <meta charset="utf-8">
    <script async src="https://cdn.ampproject.org/v0.js"></script>
    <title>amp</title>
    <link rel="canonical" href="#" />
    <meta name="viewport" content="width=device-width,minimum-scale=1,initial-scale=1">
    %s
  </head>
  <body>
    <p>In varietate concordia</p>
    <pre id="data">`
	ampTrailer = []byte(`</pre>
  </body>
</html>`)
)

var (
	ErrEncoderClosed = errors.New("encoder is already closed")
)

// Encoder is an instance of AMP HTML encoder.
type Encoder struct {
	w                io.Writer
	closed           uint32
	headerWritten    uint32
	trailerWritten   uint32
	dataEncoder      io.WriteCloser
	dataEncoderMutex sync.Mutex

	// UseOldBoilerplate sets Encoder to write
	// deprecated AMP boilerplate. As it's much shorter
	// than the new one, one may benefit from using it
	// to save some bandwidth.
	// Note that it may stop working in future.
	UseOldBoilerplate bool
}

// Close signals Encoder that there will be no data so it may write
// trailer.
func (enc *Encoder) Close() (err error) {
	enc.dataEncoderMutex.Lock()
	if enc.dataEncoder != nil {
		err = enc.dataEncoder.Close()
	}
	enc.dataEncoderMutex.Unlock()
	if err != nil {
		return err
	}
	// Write AMP trailer if we wrote the header.
	if atomic.LoadUint32(&enc.headerWritten) == 1 &&
		atomic.LoadUint32(&enc.trailerWritten) == 0 {
		_, err = enc.w.Write(ampTrailer)
		if err != nil {
			return err
		}
	}
	atomic.StoreUint32(&enc.closed, 1)
	return nil
}

func NewEncoder(w io.Writer) *Encoder {
	enc := &Encoder{
		w: w,
	}
	return enc
}

func (enc *Encoder) Write(p []byte) (n int, err error) {
	if atomic.LoadUint32(&enc.closed) == 1 {
		return 0, ErrEncoderClosed
	}
	// Write AMP header it we haven't
	if atomic.LoadUint32(&enc.headerWritten) == 0 {
		if enc.UseOldBoilerplate {
			_, err = fmt.Fprintf(enc.w, ampHeaderFormat, ampOldBoilerplate)
		} else {
			_, err = fmt.Fprintf(enc.w, ampHeaderFormat, ampBoilerplate)
		}
		atomic.StoreUint32(&enc.headerWritten, 1)
		if err != nil {
			return 0, err
		}
	}
	enc.dataEncoderMutex.Lock()
	if enc.dataEncoder == nil {
		w, ok := enc.w.(io.WriteCloser)
		if !ok {
			w = NopCloser(enc.w)
		}
		padWriter := NewPaddingWriter(w, " ", 32)
		enc.dataEncoder = base64.NewEncoder(base64.RawURLEncoding, padWriter)
	}
	enc.dataEncoderMutex.Unlock()
	return enc.dataEncoder.Write(p)
}
