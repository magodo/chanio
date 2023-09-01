package chanio

import (
	"io"
)

type ChanIO chan byte

var _ io.ReadWriteCloser = make(ChanIO)

func Pipe() (io.Reader, io.WriteCloser, error) {
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
Loop:
	for {
		if cnt == size {
			return cnt, nil
		}
		select {
		case b, ok := <-ch:
			if ok {
				buf = append(buf, b)
				cnt++
				continue
			}
			copy(p, buf)
			return cnt, io.EOF
		default:
			break Loop
		}
	}
	copy(p, buf)
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
