package ringbuffer

import (
  "io"
)

type ring struct {
	buffer []byte

	rix   int
	rchan chan int

	wix   int
	wchan chan int
}

func Ring(s int, rs io.Reader) io.ReadWriter {
	r := ring{
		buffer: make([]byte, s),
		rchan:  make(chan int),
		wchan:  make(chan int),
	}
	go r.syncboth(s / 2)
	go r.fill(rs)
	return &r
}

func (r *ring) Read(bs []byte) (int, error) {
	r.rchan <- len(bs)
	<-r.rchan
	n := copy(bs, r.buffer[r.rix:])
	if n < len(bs) {
		r.rix = copy(bs[n:], r.buffer)
	} else {
		r.rix += n
	}
	return len(bs), nil
}

func (r *ring) Write(bs []byte) (int, error) {
	n := copy(r.buffer[r.wix:], bs)
	if n < len(bs) {
		r.wix = copy(r.buffer, bs[n:])
	} else {
		r.wix += n
	}
	r.wchan <- len(bs)
	return len(bs), nil
}

func (r *ring) fill(rs io.Reader) {
  io.Copy(r, rs)
}

func (r *ring) syncboth(threshold int) {
	var available int
	for {
		select {
		case n, _ := <-r.wchan:
			available += n
			for available < threshold {
				nn := <-r.wchan
				available += nn
			}
		case n, _ := <-r.rchan:
			if n < available {
				available -= n
				r.rchan <- n
				break
			}
			for nn := range r.wchan {
				available += nn
				for available < threshold {
					nn := <-r.wchan
					available += nn
				}
				if n < available {
					available -= n
					r.rchan <- n
					break
				}
			}
		}
	}
}
