package main

import (
	"io"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/nogoegst/amper"
	_ "github.com/nogoegst/cabin/magic"
	"github.com/rs/zerolog/log"
)

func main() {
	server := &amper.Server{
		Handler: amper.HandlerFunc(func(w io.Writer, r io.Reader) error {
			_, err := io.Copy(w, r)
			return err
		}),
	}
	h := gziphandler.GzipHandler(server)

	// We listen at port 80 here because we run Caddy to manage TLS certs
	if err := http.ListenAndServe(":http", h); err != nil {
		log.Fatal().Err(err).Msg("serve HTTP")
	}

}
