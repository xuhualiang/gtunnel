package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	endpointExpr = regexp.MustCompile(
		"(\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})?:(\\d+)(:[ps])?")
)

type Endpoint string

/* flag.Value interface */
func (ep *Endpoint) Set(s string) error {
	if !endpointExpr.MatchString(s) {
		return fmt.Errorf("bad endpoint %s", s)
	}
	*ep = Endpoint(s)
	return nil
}

func (ep Endpoint) String() string {
	return string(ep)
}

func (ep *Endpoint) IsSecure() bool {
	ss := strings.Split(string(*ep), ":")
	return len(ss) == 3 && ss[2] == "s"
}

func (ep *Endpoint) Addr() string {
	ss := strings.Split(string(*ep), ":")
	return ss[0] + ":" + ss[1]
}

type EndpointPair struct {
	Src Endpoint
	Dst Endpoint
}

/* flag.Value interface */
func (pair *EndpointPair) Set(s string) error {
	ss := strings.Split(s, "-")
	if len(ss) != 2 {
		return fmt.Errorf("bad endpoint pair %s", s)
	}

	if err := pair.Src.Set(ss[0]); err != nil {
		return err
	}
	if err := pair.Dst.Set(ss[1]); err != nil {
		return err
	}
	return nil
}

type EndpointPairList []EndpointPair

/* flag.Value interface */
func (ls *EndpointPairList) Set(s string) error {
	var pair EndpointPair

	if !endpointExpr.MatchString(s) {
		return fmt.Errorf("bad endpoint %s", s)
	}

	if err := pair.Set(s); err != nil {
		return err
	}

	*ls = append(*ls, pair)
	return nil
}

func (ls EndpointPairList) String() string {
	return fmt.Sprintf("%s", []EndpointPair(ls))
}