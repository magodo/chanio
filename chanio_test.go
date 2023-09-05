package chanio

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

// This is a smoking test for the typical copy use case.
func TestPipe(t *testing.T) {
	r, wc, _ := Pipe()
	str := "Hello World!"
	go func() {
		io.Copy(wc, bytes.NewBuffer([]byte(str)))
		wc.Close()
	}()
	var buf bytes.Buffer
	n, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("expect no error, got=%v", err)
	}
	if int(n) != len(str) {
		t.Fatalf("expect n=%d, got=%d", len(str), n)
	}
	if buf.String() != str {
		t.Fatalf("expect str=%q, got=%q", str, buf.String())
	}
}

// This is testing the Writer will return the io.ErroShortWrite when the read end is closed.
func TestShortWrite(t *testing.T) {
	r, w, _ := Pipe()
	str := "Hello World!"
	type result struct {
		n   int
		err error
	}
	resCh := make(chan result)
	go func() {
		n, err := io.Copy(w, bytes.NewBuffer([]byte(str)))
		resCh <- result{int(n), err}
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

// This is testing the Writer won't return successfully with writing less than expected.
// In other words, the Writer will always block on writing until the required buffer is fully written.
func TestWriteKeepWriting(t *testing.T) {
	r, w, _ := Pipe()
	str := "Hello World!"

	go func() {
		// At most read "Hello"
		buffer := make([]byte, len("Hello"))
		r.Read(buffer)

		// Delay a while to try to block the write end
		time.Sleep(time.Millisecond * 100)

		// Read the left part
		var buf bytes.Buffer
		io.Copy(&buf, r)
	}()

	n, err := w.Write([]byte(str))
	if err != nil {
		t.Fatalf("expect no error, got=%v", err)
	}
	if int(n) != len(str) {
		t.Fatalf("expect n=%d, got=%d", len(str), n)
	}
	w.Close()
}

// This is testing the Reader can read less than expected and return without blocking.
func TestShortRead(t *testing.T) {
	r, w, _ := Pipe()
	str := "Hello World!"

	go func() {
		w.Write([]byte(str))
	}()

	buffer := make([]byte, 100)
	n, err := r.Read(buffer)
	if err != nil {
		t.Fatalf("expect no error, got=%v", err)
	}
	if n <= 0 {
		t.Fatalf("expect n > 0, got=%d", n)
	}
	if int(n) > len(str) {
		t.Fatalf("expect n<=%d, got=%d", len(str), n)
	}
	if v := string(buffer[:n]); !strings.HasPrefix(str, v) {
		t.Fatalf("read unexpected content %q (str: %q)", v, str)
	}
	w.Close()
}
