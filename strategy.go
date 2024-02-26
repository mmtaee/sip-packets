package main

import (
	"fmt"
	"sync"
)

type Request struct {
	strategy Strategy
}
type InviteStrategy struct{}

type RegisterStrategy struct{}

type Strategy interface {
	Send(conn Connection)
}

func (r *Request) setStrategy(s Strategy) {
	r.strategy = s
}

func (r *Request) startStrategy(conn Connection) {
	r.strategy.Send(conn)
}

func (i *InviteStrategy) Send(conn Connection) {
	fmt.Println("invite Strategy")
}

func (r *RegisterStrategy) Send(conn Connection) {
	fmt.Println("register Strategy")
	sendRegister(conn)
}

func __sendToStrategy(in <-chan Connection, wg *sync.WaitGroup) {
	request := Request{}
	if flags.strategy == "REGISTER" {
		request.setStrategy(&RegisterStrategy{})
	} else {
		request.setStrategy(&InviteStrategy{})
	}
	var mutex sync.Mutex
	for o := range in {
		wg.Add(1) // Add to WaitGroup for each iteration
		go func(o Connection) {
			mutex.Lock()
			err := sendRegister(o)
			if err != nil {
				return
			}
			wg.Done()
			//request.startStrategy(o, wg)
			mutex.Unlock()
		}(o)
	}
	wg.Wait()
}
