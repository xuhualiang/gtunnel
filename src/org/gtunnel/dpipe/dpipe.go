package main

import (
	"flag"
	"regexp"
	"org/gtunnel/api"
	"os"
	"net"
	"fmt"
	"time"
)

var epExpr = regexp.MustCompile(`[0-9a-zA-Z\.]:[0-9]`)

func parseInput() (string, string, uint64, uint64) {
	in := flag.String("in", "", "input source")
	out := flag.String("out", "", "output dest")
	blockSz := flag.Uint64("block", 8192, "block size")
	contentSz := flag.Uint64("size", 1024 * 1024, "content size")

	flag.Parse()
	return *in, *out, *blockSz, *contentSz
}

func sock2file(inConn net.Conn, outFile *os.File, blockSz, contentSz uint64) time.Duration {
	rwb := api.MkRWBuf(blockSz)
	start := time.Now()

	for i := uint64(0); i < contentSz; {
		if rwb.Producible() {
			n, err := inConn.Read(rwb.ProducerBuffer())
			api.Assert(err == nil,
				fmt.Sprintf("failed to read socket %s", err))
			rwb.Produce(n)
		}

		if rwb.Consumable() {
			n, err := outFile.Write(rwb.ConsumerBuffer())
			api.Assert(err == nil,
				fmt.Sprintf("failed to write file %s", err))
			rwb.Consume(n)
			i += uint64(n)
		}
	}

	return time.Now().Sub(start)
}

func file2sock(inFile *os.File, outConn net.Conn, blockSz, contentSz uint64) time.Duration {
	rwb := api.MkRWBuf(blockSz)
	start := time.Now()

	for i := uint64(0); i < contentSz; {
		if rwb.Producible() {
			n, err := inFile.Read(rwb.ProducerBuffer())
			api.Assert(err == nil,
				fmt.Sprintf("failed to read file %s", err))
			rwb.Produce(n)
		}

		if rwb.Consumable() {
			fmt.Printf("%s\n", rwb.ConsumerBuffer())
			n, err := outConn.Write(rwb.ConsumerBuffer())
			api.Assert(err == nil,
				fmt.Sprintf("failed to write socket %s", err))
			rwb.Consume(n)
			i += uint64(n)
		}
	}
	return time.Now().Sub(start)
}

func main() {
	in, out, blockSz, contentSz := parseInput()
	d := time.Duration(0)

	if epExpr.MatchString(in) {
		inEp, err := api.MkEndpoint(in, false /*=!ssl*/)
		api.Assert(err == nil,
			fmt.Sprintf("failed to solve endpoint %s", in))

		listenSock, err := api.Listen(inEp, "", "")
		api.Assert(err == nil,
			fmt.Sprintf("failed to accept connection %s", err))

		inConn, err := listenSock.Accept()
		api.Assert(err == nil,
			fmt.Sprintf("failed to accept connection %s %s", inEp, err))

		if epExpr.MatchString(out) {
			api.Assert(false, "unreachable`")
		} else {
			outFile, err := os.OpenFile(out,
				os.O_APPEND | os.O_CREATE | os.O_RDWR, 0755)
			api.Assert(err == nil,
				fmt.Sprintf("failed to open %s %s", out, err))

			defer outFile.Close()
			d = sock2file(inConn, outFile, blockSz, contentSz)
		}
	} else {
		inFile, err := os.Open(in)
		api.Assert(err == nil,
			fmt.Sprintf("failed to open %s %s", in, err))

		defer inFile.Close()
		if epExpr.MatchString(out) {
			outEp, err := api.MkEndpoint(out, false /*=!ssl*/)
			api.Assert(err == nil,
				fmt.Sprintf("failed to solve endpoint %s", out))

			outConn, err := api.DialRetry(0, outEp, true /*=skipVerify*/)
			api.Assert(err == nil,
				fmt.Sprintf("failed to connect %s %s", outEp, err))

			d = file2sock(inFile, outConn, blockSz, contentSz)
		} else {
			api.Assert(false, "unreachable`")
		}
	}

	fmt.Printf("%.2f KB/s\n", api.KBPS(contentSz, d))
}
