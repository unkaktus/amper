// server.go - AMP tunnel server related functions.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package amper

import (
	"io"
	"net/http"

	ampcodec "github.com/nogoegst/amper/codec/amp"
	getcodec "github.com/nogoegst/amper/codec/get"
)

// Handler is the interface for request handler, i.e.
// application-level logic handler.
type Handler interface {
	// Handle handles incoming data from r and writes
	// responses to w.
	Handle(w io.Writer, r io.Reader) error
}

type handlerFunc struct {
	hf func(w io.Writer, r io.Reader) error
}

func (h handlerFunc) Handle(w io.Writer, r io.Reader) error {
	return h.hf(w, r)
}

// HandlerFunc wraps a handler function hf into Handler.
func HandlerFunc(hf func(w io.Writer, r io.Reader) error) Handler {
	return handlerFunc{hf: hf}
}

// Server is an http.Handler that handles amper requests over
// AMP pages.
type Server struct {
	// Handler to handle requests
	Handler Handler
	// UseOldBoilerplate makes AMP encoder use
	// deprecated AMP boilerplate. As it's much shorter
	// than the new one, one may benefit from using it
	// to save some bandwidth.
	// Note that it doesn't work on Google AMP cache anymore.
	UseOldAMPBoilerplate bool
}

func (ah *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// We want to give a hint to AMP cache that this is request
	// will not be used anymore.
	w.Header().Set("Cache-Control", "private, max-age=0")

	// We always write AMP page even if it has no useful data.
	enc := ampcodec.NewEncoder(w)
	enc.UseOldBoilerplate = ah.UseOldAMPBoilerplate
	defer enc.Close()
	// We do not throw any HTTP errors because clients are not going
	// to get them anyway (because of the cache middleware).
	req, err := getcodec.Decode(r.URL.Path)
	if err != nil {
		return
	}
	err = ah.Handler.Handle(enc, req)
	if err != nil {
		return
	}

}
