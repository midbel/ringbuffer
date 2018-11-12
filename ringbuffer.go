package ringbuffer

type Ring struct {
	buffer []byte

	rix   int
	rchan chan int

	wix   int
	wchan chan int
}

func NewRing(t int) *Ring {
	return NewRingSize(32<<10, t)
}

func NewRingSize(s, t int) *Ring {
	r := Ring{
		buffer: make([]byte, s),
		rchan:  make(chan int),
		wchan:  make(chan int),
	}
	go r.syncboth(t)
	return &r
}

func (r *Ring) Read(bs []byte) (int, error) {
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

func (r *Ring) Write(bs []byte) (int, error) {
	n := copy(r.buffer[r.wix:], bs)
	if n < len(bs) {
		r.wix = copy(r.buffer, bs[n:])
	} else {
		r.wix += n
	}
	r.wchan <- len(bs)
	return len(bs), nil
}

func (r *Ring) syncboth(threshold int) {
	var available int
	for {
		select {
		case n, _ := <-r.wchan:
			available += n
		case n, _ := <-r.rchan:
			if n < available {
				available -= n
				r.rchan <- n
				break
			}
			for nn := range r.wchan {
				available += nn
				if n < available {
					available -= n
					r.rchan <- n
					break
				}
			}
		}
	}
}
