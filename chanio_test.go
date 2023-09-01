package chanio

import (
	"bytes"
	"io"
	"testing"
)

func TestPipe(t *testing.T) {
	r, wc, _ := Pipe()
	str := "Hello World!"
	go func() {
		wc.Write([]byte(str))
		wc.Close()
	}()
	var buf bytes.Buffer
	n, err := io.Copy(&buf, r)
	if int(n) != len(str) {
		t.Fatalf("expect n=%d, got=%d", len(str), n)
	}
	if err != nil {
		t.Fatalf("expect no error, got=%v", err)
	}
	if buf.String() != str {
		t.Fatalf("expect str=%q, got=%q", str, buf.String())
	}
}

func TestShortWrite(t *testing.T) {
	r, wc, _ := Pipe()
	str := "Hello World!"
	type result struct {
		n   int
		err error
	}
	resCh := make(chan result)
	go func() {
		n, err := wc.Write([]byte(str))
		resCh <- result{n, err}
	}()
	buf := make([]byte, len("Hello"))
	n, err := io.ReadFull(r, buf)
	if err != nil {
		t.Fatalf("expect no read error, got=%v", err)
	}
	if n != len("Hello") {
		t.Fatalf("expect read n=%d, got=%d", len("Hello"), n)
	}

	chanio := r.(ChanIO)
	chanio.Close()
	res := <-resCh

	if res.n != len("Hello") {
		t.Fatalf("expect write n=%d, got=%d", len("Hello"), res.n)
	}
	if res.err != io.ErrShortWrite {
		t.Fatalf("expect write err=%v, got=%v", io.ErrShortWrite, res.err)
	}
}
