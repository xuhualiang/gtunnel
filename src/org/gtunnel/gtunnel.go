package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"time"
)

const (
	DIAL_TIMEOUT = time.Second * 2
	BUF_SIZE     = 256 * 1024
)

type meter struct {
	rd uint64
	wr uint64
}

func (m meter) String() string {
	return fmt.Sprintf("rd=%d wr=%d", m.rd, m.wr)
}

type Wire struct {
	cfg *Cfg
	src net.Conn
	dst net.Conn

	fwb *rwbuf
	bwb *rwbuf

	atime  time.Time
	closed bool

	fmeter meter
	bmeter meter
}

func deadline(duration time.Duration) time.Time {
	return time.Now().Add(duration)
}

func (wire *Wire) Close() {
	wire.closed = true
	wire.src.Close()
	wire.dst.Close()
}

func mkWire(cfg *Cfg, src net.Conn, dst net.Conn) *Wire {
	// don't block on write
	src.SetWriteDeadline(time.Time{})
	dst.SetWriteDeadline(time.Time{})

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
		wire.cfg, wire.fmeter, wire.bmeter)
}

func listen(cfg *Cfg) (net.Listener, error) {
	ep := &cfg.Accept

	if ep.SSL {
		cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
		if err != nil {
			return nil, err
		}
		cfg := tls.Config{Certificates: []tls.Certificate{cert}}
		return tls.Listen("tcp", ep.String(), &cfg)
	} else {
		return net.Listen("tcp", ep.String())
	}
}

func dial(ep *Endpoint) (net.Conn, error) {
	if ep.SSL {
		cfg := tls.Config{}
		dialer := net.Dialer{Timeout: DIAL_TIMEOUT}

		return tls.DialWithDialer(&dialer, "tcp", ep.String(), &cfg)
	} else {
		return net.DialTimeout("tcp", ep.String(), DIAL_TIMEOUT)
	}
}

func readLoop(wire *Wire, rwb *rwbuf, from net.Conn, to net.Conn, m *meter) {
	for !wire.closed {
		b := rwb.Writter()
		if !wire.closed && len(b) > 0 {
			n, err := from.Read(b)
			if err != nil {
				break
			}
			rwb.Write(n)
			m.rd += uint64(n)
		}

		b = rwb.Reader()
		if !wire.closed && len(b) > 0 {
			n, err := to.Write(b)
			if err != nil {
				break
			}
			rwb.Read(n)
			m.wr += uint64(n)
		}
	}

	if !wire.closed {
		fmt.Printf("close connection %s\n", wire)
		wire.Close()
	}
}

func listenLoop(cfg *Cfg) {
	listenSock, err := listen(cfg)
	if err != nil {
		fmt.Printf("failed to listen on %s, error - %s\n", cfg, err)
		return
	}
	fmt.Printf("listening on %s\n", cfg)

	for {
		in, err := listenSock.Accept()
		if err != nil {
			fmt.Printf("failed to accept %s, error - %s\n", cfg.Accept.String(), err)
			continue
		}
		go func() {
			out, err := dial(&cfg.Connect)
			if err != nil {
				in.Close()
				fmt.Printf("failed to wire %s, error - %s\n", cfg, err)
				return
			}

			wire := mkWire(cfg, in, out)
			go readLoop(wire, wire.fwb, wire.src, wire.dst, &wire.fmeter)
			go readLoop(wire, wire.bwb, wire.dst, wire.src, &wire.bmeter)
		}()
	}
}

func main() {
	cfg := Configuration{}
	flag.Parse()

	err := cfg.Load(flag.Args())
	assert(err == nil, fmt.Sprintf("failed to load configuration %s", err))

	for i, _ := range cfg.Set {
		go listenLoop(&cfg.Set[i])
	}

	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
		}
	}
}
