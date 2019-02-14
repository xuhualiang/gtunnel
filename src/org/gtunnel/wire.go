package main

import (
	"net"
	"time"
	"fmt"
)

type Wire struct {
	cfg *Cfg
	src net.Conn
	dst net.Conn

	fwb *rwbuf
	bwb *rwbuf

	atime  time.Time
	closed bool

	fm *meter
	bm *meter
}

func (wire *Wire) Close() {
	wire.closed = true
	wire.src.Close()
	wire.dst.Close()
}

func mkWire(cfg *Cfg, src net.Conn, dst net.Conn) *Wire {
	wire := &Wire{
		cfg:   cfg,
		src:   src,
		dst:   dst,
		fwb:   NewRWBuf(BUF_SIZE),
		bwb:   NewRWBuf(BUF_SIZE),
		atime: time.Now(),
		closed: false,
		fm:    NewMeter(),
		bm:    NewMeter(),
	}

	return wire
}

func (wire *Wire) Touch()  {
	wire.atime = time.Now()
}

func (wire Wire) String() string {
	return wire.cfg.String()
}

type Liveness struct {
	W []*Wire
	C chan bool
}

func NewLiveness() *Liveness {
	live := &Liveness{
		W: make([]*Wire, 0),
		C: make(chan bool, 1),
	}
	live.C <- true
	return live
}

func (live *Liveness) Add(wire *Wire)  {
	<-live.C
		live.W = append(live.W, wire)
	live.C <- true
}

func (live *Liveness) Measure() (forward, backward uint64) {
	// remove dead wires
	<-live.C
		lastGood := -1
		for i := 0; i < len(live.W); i++ {
			if !live.W[i].closed {
				lastGood += 1
				live.W[lastGood], live.W[i] = live.W[i], live.W[lastGood]
			}
		}

		live.W = live.W[0:lastGood + 1]
		W := live.W[0: len(live.W)]
	live.C <- true

	for _, w := range W {
		_, wr0 := w.fm.Consume()
		forward += wr0

		_, wr1 := w.bm.Consume()
		backward += wr1
	}
	return
}