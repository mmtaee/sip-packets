package main

import (
	"sync"
)

type Request struct {
	strategy Strategy
}
type InviteStrategy struct{}

type RegisterStrategy struct{}

type Strategy interface {
	Send(conn Connection) (Connection, error)
}

func (r *Request) setStrategy(s Strategy) {
	r.strategy = s
}

func (r *Request) startStrategy(conn Connection) (Connection, error) {
	return r.strategy.Send(conn)
}

func (i *InviteStrategy) Send(conn Connection) (Connection, error) {
	return sendInvite(conn)
}

func (r *RegisterStrategy) Send(conn Connection) (Connection, error) {
	return sendRegister(conn)
}

func sendToStrategy(in <-chan Connection, wg *sync.WaitGroup) {
	request := Request{}
	if flags.strategy == "REGISTER" {
		request.setStrategy(&RegisterStrategy{})
	} else {
		request.setStrategy(&InviteStrategy{})
	}
	for o := range in {
		wg.Add(1)
		go func(o Connection) {
			o.Status = 21
			o, err = request.startStrategy(o)
			if err == nil && flags.output != "" {
				resultOutput(o.Username, o.Status.String())
			}
			defer wg.Done()
		}(o)
	}
	wg.Wait()
}
