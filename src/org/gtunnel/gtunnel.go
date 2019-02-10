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
	IO_TIMEOUT   = time.Second * 1
)

type Wire struct {
	buf    [4096]byte
	data   []byte
	dataTo net.Conn

	src net.Conn
	dst net.Conn

	atime  time.Time
	closed chan bool
	ready  chan bool
}

func deadline(duration time.Duration) time.Time {
	return time.Now().Add(duration)
}

func (wire *Wire) Close() {
	wire.closed <- true
	wire.src.Close()
	wire.dst.Close()
}

func mkWire(src net.Conn, dst net.Conn) *Wire {
	// timeout for io
	src.SetDeadline(deadline(IO_TIMEOUT))
	dst.SetDeadline(deadline(IO_TIMEOUT))

	wire := &Wire{
		src:   src,
		dst:   dst,
		atime: time.Now(),
	}
	wire.ready <- true

	return wire
}

func (wire *Wire) String() string {
	return fmt.Sprintf("%s/%s", wire.src, wire.dst)
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

func readLoop(wire *Wire, from net.Conn, to net.Conn, forwardQueue chan *Wire) {
	for {
		select {
		case <-wire.ready:
			n, err := from.Read(wire.buf[:])
			if err != nil {
				fmt.Printf("error on wire %s %s\n", wire, err)
				wire.Close()
				break
			}
			if n > 0 {
				wire.data = wire.buf[:n]
				wire.dataTo = to
				wire.atime = time.Now()
				forwardQueue <- wire
			} else {

			}

		case <-wire.closed:
			wire.closed <- true
			break
		}
	}
}

func writeLoop(forwardQueue chan *Wire) {
	for {
		wire := <-forwardQueue

		n, err := wire.dataTo.Write(wire.data)
		if err != nil {
			wire.Close()
			continue
		}

		if n < len(wire.data) {
			wire.data = wire.data[:n]
			forwardQueue <- wire
		} else {
			wire.data = wire.buf[0:0]
			wire.dataTo = nil
			wire.ready <- true
		}
	}
}

func listenLoop(cfg Cfg, forwardQueue chan *Wire) {
	listenSock, err := listen(&cfg)
	if err != nil {
		fmt.Printf("failed to listen on %s, error - %s\n", cfg.String(), err)
		return
	}
	fmt.Printf("listening on %s\n", cfg.String())

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
				fmt.Printf("failed to wire %s, error - %s\n", cfg.String(), err)
				return
			}

			wire := mkWire(in, out)
			go readLoop(wire, wire.src, wire.dst, forwardQueue)
			go readLoop(wire, wire.dst, wire.src, forwardQueue)
		}()

		fmt.Printf("listening on %s\n", cfg.Accept.String())
	}
}

func main() {
	cfg := Configuration{}
	flag.Parse()

	err := cfg.Load(flag.Args())
	assert(err == nil, fmt.Sprintf("failed to load configuration %s", err))

	forwardQueue := make(chan *Wire, 4096)

	for _, c := range cfg.Set {
		go listenLoop(c, forwardQueue)
	}

	writeLoop(forwardQueue)
}
