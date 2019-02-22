package main

import (
	"net"
	"time"
	"org/gtunnel/api"
)

type Wire struct {
	cfg *Cfg
	src net.Conn
	dst net.Conn

	fwb *api.RwBuf
	bwb *api.RwBuf

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
		cfg:    cfg,
		src:    src,
		dst:    dst,
		fwb:    api.MkRWBuf(BUF_SIZE),
		bwb:    api.MkRWBuf(BUF_SIZE),
		atime:  time.Now(),
		closed: false,
		fm:     NewMeter(),
		bm:     NewMeter(),
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

func (live *Liveness) Measure() (forward, backward, count uint64) {
	// remove dead wires
	W := make([]*Wire, 0)

	<-live.C
		for _, one := range live.W {
			if !one.closed {
				W = append(W, one)
			}
		}

		live.W = W
		N := len(W)
	live.C <- true

	for _, one := range W {
		_, wr0 := one.fm.Consume()
		forward += wr0

		_, wr1 := one.bm.Consume()
		backward += wr1
	}
	return forward, backward, uint64(N)
}
