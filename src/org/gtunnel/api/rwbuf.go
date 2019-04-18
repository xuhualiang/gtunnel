package api

import (
	"fmt"
)

type RwBuf struct {
	data []byte
	cap  uint64
	lb   uint64
	ub   uint64
}

func MkRWBuf(cap uint64) *RwBuf {
	return &RwBuf{
		data: make([]byte, cap),
		cap:  cap,
		lb:   0,
		ub:   0,
	}
}

func roundUB(p uint64, cap uint64) uint64 {
	return p - p % cap
}

func (rwb *RwBuf) invariant() {
	Assert(rwb.lb <= rwb.ub && rwb.ub <= rwb.lb + rwb.cap,
		fmt.Sprintf("bad buffer, [%d %d) cap=%d\n", rwb.lb, rwb.ub, rwb.cap))
}

func (rwb *RwBuf) Consumable() bool {
	return rwb.lb < rwb.ub
}

func (rwb *RwBuf) ConsumerBuffer() []byte {
	rwb.invariant()

	lb := rwb.lb % rwb.cap
	ub := rwb.ub % rwb.cap
	if rwb.lb / rwb.cap != rwb.ub / rwb.cap {
		ub = rwb.cap
	}
	return rwb.data[lb: ub]
}

func (rwb *RwBuf) Consume(n int) bool {
	rwb.lb += uint64(n)
	rwb.invariant()

	// normalize
	if rwb.ub == rwb.lb {
		rwb.ub = 0
		rwb.lb = 0
	}
	return rwb.Consumable()
}

func (rwb *RwBuf) Producible() bool {
	return rwb.ub < rwb.lb + rwb.cap
}

func (rwb *RwBuf) ProducerBuffer() []byte {
	lb := rwb.ub % rwb.cap
	sz := roundUB(rwb.lb + rwb.cap, rwb.cap) - rwb.ub

	return rwb.data[lb: (lb + sz)]
}

func (rwb *RwBuf) Produce(n int) bool {
	rwb.ub += uint64(n)
	rwb.invariant()

	return rwb.Producible()
}

func (rwb *RwBuf) Close()  {
	rwb.data = nil
}