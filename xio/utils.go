package xio

import (
	"net"
	"syscall"
)

func socketAddrToAddr(sa *syscall.Sockaddr) (net.Addr, error) {
	// TODO:
	return nil, nil
}

func SetKeepAlive(fd, secs int) error {
	return nil
}
