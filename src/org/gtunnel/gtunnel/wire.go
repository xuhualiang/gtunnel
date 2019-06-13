package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Wire struct {
	src net.Conn
	dst net.Conn

	atime  time.Time
	closed bool

	sndBytes uint64
	rcvBytes uint64
	mtx      sync.Mutex
}

func (wire *Wire) Close() {
	wire.mtx.Lock()
	defer wire.mtx.Unlock()

	wire.closed = true
	wire.src.Close()
	wire.dst.Close()
}

func mkWire(src net.Conn, dst net.Conn) *Wire {
	return &Wire{
		src:    src,
		dst:    dst,
		atime:  time.Now(),
		closed: false,
	}
}

func (wire *Wire) Meter(snd, rcv int) {
	wire.mtx.Lock()
	defer wire.mtx.Unlock()

	wire.sndBytes += uint64(snd)
	wire.rcvBytes += uint64(rcv)
	wire.atime = time.Now()
}

func (wire *Wire) GetAndReset() (snd, rcv uint64, closed bool) {
	wire.mtx.Lock()
	defer wire.mtx.Unlock()

	snd, rcv = wire.sndBytes, wire.rcvBytes
	wire.sndBytes, wire.rcvBytes = 0, 0
	closed = wire.closed
	return
}

type Measure struct {
	sndBytes uint64
	rcvBytes uint64
	nConn    int
}

func (m *Measure) Equals(other *Measure) bool {
	return m.sndBytes == other.sndBytes &&
		m.rcvBytes == other.rcvBytes && m.nConn == other.nConn
}

func (m *Measure) String(d time.Duration) string {
	return fmt.Sprintf("%d wires, send: %s, recv: %s",
		m.nConn, normalize(m.sndBytes, d), normalize(m.rcvBytes, d))
}

func normalize(b uint64, d time.Duration) string {
	if b >= 1024 * 1024 {
		f := float64(b) / 1024 / 1024
		return fmt.Sprintf("%.2f MB %.2f MB/s", f, f/d.Seconds())
	}

	f := float64(b) / 1024
	return fmt.Sprintf("%.2f KB %.2f KB/s", f, f/d.Seconds());
}

type Meter struct {
	ls   []*Wire
	mtx  sync.Mutex
}

func (m *Meter) Append(wire *Wire) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.ls = append(m.ls, wire)
}

func (m *Meter) GetAndReset() Measure {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	r := Measure{
		sndBytes: 0,
		rcvBytes: 0,
		nConn:    len(m.ls),
	}

	nextGoodWire := 0
	for i := 0; i < len(m.ls); i += 1 {
		snd, rcv, closed := m.ls[i].GetAndReset()
		r.sndBytes += snd
		r.rcvBytes += rcv

		if closed {
			m.ls[nextGoodWire], m.ls[i] = m.ls[i], m.ls[nextGoodWire]
			nextGoodWire += 1
		}
	}

	if nextGoodWire > 0 {
		m.ls = m.ls[nextGoodWire:]
	}
	return r
}

type T struct {
	b uint64
	d time.Duration
}