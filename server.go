// server.go - AMP tunnel server related functions.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package amper

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
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

type Handler interface {
	Handle(w io.Writer, r io.Reader) error
}

// decodeRequest decodes request data from the path.
func decodeRequest(path string) ([]byte, error) {
	sp := strings.SplitN(path, "/", 3)
	if len(sp) < 3 {
		return nil, errors.New("path is too short")
	}
	b, err := base64.RawURLEncoding.DecodeString(sp[2])
	if err != nil {
		return nil, err
	}
	return b, nil
}

var (
	// Google AMP subnets
	allowedSubnets = []string{
		"66.249.64.0/19",
		"66.102.0.0/20",
	}
)

type Server struct {
	// Handler to handle requests
	Handler Handler
	// Allow non-AMP remotes
	AllowAllRemotes bool
}

func (ah *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We do not throw any HTTP errors because clients are not going
	// to get them anyway (because of the cache middleware).
	var invalid = false
	data, err := decodeRequest(r.URL.Path)
	if err != nil {
		invalid = true
	}
	// We want to give a hint to AMP cache that this is request
	// will not be used anymore.
	w.Header().Set("Cache-Control", "private, max-age=0")

	// Write AMP header
	_, err = w.Write(ampHeader)
	if err != nil {
		return
	}

	// Write the data itself. If data is invalid we send
	// empty HTML tag which will be trimmed by AMP as non-printable.
	if !invalid {
		enc := base64.NewEncoder(base64.RawURLEncoding, w)
		ah.Handler.Handle(enc, bytes.NewReader(data))
		enc.Close()
	}

	// Write AMP trailer.
	_, err = w.Write(ampTrailer)
	if err != nil {
		return
	}
}
