package main

import (
	"bytes"
	"io/ioutil"

	"github.com/nogoegst/amper"
	"github.com/rs/zerolog/log"
)

func main() {
	c := &amper.Client{
		Host:  "amp.nogoegst.net",
		Front: "www.google.com",
	}

	req := &bytes.Buffer{}
	_, err := req.Write([]byte("request data"))
	if err != nil {
		log.Fatal().Err(err).Msg("write payload to a buffer")
	}

	resp, err := c.RoundTrip(req)
	if err != nil {
		log.Fatal().Err(err).Msg("perform round trip")
	}

	data, err := ioutil.ReadAll(resp)
	if err != nil {
		log.Fatal().Err(err).Msg("read out response")
	}
	log.Printf("response: %s", data)
}
