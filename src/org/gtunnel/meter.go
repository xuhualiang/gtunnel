package main

import "fmt"

type meter struct {
	rd uint64
	wr uint64
}

func (m meter) String() string {
	return fmt.Sprintf("rd=%d wr=%d", m.rd, m.wr)
}