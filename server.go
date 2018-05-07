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

type Handler interface {
	Handle(w io.Writer, r io.Reader) error
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
	// We want to give a hint to AMP cache that this is request
	// will not be used anymore.
	w.Header().Set("Cache-Control", "private, max-age=0")

	// We always write AMP page even if it has no useful data.
	enc := ampcodec.NewEncoder(w)
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
