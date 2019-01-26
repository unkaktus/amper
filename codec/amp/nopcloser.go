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

type nop struct {
	io.Writer
}

func (nop) Close() error {
	return nil
}

func nopCloser(w io.Writer) io.WriteCloser {
	return nop{w}
}
