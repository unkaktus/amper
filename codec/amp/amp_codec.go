// amp_codec.go - AMP HTML data codec.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package ampcodec

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

	"golang.org/x/net/html"
)

var (
	ampHeader = []byte(`<!doctype html>
<html amp>
  <head>
    <meta charset="utf-8">
    <title>amp</title>
    <link rel="canonical" href="https://ampproject.org" />
    <meta name="viewport" content="width=device-width,minimum-scale=1,initial-scale=1">
    <style>body {opacity: 0}</style><noscript><style>body {opacity: 1}</style></noscript>
    <script async src="https://cdn.ampproject.org/v0.js"></script>
  </head>
  <body>
    <p>In varietate concordia</p>
    <pre id="data">`)
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
		_, err = enc.w.Write(ampHeader)
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

/******/

func getAttribute(n *html.Node, key string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func getNodeByID(n *html.Node, id string) *html.Node {
	if n.Type == html.ElementNode {
		s, ok := getAttribute(n, "id")
		if ok && s == id {
			return n
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := getNodeByID(c, id)
		if result != nil {
			return result
		}
	}
	return nil
}

// NewDecoder extracts payload from an AMP page body r.
func NewDecoder(r io.Reader) (io.ReadCloser, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	n := getNodeByID(doc, "data")
	if n == nil {
		err = errors.New("no element with this ID")
		return nil, err
	}
	// The node found but there is no child.
	if n.FirstChild == nil {
		return ioutil.NopCloser(bytes.NewReader(nil)), nil
	}
	data := n.FirstChild.Data
	// Remove all whitespaces
	data = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, data)
	b, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	br := bytes.NewReader(b)
	return ioutil.NopCloser(br), nil
}
