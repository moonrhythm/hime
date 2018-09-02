package hime

import (
	"net"
	"time"
)

// tcpKeepAliveListener edited from http.tcpKeepAliveListener
type tcpKeepAliveListener struct {
	*net.TCPListener
	period time.Duration
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(ln.period)
	return tc, nil
}
