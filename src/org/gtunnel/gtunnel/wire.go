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
	return live
}

func (live *Liveness) Add(wire *Wire)  {
	live.C <- true
	live.W = append(live.W, wire)
	<- live.C
}

func (live *Liveness) Measure() (uint64, uint64, int) {
	c := make(chan *Wire)

	go func() {
		live.C <- true
		lastGood := -1
		for i, one := range live.W {
			c <- one

			if !one.closed {
				lastGood += 1
				live.W[i], live.W[lastGood] = live.W[lastGood], live.W[i]
			}
		}
		live.W = live.W[0: lastGood + 1]
		<-live.C

		close(c)
	}()

	forward, backward, N := uint64(0), uint64(0), 0
	for one := range c {
		_, f := one.fm.Consume()
		_, b := one.bm.Consume()

		forward += f
		backward += b
		N += 1
	}

	return forward, backward, N
}
