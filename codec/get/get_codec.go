// get_codec.go - GET request data codec
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package getcodec

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"io/ioutil"
	"path"
	"strings"
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

// Encode encodes data from reader r into URL path.
// The format is "/{random}/{payload}" where random is
// a random string to disable cache, and payload is the payload
// encoded into URL-safe Base64.
func Encode(r io.Reader) (string, error) {
	slug := randomID()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	req := base64.RawURLEncoding.EncodeToString(data)
	return path.Join(slug, req), nil
}

// Decode decodes request data from the path.
func Decode(path string) (*bytes.Reader, error) {
	sp := strings.Split(path, "/")
	b, err := base64.RawURLEncoding.DecodeString(sp[len(sp)-1])
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
