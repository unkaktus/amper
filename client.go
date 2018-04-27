// client.go - AMP tunnel client related functions.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package amper

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// Produce a random ID as a URL-safe Base64 string.
func randomID() string {
	b := make([]byte, 10)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

const (
	CDNDomain = "cdn.ampproject.org"
)

// hostToAMPHost transforms a DNS name into the name AMP CDN accepts.
func hostToAMPHost(h string) string {
	h = strings.Replace(h, "-", "--", -1)
	h = strings.Replace(h, ".", "-", -1)
	return h + "." + CDNDomain
}

func getAttribute(n *html.Node, key string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func getNodeByID(n *html.Node, id string) *html.Node {
	if n.Type == html.ElementNode {
		s, ok := getAttribute(n, "id")
		if ok && s == id {
			return n
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := getNodeByID(c, id)
		if result != nil {
			return result
		}
	}
	return nil
}

// extractData extracts payload from an AMP page body r.
func extractData(r io.Reader) (io.Reader, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	n := getNodeByID(doc, "data")
	if n == nil {
		err = errors.New("no element with this ID")
		return nil, err
	}
	// The node found but there is no child.
	if n.FirstChild == nil {
		return bytes.NewReader(nil), nil
	}
	data := n.FirstChild.Data
	b, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	br := bytes.NewReader(b)
	return br, nil
}

// assemblePath encodes data from reader r into URL path.
// The format is "/{random}/{payload}" where random is
// a random string to disable cache, and payload is the payload
// encoded into URL-safe Base64.
func assemblePath(r io.Reader) (string, error) {
	slug := randomID()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	req := base64.RawURLEncoding.EncodeToString(data)
	return path.Join(slug, req), nil
}

// Client desribes a client state.
type Client struct {
	// Host is the hostname of the backend to use.
	Host string
	// Front is the hostname sent in TLS SNI.
	Front string
	// Transport is the http.RoundTripper to use to perform requests.
	// If Transport is nil then http.DefaultTransport is used.
	Transport http.RoundTripper
}

// RoundTrip writes data from reader r to the server and returns
// reply from the server.
func (c *Client) RoundTrip(r io.Reader) (io.Reader, error) {
	reqPath, err := assemblePath(r)
	if err != nil {
		return nil, err
	}
	u := &url.URL{
		Scheme:   "https",
		Host:     c.Front,
		Path:     path.Join("v", "s", c.Host, reqPath),
		RawQuery: "amp_js_v=0.1",
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	// Rewrite Host header to the AMP one
	req.Host = hostToAMPHost(c.Host)

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
	data, err := extractData(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
