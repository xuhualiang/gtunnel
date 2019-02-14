package main

import "fmt"

type meter struct {
	rd uint64
	wr uint64
	C chan bool
}

func NewMeter() *meter {
	m := &meter{}
	m.C <- true
	return m
}

func (m *meter) Produce(rd, wr uint64)  {
	<- m.C
		m.rd, m.wr = m.rd + rd, m.wr + wr
	m.C <- true
}

func (m *meter) Consume() (rd, wr uint64) {
	<- m.C
		rd, wr = m.rd, m.wr
		m.rd, m.wr = 0, 0
	m.C <- true
	return
}

func (m meter) String() string {
	return fmt.Sprintf("rd=%d wr=%d", m.rd, m.wr)
}