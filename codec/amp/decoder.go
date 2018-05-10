// decoder.go - AMP HTML data codec.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package ampcodec

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

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

// NewDecoder extracts payload from an AMP page body r.
func NewDecoder(r io.Reader) (io.ReadCloser, error) {
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
		return ioutil.NopCloser(bytes.NewReader(nil)), nil
	}
	data := n.FirstChild.Data
	// Remove all whitespaces
	data = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, data)
	b, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	br := bytes.NewReader(b)
	return ioutil.NopCloser(br), nil
}
