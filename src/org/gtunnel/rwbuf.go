package main

import "fmt"

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

func roundUB(p uint64, cap uint64) uint64 {
	return p - p % cap
}

func (rwb *rwbuf) invariant() {
	assert(rwb.lb <= rwb.ub && rwb.ub <= rwb.lb + rwb.cap,
		fmt.Sprintf("bad buffer, [%d %d) cap=%d\n", rwb.lb, rwb.ub, rwb.cap))
}

func (rwb *rwbuf) Consumable() bool {
	return rwb.lb < rwb.ub
}

func (rwb *rwbuf) ConsumerBuffer() []byte {
	rwb.invariant()

	lb := rwb.lb % rwb.cap
	ub := rwb.ub % rwb.cap
	if rwb.lb / rwb.cap != rwb.ub / rwb.cap {
		ub = rwb.cap
	}
	return rwb.data[lb: ub]
}

func (rwb *rwbuf) Consume(n int) bool {
	rwb.invariant()
	rwb.lb += uint64(n)
	rwb.invariant()

	// normalize
	if rwb.ub == rwb.lb {
		rwb.ub = 0
		rwb.lb = 0
	}
	return rwb.Consumable()
}

func (rwb *rwbuf) Producible() bool {
	return rwb.ub < rwb.lb + rwb.cap
}

func (rwb *rwbuf) ProducerBuffer() []byte {
	lb := rwb.ub % rwb.cap
	sz := roundUB(rwb.lb + rwb.cap, rwb.cap) - rwb.ub

	return rwb.data[lb: (lb + sz)]
}

func (rwb *rwbuf) Produce(n int) bool {
	rwb.invariant()
	rwb.ub += uint64(n)
	rwb.invariant()

	return rwb.Producible()
}