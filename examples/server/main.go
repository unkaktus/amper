package main

import (
	"io"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/nogoegst/amper"
	_ "github.com/nogoegst/cabin/magic"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	server := &amper.Server{
		Handler: amper.HandlerFunc(func(w io.Writer, r io.Reader) error {
			_, err := io.Copy(w, r)
			return err
		}),
	}
	h := gziphandler.GzipHandler(server)

	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("amp.nogoegst.net"),
		Cache:      autocert.DirCache("acme-cache"),
	}

	httpServer := &http.Server{
		Addr:      ":https",
		TLSConfig: m.TLSConfig(),
		Handler:   h,
	}

	if err := httpServer.ListenAndServeTLS("", ""); err != nil {
		log.Fatal().Err(err).Msg("serve HTTP")
	}

}
