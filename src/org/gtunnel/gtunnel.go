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
	IO_TIMEOUT   = time.Second
	BUF_SIZE     = 256 * 1024
)

func deadline(d time.Duration) time.Time {
	return time.Now().Add(d)
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
		// 1 - read
		b := rwb.ProducerBuffer()
		if !wire.closed && len(b) > 0 {
			from.SetReadDeadline(deadline(IO_TIMEOUT))
			n, err := from.Read(b)
			if err != nil {
				fmt.Printf("read error - %s \n", err)
				break
			}
			rwb.Produce(n)
			m.rd += uint64(n)
		}

		// 2 - write
		b = rwb.ConsumerBuffer()
		if !wire.closed && len(b) > 0 {
			to.SetWriteDeadline(deadline(IO_TIMEOUT))
			n, err := to.Write(b)
			if err != nil {
				fmt.Printf("write error - %s \n", err)
				break
			}
			rwb.Consume(n)
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
			fmt.Printf("failed to accept %s, error - %s\n", cfg.Accept, err)
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
			go readLoop(wire, wire.fwb, wire.src, wire.dst, &wire.fm)
			go readLoop(wire, wire.bwb, wire.dst, wire.src, &wire.bm)
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

	ticker := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ticker.C:
		}
	}
}
