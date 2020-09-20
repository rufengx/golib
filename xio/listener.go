package xio

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

type listener struct {
	ln     net.Listener // tcp or unix
	lnAddr net.Addr
	pconn  net.PacketConn // udp

	file *os.File // per connection is a file
	fd   int

	network string
	addr    string
}

func newListener(addr string) (*listener, error) {
	var ln net.Listener      // tcp or unix
	var pconn net.PacketConn // udp
	var lnAddr net.Addr
	var file *os.File
	var err error

	// step 1: listen
	network, address := parseAddr(addr)
	if "udp" == network {
		pconn, err = net.ListenPacket(network, address)
	} else {
		ln, err = net.Listen(network, address)
		if nil != err {
			return nil, err
		}
	}

	// step 2: get file
	if nil != ln {
		switch ln := ln.(type) {
		case *net.TCPListener:
			file, err = ln.File()
			lnAddr = ln.Addr()
		case *net.UnixListener:
			file, err = ln.File()
			lnAddr = ln.Addr()
		default:
			return nil, fmt.Errorf("unknow protocol, network: %s, addr: %s", network, addr)
		}
	} else {
		switch pconn := pconn.(type) {
		case *net.UDPConn:
			file, err = pconn.File()
			lnAddr = pconn.LocalAddr()
		default:
			return nil, fmt.Errorf("unknow protocol, network: %s, addr: %s", network, addr)
		}
	}

	if nil != err {
		return nil, err
	}

	listener := &listener{
		ln:     ln,
		lnAddr: lnAddr,
		pconn:  pconn,

		file: file,
		fd:   int(file.Fd()),

		network: network,
		addr:    addr,
	}
	return listener, nil
}

func parseAddr(addr string) (network, address string) {
	// TODO:
	return "", ""
}

func (l *listener) close() {
	if 0 != l.fd {
		syscall.Close(l.fd)
	}

	if nil != l.file {
		l.file.Close()
	}

	if nil != l.ln {
		l.ln.Close()
	}

	if nil != l.pconn {
		l.pconn.Close()
	}

	if "unix" == l.network {
		os.RemoveAll(l.addr)
	}
}
