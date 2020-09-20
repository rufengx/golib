package xio

import "time"

type Option struct {
	ReusePort    bool
	TCPKeepAlive time.Duration
}
