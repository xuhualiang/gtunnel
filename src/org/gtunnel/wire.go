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

	fm meter
	bm meter
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
	}

	return wire
}

func (wire Wire) String() string {
	return fmt.Sprintf("%s f: %s b: %s",
		wire.cfg, wire.fm, wire.bm)
}
