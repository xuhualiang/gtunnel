package main

import (
	"net"
	"time"
	"org/gtunnel/api"
	"fmt"
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

type Metrics struct {
	forward  uint64
	backward uint64
	n        int
}

func (m *Metrics) Aggregate(f, b uint64)  {
	m.forward += f
	m.backward += b
	m.n += 1
}

func (m *Metrics) Equas(other *Metrics) bool {
	return m.forward == other.forward &&
		m.backward == other.backward && m.n == other.n
}

func (m *Metrics) String() string {
	return fmt.Sprintf("%d wires, foward: %.2f KB %.2f KB/s, backward: %.2f KB %.2f KB/s",
		m.n, api.KB(m.forward), api.KBPS(m.forward, METER_PERIOD),
		api.KB(m.backward), api.KBPS(m.backward, METER_PERIOD))
}

func (live *Liveness) Measure() Metrics {
	c := make(chan *Wire)

	go func() {
		live.C <- true
		lastGood := -1
		for i, one := range live.W {
			c <- one

			if !one.closed {
				lastGood += 1
				live.W[i], live.W[lastGood] = live.W[lastGood], live.W[i]
			} else {
				live.W[i] = nil
			}
		}
		live.W = live.W[0: lastGood + 1]
		<-live.C

		close(c)
	}()

	metrics := Metrics{}
	for one := range c {
		_, f := one.fm.Consume()
		_, b := one.bm.Consume()

		if one.closed {
			one.fm.Close()
			one.bm.Close()
			one.fwb.Close()
			one.bwb.Close()
		}

		metrics.Aggregate(f, b)
	}

	return metrics
}
