package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	LatencyLoop    = 1024 * 16
)

func parseArgs() (string, int) {
	ep := flag.String("ep", "127.0.0.1:10002", "endpoint")
	dup := flag.Int("dup", 1, "number of concurrently running duplicated workload")

	flag.Parse()
	return *ep, *dup
}

func main() {
	ep, dup := parseArgs()
	t0 := time.Now()
	done := make(chan struct{}, dup)

	for i := 0; i < dup; i += 1 {
		go latency(ep, done)
	}

	// wait to be done
	for i := 0; i < dup; i += 1 {
		<-done
	}

	d := time.Now().Sub(t0)
	fmt.Printf("latency: %.2f seconds, %d messages, %.2f milli sec/message\n",
		d.Seconds(), LatencyLoop, float64(d/time.Millisecond)/LatencyLoop)
}

func latency(ep string, done chan struct{}) {
	conn, err := net.Dial("tcp", ep)
	if err != nil {
		fmt.Printf("failed to dial to %s, err=%s\n", ep, err)
		os.Exit(-1)
	}
	data := make([]byte, 32)

	for i := 0; i < LatencyLoop; i += 1 {
		if err := sendFull(conn, data); err != nil {
			fmt.Printf("failed to send err=%s\n", err)
			os.Exit(-1)
		}

		if err := recvFull(conn, data); err != nil {
			fmt.Printf("failed to recv err=%s\n", err)
			os.Exit(-1)
		}
	}

	done <- struct{}{}
}

func sendFull(conn net.Conn, buf []byte) error {
	for i := 0; i < len(buf); {
		wr, err := conn.Write(buf[i:])
		if err != nil {
			return nil
		}
		i += wr
	}
	return nil
}

func recvFull(conn net.Conn, buf []byte) error {
	for i := 0; i < len(buf); {
		rd, err := conn.Read(buf[i:])
		if err != nil {
			return err
		}
		i += rd
	}
	return nil
}