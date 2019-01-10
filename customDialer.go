package main

import (
	"net"
	"syscall"
	"time"
)

type CustomDialer struct {
	dialer net.Dialer
}

func NewDialer(localAddr string) (customDialor *CustomDialer, err error) {
	laddr, err := net.ResolveTCPAddr("tcp", localAddr)

	d := net.Dialer{
		Timeout:   2 * time.Second,
		LocalAddr: laddr,
		Control:   controlSetup,
	}

	customDialor = &CustomDialer{
		dialer: d,
	}

	return
}

func controlSetup(network string, address string, c syscall.RawConn) error {
	var operr error

	fn := func(fd uintptr) {
		i := int(fd)
		operr = syscall.SetsockoptInt(i, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		operr = syscall.SetsockoptInt(i, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	}

	if err := c.Control(fn); err != nil {
		return err
	}

	if operr != nil {
		return operr
	}

	return nil

}
