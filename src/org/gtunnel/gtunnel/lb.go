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
type EasyLoadBalancer struct {
	ep *api.Endpoint
}

func mkEasyLoadBalancer(ep *api.Endpoint) LoadBalancer {
	return &EasyLoadBalancer{
		ep: ep,
	}
}

func (lb *EasyLoadBalancer) Pick(_ time.Duration) *api.Endpoint {
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

func mkRandomLoadBalancer(epl *api.EndpointList) LoadBalancer {
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