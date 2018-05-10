// nopcloser.go
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package ampcodec

import (
	"io"
)

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
}

func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}
