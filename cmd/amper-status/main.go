package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nogoegst/amper"
	"github.com/rs/zerolog/log"
)

func ping(c *amper.Client, payloadSize int64) error {
	req := &bytes.Buffer{}
	io.CopyN(req, rand.Reader, payloadSize)
	reqData := req.Bytes()

	resp, err := c.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("perform round trip: %w", err)
	}

	respData, err := ioutil.ReadAll(resp)
	if err != nil {
		return fmt.Errorf("read out response: %w", err)
	}
	if diff := cmp.Diff(respData, reqData); diff != "" {
		return fmt.Errorf("response differs: %s", diff)
	}
	return nil
}

type Status struct {
	Works       bool
	AmperHost   string
	FrontDomain string
	PayloadSize int64
	RTT         string
	DateChecked string
}

var status = Status{
	Works:       false,
	AmperHost:   "n/a",
	FrontDomain: "n/a",
	RTT:         "n/a",
	DateChecked: "not checked",
}

var statusPageTemplate = `<html>
<head>
	<title>amper status</title>
</head>
<body>
	<pre>
	works: {{ .Works }}
	amper host: {{ .AmperHost }}
	front domain: {{ .FrontDomain }}
	payload size: {{ .PayloadSize }}
	rtt: {{ .RTT }}
	date checked: {{ .DateChecked }}
	</pre>
</body>
</html>
`

func statusPageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("status.html").Parse(statusPageTemplate))
		err := t.Execute(w, status)
		if err != nil {
			log.Error().Err(err).Msg("execute template")
		}
	})
}

func main() {
	var payloadSize = flag.Int64("payload-size", 1550, "size of echo payload")
	var host = flag.String("host", "amp.nogoegst.net", "AMP host (amper-server)")
	var front = flag.String("front", "www.google.com", "Fronting domain")
	var interval = flag.Duration("interval", time.Second, "Ping interval")
	flag.Parse()

	c := &amper.Client{
		Host:  *host,
		Front: *front,
	}

	status.AmperHost = *host
	status.FrontDomain = *front
	status.PayloadSize = *payloadSize

	ticker := time.NewTicker(*interval)
	go func() {
		for {
			start := time.Now()
			err := ping(c, *payloadSize)
			rtt := time.Since(start)
			log.Info().Str("rtt", rtt.String()).Msg("got response")
			if err != nil {
				log.Error().Err(err).Msg("ping")
			}
			status.Works = err == nil
			status.RTT = rtt.String()
			status.DateChecked = time.Now().UTC().Format(time.RFC822)

			<-ticker.C
		}
	}()

	h := statusPageHandler()
	if err := http.ListenAndServe(":http", h); err != nil {
		log.Fatal().Err(err).Msg("serve HTTP")
	}
}
