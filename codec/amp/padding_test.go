package ampcodec

import (
	"bytes"
	"io"
	"testing"

	"github.com/matryer/is"
)

func TestPaddingWriter(t *testing.T) {
	is := is.New(t)
	input := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	output := []byte("aaa aaa aaa aaa aaa aaa aaa aaa aaa aaa ")
	r1 := bytes.NewReader(input[:13])
	r2 := bytes.NewReader(input[13:])
	buf := &bytes.Buffer{}
	w := NewPaddingWriter(nopCloser(buf), " ", 3)
	_, err := io.Copy(w, r1)
	is.NoErr(err)
	_, err = io.Copy(w, r2)
	is.NoErr(err)
	is.Equal(buf.Bytes(), output)
}
