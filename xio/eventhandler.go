package xio

import (
	"io"
	"time"
)

type Action int

const (
	None Action = iota
	Read
	Write
	Close
	Shutdown
)

type LoadBalance int

const (
	Random LoadBalance = iota
	RoundRobin
	LeastConnection
)

type Server struct {
}

type EventHandler struct {
	NumLoops    int
	LoadBalance LoadBalance
	KeepAlive   time.Duration

	Serving func(server Server) (action Action)

	OpenedFunc func(c conn) (out []byte, action Action)

	ClosedFunc func(c conn, err error) (action Action)

	DetachedFunc func(c conn, rwc io.ReadWriteCloser) (action Action)

	PreWriteFunc func()

	DataFunc func(c conn, in []byte) (out []byte, action Action)

	TickFunc func() (delay time.Duration, action Action)
}
