package main

import (
	"org/gtunnel/api"
	"time"
)

const LB_CHECK_PERIOD = time.Second * 2

type LoadBalancer interface {
	Pick(d time.Duration) *api.Endpoint
}

// not a load balancer
type NotALoadBalancer struct {
	ep *api.Endpoint
}

func mkNotALoadBalancer(ep *api.Endpoint) LoadBalancer {
	return &NotALoadBalancer{
		ep: ep,
	}
}

func (lb *NotALoadBalancer) Pick(_ time.Duration) *api.Endpoint {
	return lb.ep
}

// randomized load balancer
type RoundRobinBalancer struct {
	epl []*api.Endpoint
	c chan *api.Endpoint
}

func (lb *RoundRobinBalancer) monitor(){
	for ticker := time.NewTicker(LB_CHECK_PERIOD); ; <-ticker.C {
		// clear chan
		for len(lb.c) > 0 {
			<-lb.c
		}

		for _, ep := range lb.epl {
			conn, err := api.Dial(ep, true /*=!ssl*/)
			if err == nil {
				conn.Close()
				lb.c <- ep
			}
		}
	}

	api.Assert(false, "unreachable")
}

func mkRoundRobinLoadBalancer(epl *api.EndpointList) LoadBalancer {
	lb := &RoundRobinBalancer{
		epl: epl.EP,
		c:   make(chan *api.Endpoint, len(epl.EP)),
	}

	// routine checks health
	go lb.monitor()
	return lb
}

func (lb *RoundRobinBalancer) Pick(d time.Duration) *api.Endpoint {
	select {
	case <-time.After(d):
		return nil

	case ep := <- lb.c:
		lb.c <- ep
		return ep
	}
}

func MkLoadBalancer(epl *api.EndpointList) LoadBalancer {
	if len(epl.EP) == 1 {
		return mkNotALoadBalancer(epl.EP[0])
	} else {
		return mkRoundRobinLoadBalancer(epl)
	}
}