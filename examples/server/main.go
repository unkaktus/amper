package main

import (
	"io"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/nogoegst/amper"
	"github.com/rs/zerolog/log"
)

type copier struct{}

func (c *copier) Handle(w io.Writer, r io.Reader) error {
	_, err := io.Copy(w, r)
	return err
}

func main() {
	server := &amper.Server{
		Handler: &copier{},
	}
	h := gziphandler.GzipHandler(server)

	if err := http.ListenAndServeTLS(":https", "/tls/fullchain.pem", "/tls/privkey.pem", h); err != nil {
		log.Fatal().Err(err).Msg("serve HTTP")
	}

}
