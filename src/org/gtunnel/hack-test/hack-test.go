package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	DataSize       = 1024 * 1024 * 8
	ThroughputLoop = 64
	LatencyLoop    = 4096
)

func parseArgs() string {
	ep := flag.String("ep", "127.0.0.1:10002", "endpoint")
	flag.Parse()
	return *ep
}

func main() {
	ep := parseArgs()

	conn, err := net.Dial("tcp", ep)
	if err != nil {
		fmt.Printf("failed to dial to %s, err=%s\n", ep, err)
		os.Exit(-1)
	}

	throughput(conn)
	latency(conn)
}

func throughput(conn net.Conn) {
	data := make([]byte, DataSize)
	rand.Read(data)
	t0 := time.Now()

	// send routing
	go func() {
		for i := 0; i < ThroughputLoop; i += 1 {
			if err := sendFull(conn, data); err != nil {
				fmt.Printf("failed to send data err=%s", err)
				os.Exit(-1)
			}
		}
	}()
	// recv loop
	buf := make([]byte, DataSize)
	for i := 0; i < ThroughputLoop; i += 1 {
		if err := recvFull(conn, buf); err != nil {
			fmt.Printf("failed to recv data err=%s\n", err)
			os.Exit(-1)
		}
	}
	duration := time.Now().Sub(t0)
	fmt.Printf("%s\n", normalize(DataSize*ThroughputLoop, duration))
	if bytes.Compare(data, buf) != 0 {
		fmt.Printf("but wait, data corrupted\n")
		os.Exit(-1)
	}
}

func latency(conn net.Conn) {
	data := make([]byte, 32)
	t0 := time.Now()

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

	d := time.Now().Sub(t0)

	fmt.Printf("%.2f seconds, %d messages, %.2f milli sec/message\n",
		d.Seconds(), LatencyLoop, float64(d/time.Millisecond)/LatencyLoop)
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

func normalize(b uint64, d time.Duration) string {
	if b >= 1024 * 1024 {
		f := float64(b) / 1024 / 1024
		return fmt.Sprintf("%.2f MB %.2f MB/s", f, f/d.Seconds())
	}

	f := float64(b) / 1024
	return fmt.Sprintf("%.2f KB %.2f KB/s", f, f/d.Seconds());
}