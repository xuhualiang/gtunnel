package main

import (
	"testing"
	"math/rand"
	"bytes"
	"time"
)

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func randomize(b []byte)  {
	for i := range b {
		b[i] = byte(rand.Int())
	}
}

func Test0(t *testing.T)  {
	const CAP = 64
	const SIZE = 1024 * 64

	rand.Seed(time.Now().Unix())
	src := make([]byte, SIZE)
	dst := make([]byte, SIZE)
	rwb := NewRWBuf(CAP)

	randomize(src)

	for i, j := 0, 0; i < SIZE || j < SIZE ; {
		b := rwb.ProducerBuffer()
		n := minInt(minInt(len(b), SIZE - i), rand.Intn(CAP))
		if n > 0 {
			copy(b[:n], src[i:])
			rwb.Produce(n)
		}

		b = rwb.ConsumerBuffer()
		m := minInt(minInt(len(b), SIZE - j), rand.Intn(CAP))
		if m > 0 {
			copy(dst[j:], b[:m])
			rwb.Consume(m)
		}

		i += n
		j += m
	}

	assert(bytes.Equal(src, dst), "")
}