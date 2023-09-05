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
	var unblk bool
	for {
		if cnt == size {
			copy(p, buf)
			return cnt, nil
		}

		if unblk {
			// Non-first read from channel, which should continue until blocked/channel closed.
			select {
			case b, ok := <-ch:
				if ok {
					buf = append(buf, b)
					cnt++
					continue
				}
				// channel is closed
				copy(p, buf)
				return cnt, io.EOF
			default:
				// short read
				copy(p, buf)
				return cnt, nil
			}
		}

		// First read from channel that might block.
		b, ok := <-ch
		if ok {
			unblk = true
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
