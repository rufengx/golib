package xio

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "go.uber.org/automaxprocs"
)

type server struct {
	wg   sync.WaitGroup
	cond *sync.Cond

	eventHandler EventHandler // process per conn by custom.
	balance      LoadBalance
	accepted     uintptr
	listeners    []*listener
	loops        []*loop // help load balance
}

// Listen quit or interrupt signal.
func (s *server) listenSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		<-c

		// TODO:
		// 1. stop accept connection
		// 2. flush buffer
		// 3. close connection
		// 4. dump crash data
	}()
}

// Exit xio, clear xio data.
func (s *server) Exit() {
	s.cond.L.Lock()
	s.cond.Wait()
	s.cond.L.Unlock()
}

// Notify xio exit.
func (s *server) signalExit() {
	s.cond.L.Lock()
	s.cond.Signal()
	s.cond.L.Unlock()
}
