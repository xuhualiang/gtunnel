package main

type rwbuf struct {
	data []byte
	cap  uint64
	lb   uint64
	ub   uint64
}

func NewRWBuf(cap uint64) *rwbuf {
	return &rwbuf{
		data: make([]byte, cap),
		cap:  cap,
		lb:   0,
		ub:   0,
	}
}

func (rwb *rwbuf) invariant() {
	assert(rwb.lb <= rwb.ub && rwb.ub <= rwb.lb + rwb.cap, "bad buffer")
}

func (rwb *rwbuf) Reader() []byte {
	rwb.invariant()
	lb := rwb.lb % rwb.cap
	ub := rwb.ub - (rwb.lb - lb /*shift*/)

	return rwb.data[lb: ub]
}

func (rwb *rwbuf) Read(n uint64) bool {
	rwb.invariant()
	rwb.lb += n
	rwb.invariant()
	return rwb.lb < rwb.ub
}

func (rwb *rwbuf) Writter() []byte {
	return rwb.data[rwb.ub % rwb.cap: rwb.lb % rwb.cap]
}

func (rwb *rwbuf) Write(n uint64) bool {
	rwb.invariant()
	rwb.ub += n
	rwb.invariant()
	return rwb.ub < rwb.lb + rwb.cap
}