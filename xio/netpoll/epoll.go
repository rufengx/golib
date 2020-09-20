package netpoll

import "syscall"

// 1. epoll
// I/O event notification facility.
// See: http://man7.org/linux/man-pages/man7/epoll.7.html

// 2. epoll_create
// Open an epoll file descriptor.
// See: http://man7.org/linux/man-pages/man2/epoll_create.2.html

// 3. epoll_ctl
// Control interface for an epoll file descriptor.
// See: http://man7.org/linux/man-pages/man2/epoll_ctl.2.html

// 4. epoll_wait
// Wait  for  an I/O event on an epoll file descriptor.
// See: http://man7.org/linux/man-pages/man2/epoll_wait.2.html

type Poll struct {
	epfd   int // representative epoll instance.
	wakeFd int // event notification fd, use in user space application. See: https://linux.die.net/man/2/eventfd2
	note   noteQueue
}

func Create() (*Poll, error) {
	// If flags is 0, then, other than the fact that the obsolete size
	// argument is dropped, epoll_create1() is the same as epoll_create().
	epfd, err := syscall.EpollCreate1(0)
	if nil != err || -1 == epfd {
		panic(err)
	}

	wakeFd, _, errno := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if 0 != errno {
		syscall.Close(epfd)
		panic(err)
	}

	poll := &Poll{
		epfd:   epfd,
		wakeFd: int(wakeFd),
	}

	poll.AddRead(poll.wakeFd)
	return poll, nil
}

func (p *Poll) AddRead(fd int) {
	event := &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: syscall.EPOLLIN, // read
	}

	err := syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_ADD, fd, event)
	if nil != err {
		panic(err)
	}
}

func (p *Poll) AddReadWrite(fd int) {
	event := &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: syscall.EPOLLIN | syscall.EPOLLOUT, // read & write
	}

	err := syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_ADD, fd, event)
	if nil != err {
		panic(err)
	}
}

func (p *Poll) ModifyRead(fd int) {
	event := &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: syscall.EPOLLIN, // read
	}

	err := syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_MOD, fd, event)
	if nil != err {
		panic(err)
	}
}

func (p *Poll) ModifyReadWrite(fd int) {
	event := &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: syscall.EPOLLIN | syscall.EPOLLOUT, // read & write
	}

	err := syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_MOD, fd, event)
	if nil != err {
		panic(err)
	}

}

func (p *Poll) RemoveRead(fd int) {
	event := &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: syscall.EPOLLIN, // read
	}

	err := syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_DEL, fd, event)
	if nil != err {
		panic(err)
	}
}

func (p *Poll) RemoveReadWrite(fd int) {
	event := &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: syscall.EPOLLIN | syscall.EPOLLOUT, // read & write
	}

	err := syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_DEL, fd, event)
	if nil != err {
		panic(err)
	}
}

func (p *Poll) Wait(callback func(fd int, ev uint32) error) error {
	events := make([]syscall.EpollEvent, 64)
	for {
		nfds, err := syscall.EpollWait(p.epfd, events, 100)

		// EINTR:
		// The call was interrupted by a signal handler before either
		// any of the requested events occurred or the timeout expired.
		if nfds == -1 || (nil != err && err != syscall.EINTR) {
			return err
		}

		for i := 0; i < nfds; i++ {
			fd := int(events[i].Fd)
			if fd != p.wakeFd {
				err := callback(fd, events[i].Events)
				if nil != err {
					return err
				}
			} else if fd == p.wakeFd {
				var data [8]byte
				syscall.Read(p.wakeFd, data[:])
			}
		}
	}
}

func (p *Poll) Trigger() {
}

func (p *Poll) Close() error {
	err := syscall.Close(p.wakeFd)
	if nil != err {
		return err
	}
	return syscall.Close(p.epfd)
}
