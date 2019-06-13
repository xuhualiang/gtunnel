package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

const (
	BufSize = 512 * 1024
)

func parseArgs() string {
	ep := flag.String("ep", "127.0.0.1:10000", "endpoint")
	flag.Parse()
	return *ep
}

func main() {
	ep := parseArgs()

	listener, err := net.Listen("tcp", ep)
	if err != nil {
		fmt.Printf("failed to listen %s, err=%s\n", ep, err)
		os.Exit(-1)
	}
	fmt.Printf("listening on %s\n", ep)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("acceptor err=%s\n", err)
			continue
		}
		fmt.Printf("+ connection\n")

		go func() {
			buf := make([]byte, BufSize)

			for {
				rd, err := conn.Read(buf)
				if err != nil {
					fmt.Printf("- read err=%s\n", err)
					break
				}

				if err := sendFull(conn, buf[:rd]); err != nil {
					fmt.Printf("- write err=%s\n", err)
					break
				}
			}

			conn.Close()
		}()
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
