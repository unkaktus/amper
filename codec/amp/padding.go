// padding.go
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of amper, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package ampcodec

import (
	"io"
)

// PaddingWriter in an io.WriteCloser wrapper
// which inserts pads each step.
type PaddingWriter struct {
	w    io.WriteCloser
	pad  string
	step int

	// written bytes state
	n int
}

// NewPaddingWriter creates a PaddingWriter which writes pad each step bytes.
func NewPaddingWriter(w io.WriteCloser, pad string, step int) *PaddingWriter {
	pw := &PaddingWriter{
		w:    w,
		pad:  pad,
		step: step,
	}
	return pw
}

func (pw *PaddingWriter) Write(p []byte) (int, error) {
	nn := 0
	for len(p) != 0 {
		// count bytes before pad
		l := pw.step - pw.n
		wp := p
		// data remainder is longer than desired length
		if l < len(p) {
			wp = p[:l]
		}
		n, err := pw.w.Write(wp)
		nn += n
		pw.n += n
		// step count
		c := pw.n / pw.step
		if c != 0 {
			_, err = pw.w.Write([]byte(pw.pad))
			if err != nil {
				return nn, err
			}
		}
		// shrink state
		pw.n = pw.n - c*pw.step
		// drop written data from p
		p = p[len(wp):]
		if err != nil {
			return nn, err
		}
	}
	return nn, nil
}

// Close closes underlying io.WriteCloser.
// It's an error to Write after Close.
func (pw *PaddingWriter) Close() error {
	return pw.w.Close()
}
