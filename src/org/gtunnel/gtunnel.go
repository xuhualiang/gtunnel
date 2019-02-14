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
	METER_TIME   = time.Second * 10
	BUF_SIZE     = 256 * 1024
)

func deadline(d time.Duration) time.Time {
	return time.Now().Add(d)
}

func isTimeout(err error) bool {
	other, ok := err.(net.Error)
	return ok && other.Timeout()
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

func readLoop(wire *Wire, cfg *Cfg, rwb *rwbuf, from net.Conn, to net.Conn, m *meter) {
	for !wire.closed && !cfg.Timeout(wire.atime) {
		rd, wr := 0, 0

		// 1 - read
		b := rwb.ProducerBuffer()
		if !wire.closed && len(b) > 0 {
			var err error = nil

			from.SetReadDeadline(deadline(IO_TIMEOUT))
			rd, err = from.Read(b)
			if err != nil && !isTimeout(err) {
				break
			}
			rwb.Produce(rd)
			wire.Touch()
		}

		// 2 - write
		b = rwb.ConsumerBuffer()
		if !wire.closed && len(b) > 0 {
			var err error = nil

			to.SetWriteDeadline(deadline(IO_TIMEOUT))
			wr, err = to.Write(b)
			if err != nil && !isTimeout(err)  {
				break
			}
			rwb.Consume(wr)
			wire.Touch()
		}

		// 3. meter
		m.Produce(rd, wr)
	}

	if !wire.closed {
		fmt.Printf("D connection %s\n", wire)
		wire.Close()
	}
}

func listenLoop(cfg *Cfg, live *Liveness) {
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
			go readLoop(wire, cfg, wire.fwb, wire.src, wire.dst, wire.fm)
			go readLoop(wire, cfg, wire.bwb, wire.dst, wire.src, wire.bm)

			live.Add(wire)
			fmt.Printf("U connection %s\n", wire)
		}()
	}
}

func main() {
	cfg := Configuration{}
	flag.Parse()

	err := cfg.Load(flag.Args())
	assert(err == nil, fmt.Sprintf("failed to load configuration %s", err))

	live := NewLiveness()
	for i, _ := range cfg.Set {
		go listenLoop(&cfg.Set[i], live)
	}

	time.Sleep(METER_TIME)
	for ticker := time.NewTicker(METER_TIME); ; <-ticker.C  {
		forward, backward := live.Measure()

		fmt.Printf("foward: %d bytes %f KB/s, backward: %d bytes %f KB/s\n",
			forward, float64(forward) / 1024.0 / METER_TIME.Seconds(),
				backward, float64(backward) / 1024.0 / METER_TIME.Seconds())
	}
}
