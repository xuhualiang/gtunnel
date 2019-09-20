package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"
)

const (
	MeterPeriod   = time.Second * 2
	BufSize       = 512 * 1024
	DialAllowance = time.Second * 2
)

func parseArgs() (ls EndpointPairList, cert, key string, verifyOpt VerifyOpt) {
	flag.CommandLine.Var(&ls, "pair", "Endpoint pair")
	flag.StringVar(&cert,"cert", "", "certificate file")
	flag.StringVar(&key, "key", "", "private key file")
	flag.CommandLine.Var(&verifyOpt, "verify", "ca=file:servername=hello.com")
	flag.Parse()
	return
}

func main() {
	fmt.Printf("go-version: %s %v\n", runtime.Version(), os.Args)
	ls, cert, key, verifyOpt := parseArgs()

	meter := &Meter{}
	for _, pair := range ls {
		go listenLoop(pair, cert, key, verifyOpt, meter)
	}

	time.Sleep(MeterPeriod)
	prevMetrics := Measure{0, 0, 0}
	for ticker := time.NewTicker(MeterPeriod); ; <-ticker.C  {
		measure := meter.GetAndReset()

		if !measure.Equals(&prevMetrics) {
			fmt.Println(measure.String(MeterPeriod))
			prevMetrics = measure
		}
	}
}

func redirectLoop(wire *Wire, from net.Conn, to net.Conn, master bool) {
	buf := make([]byte, BufSize)

	for !wire.closed {
		sz, err := from.Read(buf)
		if err != nil {
			break
		}

		if err := sendFull(to, buf[:sz]); err != nil {
			break
		}

		// 3. meter
		if master {
			wire.Meter(sz, 0)
		} else {
			wire.Meter(0, sz)
		}
	}

	if master {
		wire.Close()
	}
}

func listenLoop(pair EndpointPair, cert, key string, verifyOpt VerifyOpt, m *Meter) {
	listenSock, err := listenOn(pair.Src, cert, key)
	if err != nil {
		fmt.Printf("failed to listen on %s, error=%s\n", pair.Src, err)
		os.Exit(-1)
	}
	fmt.Printf("listening on %s\n", pair.Src)

	for {
		fromConn, err := listenSock.Accept()
		if err != nil {
			fmt.Printf("accept error=%s\n", err)
			continue
		}

		go func() {
			toConn, err := dialTo(pair.Dst, verifyOpt)
			if err != nil {
				fromConn.Close()
				fmt.Printf("failed to connect %s, error=%s\n", pair.Dst, err)
				return
			}

			wire := mkWire(fromConn, toConn)
			go redirectLoop(wire, wire.src, wire.dst, true)
			go redirectLoop(wire, wire.dst, wire.src, false)
			m.Append(wire)
		} ()
	}
}

func listenOn(ep Endpoint, cert, key string) (net.Listener, error) {
	if ep.IsSecure() {
		cert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		cfg := tls.Config{Certificates: []tls.Certificate{cert}}
		return tls.Listen("tcp", ep.Addr(), &cfg)
	} else {
		return net.Listen("tcp", ep.Addr())
	}
}

func dialTo(ep Endpoint, verifyOpt VerifyOpt) (net.Conn, error) {
	if ep.IsSecure() {
		cfg := tls.Config{
			InsecureSkipVerify: !verifyOpt.DoVerify,
			RootCAs: verifyOpt.RootCA,
			ServerName: verifyOpt.ServerName,
		}
		dialer := net.Dialer{Timeout: DialAllowance}

		return tls.DialWithDialer(&dialer, "tcp", ep.Addr(), &cfg)
	} else {
		return net.DialTimeout("tcp", ep.Addr(), DialAllowance)
	}
}

func sendFull(c net.Conn, buf []byte) error {
	for snd := 0; snd < len(buf); {
		if wr, err := c.Write(buf[snd:]); err != nil {
			return err
		} else {
			snd += wr
		}
	}
	return nil
}