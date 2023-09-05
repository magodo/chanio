package chanio

import (
	"io"
)

type ChanIO chan byte

var _ io.ReadWriteCloser = make(ChanIO)

func Pipe() (io.ReadCloser, io.WriteCloser, error) {
	ch := make(ChanIO)
	return ch, ch, nil
}

func (ch ChanIO) Read(p []byte) (n int, err error) {
	size := len(p)
	if size == 0 {
		return 0, nil
	}
	var cnt int
	var buf []byte
	if cnt == size {
		copy(p, buf)
		return cnt, nil
	}

	chOk := false

	// block the first time
	b, chOk := <-ch
	if chOk {
		buf = append(buf, b)
		cnt++
	}

	for len(ch) > 0 {
		b, chOk = <-ch
		if chOk {
			buf = append(buf, b)
			cnt++
		}
		copy(p, buf)
		if cnt == size {
			return cnt, nil
		}
	}

	copy(p, buf)

	if !chOk {
		return cnt, io.EOF
	}
	return cnt, nil
}

func (c ChanIO) Write(p []byte) (n int, err error) {
	var cnt int
	defer func() {
		if r := recover(); r != nil {
			n = cnt
			err = io.ErrShortWrite
			return
		}
	}()
	for _, b := range p {
		c <- b
		cnt++
	}
	return len(p), nil
}

func (c ChanIO) Close() (err error) {
	close(c)
	return nil
}
