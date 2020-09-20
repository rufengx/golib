package xio

import (
	"errors"
	"net"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/xy1884/golib/xio/netpoll"
)

type loop struct {
	index   int           // loop lnidx in the xio loops list
	poll    *netpoll.Poll // epoll or kqueue
	packet  []byte        // read package buffer
	fdconns map[int]*conn // loop connection fd --> conn
	count   int32         // connection count
}

type conn struct {
	fd    int
	sa    syscall.Sockaddr // from sock addr
	lnidx int              // listener lnidx

	//reuse  bool
	//opened bool
	action Action

	out []byte // write buffer

	localAddr  net.Addr
	remoteAddr net.Addr

	loop *loop
}

func newLoop(index int) (*loop, error) {
	poll, err := netpoll.Create()
	if nil != err {
		return nil, err
	}
	return &loop{
		index:   index,
		poll:    poll,
		packet:  make([]byte, 65535),
		fdconns: make(map[int]*conn),
	}, nil
}

func (l *loop) run(srv *server) {
	defer func() {
		srv.signalExit()
		srv.wg.Done()
	}()

	// trigger
	if 0 == l.index && nil != srv.eventHandler.TickFunc {
		go l.trigger()
	}

	// loop start
	callback := func(fd int, ev uint32) error {
		// if no descriptors became ready during timeout.
		if 0 == fd {
			return l.note()
		}

		// select different func order by conn status.
		conn, ok := l.fdconns[fd]
		if !ok {
			return l.accept(srv, fd)
		} else {
			switch len(conn.out) > 0 {
			case false:
				return l.read(srv, conn)
			case true:
				return l.write(srv, conn)
				//case conn.action != None:
				//	return l.action()
				//default:
				//	return l.read()
			}
		}
	}

	l.poll.Wait(callback)
}

func (l *loop) note() error {
	return nil
}

func (l *loop) accept(srv *server, fd int) error {
	// step 1: load balance
	loopsLen := len(srv.loops)
	if loopsLen > 1 {
		switch srv.balance {
		case LeastConnection:
			n := atomic.LoadInt32(&l.count)
			for _, loop := range srv.loops {
				if loop.index != l.index && atomic.LoadInt32(&loop.count) < n {
					return nil // do not accept, wait next loop process.
				}
			}

		case RoundRobin:
			idx := int(atomic.LoadUintptr(&srv.accepted)) % loopsLen
			if idx != l.index {
				return nil // do not accept, wait next loop process.
			}
			atomic.AddUintptr(&srv.accepted, 1)
		case Random:
			// default strategy
		}
	}

	for i, listener := range srv.listeners {
		if listener.fd == fd {
			// step 2: udp conn
			if nil != listener.pconn {
				return l.acceptUdpConn(srv, i, fd)

			} else {
				// step 3: tcp or unix conn
				return l.acceptTcpConn(srv, i, fd)
			}
		}
	}
	return nil
}

func (l *loop) acceptTcpConn(srv *server, idx, fd int) error {
	// step 1: accept
	nfd, rsa, err := syscall.Accept(fd)
	if nil != err {
		if err == syscall.EAGAIN {
			return nil
		}
		return err
	}

	// step 2: set non-block
	err = syscall.SetNonblock(nfd, true)
	if nil != err {
		return err
	}

	conn := &conn{
		lnidx: idx,
		fd:    nfd,
		sa:    rsa,
		loop:  l,
	}

	l.fdconns[nfd] = conn
	l.poll.AddReadWrite(nfd)
	atomic.AddInt32(&l.count, 1)

	// step 3: open conn, read data.
	return l.opened(srv, conn)
}

func (l *loop) acceptUdpConn(srv *server, idx, fd int) error {
	// step 1: receive data
	n, fromsa, err := syscall.Recvfrom(fd, l.packet, 0)
	if nil != err || 0 == n {
		return nil
	}

	if nil == srv.eventHandler.DataFunc {
		return nil
	}

	// step 2: construct conn
	localAddr := srv.listeners[idx].lnAddr
	remoteAddr, err := socketAddrToAddr(&fromsa)
	if nil != err {
		return err
	}

	conn := &conn{
		fd:    fd,
		sa:    fromsa,
		lnidx: idx,

		localAddr:  localAddr,
		remoteAddr: remoteAddr,
		loop:       l,
	}

	// step 3: process data
	in := []byte{}
	out, action := srv.eventHandler.DataFunc(*conn, in)
	if len(out) > 0 && nil != srv.eventHandler.PreWriteFunc {
		srv.eventHandler.PreWriteFunc()
		syscall.Sendto(fd, out, 0, fromsa)
	}

	if action == Shutdown {
		return errors.New("closing")
	}
	return nil
}

func (l *loop) opened(srv *server, conn *conn) error {
	if nil == srv.eventHandler.OpenedFunc {
		return nil
	}

	out, action := srv.eventHandler.OpenedFunc(*conn)
	if len(out) == 0 && action == None {
		l.poll.ModifyRead(conn.fd)
		return nil
	}

	conn.out = append(conn.out, out...)
	// It's need to check listener is tcp conn before set keepalive.
	if _, ok := srv.listeners[conn.lnidx].ln.(*net.TCPListener); ok {
		SetKeepAlive(conn.fd, int(srv.eventHandler.KeepAlive/time.Second))
	}
	return nil
}

func (l *loop) read(srv *server, conn *conn) error {
	return nil
}

func (l *loop) write(srv *server, conn *conn) error {
	return nil
}

func (l *loop) action(conn *conn, action Action) error {
	return nil
}

func (l *loop) trigger() {
	// TODO:
	l.poll.Trigger()
}
