package main

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nogoegst/amper"
	"github.com/rs/zerolog/log"
)

func ping(c *amper.Client) {
	req := &bytes.Buffer{}
	io.CopyN(req, rand.Reader, 1550)
	reqData := req.Bytes()

	resp, err := c.RoundTrip(req)
	if err != nil {
		log.Fatal().Err(err).Msg("perform round trip")
	}

	respData, err := ioutil.ReadAll(resp)
	if err != nil {
		log.Fatal().Err(err).Msg("read out response")
	}
	if diff := cmp.Diff(respData, reqData); diff != "" {
		log.Fatal().Msgf("response differs: %s", diff)
	}
}

func main() {
	c := &amper.Client{
		Host:  "amp.nogoegst.net",
		Front: "www.google.com",
	}

	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		go func() {
			start := time.Now()
			ping(c)
			rtt := time.Since(start)
			log.Info().Str("rtt", rtt.String()).Msg("got response")
		}()
	}
}
