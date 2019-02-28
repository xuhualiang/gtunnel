package main

import (
	"flag"
	"fmt"
	"net"
	"time"
	"os"
	"math/rand"
	"org/gtunnel/api"
)

const (
	IO_ALLOWANCE   = time.Millisecond * 100
	METER_PERIOD   = time.Second * 10
	BUF_SIZE       = 256 * 1024
)

func redirectLoop(wire *Wire, cfg *Cfg, rwb *api.RwBuf, from net.Conn, to net.Conn, m *meter) {
	for !wire.closed && !cfg.Timeout(wire.atime) {
		rd, wr := 0, 0
		touch := false

		// 1 - read
		b := rwb.ProducerBuffer()
		if !wire.closed && len(b) > 0 {
			var err error

			// don't block if rwb is consumable
			deadline := time.Time{}
			if !rwb.Consumable() {
				deadline = api.Deadline(IO_ALLOWANCE)
			}
			from.SetReadDeadline(deadline)

			rd, err = from.Read(b)
			if err != nil && !api.IsTimeoutError(err) {
				break
			} else if rd > 0 {
				rwb.Produce(rd)
				touch = true
			}
		}

		// 2 - write
		b = rwb.ConsumerBuffer()
		if !wire.closed && len(b) > 0 {
			var err error

			// don't block if rwb is producible
			deadline := time.Time{}
			if !rwb.Producible() {
				deadline = api.Deadline(IO_ALLOWANCE)
			}
			to.SetWriteDeadline(deadline)

			wr, err = to.Write(b)
			if err != nil && !api.IsTimeoutError(err)  {
				break
			} else if wr > 0 {
				rwb.Consume(wr)
				touch = true
			}
		}

		if touch {
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
	listenSock, err := api.Listen(cfg.Accept, cfg.Cert(), cfg.Key())
	if err != nil {
		fmt.Printf("failed to listen on %s, error - %s\n", cfg, err)
		return
	}
	fmt.Printf("listening on %s\n", cfg)

	lb := MkLoadBalancer(cfg.Connect)

	for {
		in, err := listenSock.Accept()
		if err != nil {
			fmt.Printf("failed to accept %s, error - %s\n", cfg.Accept, err)
			continue
		}
		go func() {
			ep := lb.Pick(api.DIAL_ALLOWANCE)
			if ep == nil {
				in.Close()
				fmt.Printf("failed to discover endpoint")
				return
			}

			out, err := api.Dial(ep, cfg.SkipVerify())
			if err != nil {
				in.Close()
				fmt.Printf("failed to wire %s, error - %s\n", cfg, err)
				return
			}

			wire := mkWire(cfg, in, out)
			go redirectLoop(wire, cfg, wire.fwb, wire.src, wire.dst, wire.fm)
			go redirectLoop(wire, cfg, wire.bwb, wire.dst, wire.src, wire.bm)

			live.Add(wire)
			fmt.Printf("U connection %s (%s)\n", wire, ep)
		}()
	}
}

func main() {
	cfg := Configuration{}
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [path-to-config-file]+\n", os.Args[0])
	}
	flag.Parse()
	rand.Seed(time.Now().Unix())

	err := cfg.Load(flag.Args())
	api.Assert(err == nil, fmt.Sprintf("failed to load configuration %s", err))

	live := NewLiveness()
	for i, _ := range cfg.Set {
		go listenLoop(cfg.Set[i], live)
	}

	time.Sleep(METER_PERIOD)
	for ticker := time.NewTicker(METER_PERIOD); ; <-ticker.C  {
		forward, backward, N := live.Measure()

		fmt.Printf("%d wires, foward: %.2f KB %.2f KB/s, backward: %.2f KB %.2f KB/s\n",
			N, api.KB(forward), api.KBPS(forward, METER_PERIOD),
				api.KB(backward), api.KBPS(backward, METER_PERIOD))
	}
}
