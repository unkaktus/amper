// client.go - AMP tunnel client related functions.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package amper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	ampcodec "github.com/nogoegst/amper/codec/amp"
	getcodec "github.com/nogoegst/amper/codec/get"
)

const (
	DefaultCDNDomain = "cdn.ampproject.org"
)

// hostToAMPHost transforms a DNS name into the name AMP CDN accepts.
// If cdnDomain is empty, DefaultCDNDomain is used.
func hostToAMPHost(cdnDomain, host string) string {
	if cdnDomain == "" {
		cdnDomain = DefaultCDNDomain
	}
	host = strings.Replace(host, "-", "--", -1)
	host = strings.Replace(host, ".", "-", -1)
	return host + "." + cdnDomain
}

// Client desribes a client state.
type Client struct {
	// Host is the hostname of the backend to use.
	Host string
	// Front is the hostname sent in TLS SNI.
	Front string
	// Path is the prefix path for making requests.
	Path string
	// Transport is the http.RoundTripper to use to perform requests.
	// If Transport is nil then http.DefaultTransport is used.
	Transport http.RoundTripper
	// CDNDomain is the domain suffix of the AMP CDN.
	// If empty, DefaultCDNDomain is used.
	CDNDomain string
}

// RoundTrip writes data from reader r to the server and returns
// reply from the server.
func (c *Client) RoundTrip(r io.Reader) (io.ReadCloser, error) {
	reqPath, err := getcodec.Encode(r)
	if err != nil {
		return nil, err
	}
	// Compile plain URL
	u := &url.URL{
		Scheme:   "https",
		Host:     c.Host,
		Path:     path.Join(c.Path, reqPath),
		RawQuery: "amp_js_v=0.1",
	}
	// Do domain fronting
	if c.Front != "" {
		u.Host = c.Front
		u.Path = path.Join("v", "s", c.Host, u.Path)
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Rewrite Host header to the AMP one
	if c.Front != "" {
		req.Host = hostToAMPHost(c.CDNDomain, c.Host)
	}

	transport := http.DefaultTransport
	if c.Transport != nil {
		transport = c.Transport
	}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code %d", resp.StatusCode)
	}
	data, err := ampcodec.NewDecoder(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
