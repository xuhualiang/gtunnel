package api

import (
	"net"
	"fmt"
	"strings"
	"math/rand"
	"regexp"
	"crypto/tls"
	"time"
)

const (
	DIAL_ALLOWANCE = time.Second
)

var (
	protExp = regexp.MustCompile("[ps]{2}")
)

func Deadline(d time.Duration) time.Time {
	return time.Now().Add(d)
}

func Due(deadline time.Time) bool {
	return time.Now().After(deadline)
}

func IsTimeoutError(err error) bool {
	other, ok := err.(net.Error)
	return ok && other.Timeout()
}

func KB(i uint64) float64 {
	return float64(i) / 1024.0
}

func KBPS(i uint64, d time.Duration) float64 {
	return KB(i) / d.Seconds()
}

type Endpoint struct {
	Addr       *net.TCPAddr
	SSL        bool
}

func (ep Endpoint) String() string {
	if ep.SSL {
		return fmt.Sprintf("%s", ep.Addr)
	} else {
		return fmt.Sprintf("%s", ep.Addr)
	}
}

type EndpointList struct {
	EP []*Endpoint
}

func (el *EndpointList) String() string {
	s := ""
	for i, ep := range el.EP {
		if i > 0 {
			s += ","
		}
		s += ep.String()
	}
	return s
}

func (el *EndpointList) Len() int {
	return len(el.EP)
}

func (el *EndpointList) Pick() *Endpoint {
	// random pick endpoint
	if len(el.EP) == 0 {
		return nil
	}
	return el.EP[rand.Int() % len(el.EP)]
}

func Assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

func MkEndpoint(s string, ssl bool) (*Endpoint, error) {
	addr, err := net.ResolveTCPAddr("tcp", s)
	if err != nil {
		return nil, err
	}
	return &Endpoint{Addr: addr, SSL: ssl}, nil
}

func MkEndpointList(s string, ssl bool) (*EndpointList, error) {
	r := &EndpointList{}

	for _, ss := range strings.Split(s, ",") {
		ep, err := MkEndpoint(ss, ssl)
		if err != nil {
			return nil, err
		}
		r.EP = append(r.EP, ep)
	}
	return r, nil
}

func Solve(prot string, src string, dst string) (*Endpoint, *EndpointList, error) {
	if strings.Trim(prot, " ") == "" {
		prot = "pp"
	}
	Assert(protExp.MatchString(prot), prot)

	srcEp, err := MkEndpoint(src, prot[0] == 's')
	if err != nil {
		return nil, nil, err
	}
	dstEp, err := MkEndpointList(dst, prot[1] == 's')
	if err != nil {
		return nil, nil, err
	}
	return srcEp, dstEp, nil
}

func Listen(ep *Endpoint, cert, key string) (net.Listener, error) {
	if ep.SSL {
		cert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		cfg := tls.Config{Certificates: []tls.Certificate{cert}}
		return tls.Listen("tcp", ep.String(), &cfg)
	} else {
		return net.Listen("tcp", ep.String())
	}
}

func Dial(ep *Endpoint, skipVerify bool) (net.Conn, error) {
	if ep.SSL {
		cfg := tls.Config{InsecureSkipVerify: skipVerify}
		dialer := net.Dialer{Timeout: DIAL_ALLOWANCE}

		return tls.DialWithDialer(&dialer, "tcp", ep.String(), &cfg)
	} else {
		return net.DialTimeout("tcp", ep.String(), DIAL_ALLOWANCE)
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "Dial timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func DialRetry(n int, ep *Endpoint, skipVerify bool) (net.Conn, error) {
	for i := 0; n == 0 || i < n; i++ {
		if conn, err := Dial(ep, skipVerify); err == nil {
			return conn, nil
		}
	}
	return nil, timeoutError{}
}