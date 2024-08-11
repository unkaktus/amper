package main

import (
	"flag"
	"io"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/rs/zerolog/log"
	"github.com/unkaktus/amper"
	_ "github.com/unkaktus/cabin/magic"
)

func main() {
	listenAddress := flag.String("l", ":http", "Address to listen on, in format hostname:port")
	flag.Parse()

	server := &amper.Server{
		Handler: amper.HandlerFunc(func(w io.Writer, r io.Reader) error {
			_, err := io.Copy(w, r)
			return err
		}),
	}
	h := gziphandler.GzipHandler(server)

	// We listen at port 80, the TLS certs are managed by the frontend server
	if err := http.ListenAndServe(*listenAddress, h); err != nil {
		log.Fatal().Err(err).Msg("serve HTTP")
	}

}
