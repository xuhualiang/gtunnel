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
)

func parseArgs() (string, int, int) {
	ep := flag.String("ep", "127.0.0.1:10002", "endpoint")
	loop := flag.Int("loop", 16, "loops of 8MB data")
	dup := flag.Int("dup", 1, "number of concurrently running duplicated workload")

	flag.Parse()
	return *ep, *loop, *dup
}

func main() {
	ep, loop, dup := parseArgs()
	t0 := time.Now()
	done := make(chan struct{}, loop)

	for i := 0; i < dup; i += 1 {
		go throughput(loop, ep, done)
	}

	// wait for all done
	for i := 0; i < dup; i += 1 {
		<-done
	}

	duration := time.Now().Sub(t0)
	fmt.Printf("thoughput: %s\n", normalize(DataSize*uint64(loop*dup), duration))
}

func throughput(loop int, ep string, done chan struct{}) {
	data := make([]byte, DataSize)
	rand.Read(data)

	conn, err := net.Dial("tcp", ep)
	if err != nil {
		fmt.Printf("failed to dial to %s, err=%s\n", ep, err)
		os.Exit(-1)
	}

	// send routing
	go func() {
		for i := 0; i < loop; i += 1 {
			if err := sendFull(conn, data); err != nil {
				fmt.Printf("failed to send data err=%s", err)
				os.Exit(-1)
			}
		}
	}()
	// recv loop
	buf := make([]byte, DataSize)
	for i := 0; i < loop; i += 1 {
		if err := recvFull(conn, buf); err != nil {
			fmt.Printf("failed to recv data err=%s\n", err)
			os.Exit(-1)
		}
	}

	if bytes.Compare(data, buf) != 0 {
		fmt.Printf("data corrupted\n")
		os.Exit(-1)
	}

	done <- struct {}{}
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