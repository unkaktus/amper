package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
	"github.com/unkaktus/amper"
)

func ping(c *amper.Client, payloadSize int64) error {
	req := &bytes.Buffer{}
	io.CopyN(req, rand.Reader, payloadSize)
	reqData := req.Bytes()

	resp, err := c.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("perform round trip: %w", err)
	}

	respData, err := io.ReadAll(resp)
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

var worksBadge = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="86" height="20" role="img" aria-label="status: works"><title>status: works</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="86" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="43" height="20" fill="#555"/><rect x="43" width="43" height="20" fill="#97ca00"/><rect width="86" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="225" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="330">status</text><text x="225" y="140" transform="scale(.1)" fill="#fff" textLength="330">status</text><text aria-hidden="true" x="635" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="330">works</text><text x="635" y="140" transform="scale(.1)" fill="#fff" textLength="330">works</text></g></svg>`
var brokenBadge = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="92" height="20" role="img" aria-label="status: broken"><title>status: broken</title><linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="r"><rect width="92" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#r)"><rect width="43" height="20" fill="#555"/><rect x="43" width="49" height="20" fill="#e05d44"/><rect width="92" height="20" fill="url(#s)"/></g><g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110"><text aria-hidden="true" x="225" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="330">status</text><text x="225" y="140" transform="scale(.1)" fill="#fff" textLength="330">status</text><text aria-hidden="true" x="665" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="390">broken</text><text x="665" y="140" transform="scale(.1)" fill="#fff" textLength="390">broken</text></g></svg>`

var statusPageHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("status.html").Parse(statusPageTemplate))
	err := t.Execute(w, status)
	if err != nil {
		log.Error().Err(err).Msg("execute template")
	}
})

var statusBadgeHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml;charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	if status.Works {
		w.Write([]byte(worksBadge))
	} else {
		w.Write([]byte(brokenBadge))
	}
})

func main() {
	payloadSize := flag.Int64("payload-size", 1550, "size of echo payload")
	host := flag.String("host", "amp.unkaktus.art", "AMP host (amper-server)")
	front := flag.String("front", "www.google.com", "Fronting domain")
	interval := flag.Duration("interval", time.Second, "Ping interval")
	listenAddress := flag.String("l", ":http", "Address to listen on, in format hostname:port")
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

	h := http.NewServeMux()
	h.Handle("/status.svg", statusBadgeHandler)
	h.Handle("/status", statusPageHandler)

	if err := http.ListenAndServe(*listenAddress, h); err != nil {
		log.Fatal().Err(err).Msg("serve HTTP")
	}
}
