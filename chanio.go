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
	for {
		if cnt == size {
			copy(p, buf)
			return cnt, nil
		}
		b, ok := <-ch
		if ok {
			buf = append(buf, b)
			cnt++
			continue
		}
		// channel is closed
		copy(p, buf)
		return cnt, io.EOF
	}
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
