package xio

import (
	"runtime"
	"sync"
)

type Conn interface {
}

func Serve(eventHandler EventHandler, addrs ...string) error {
	listeners := make([]*listener, len(addrs))
	for _, addr := range addrs {
		listener, err := newListener(addr)
		if nil != err {
			return err
		}
		listeners = append(listeners, listener)
	}
	return run(eventHandler, listeners)
}

func run(eventHandler EventHandler, listeners []*listener) error {
	numLoops := eventHandler.NumLoops
	if 0 == numLoops {
		// this section need to consider run in virtual machine.
		// use "go.uber.org/automaxprocs/maxprocs" package, init in xio start.
		numLoops = runtime.GOMAXPROCS(0)
	}

	srv := &server{
		wg:   sync.WaitGroup{},
		cond: sync.NewCond(&sync.Mutex{}),

		eventHandler: eventHandler,
		balance:      eventHandler.LoadBalance,
		listeners:    listeners,
	}

	for i := 0; i < numLoops; i++ {
		loop, err := newLoop(i)
		if nil != err {
			return err
		}
		for _, listener := range listeners {
			loop.poll.AddRead(listener.fd)
		}
		srv.loops = append(srv.loops, loop)
	}

	// start loops in background.
	srv.wg.Add(numLoops)
	for _, loop := range srv.loops {
		go loop.run(srv)
	}
	return nil
}
