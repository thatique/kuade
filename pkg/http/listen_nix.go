// +build linux darwin dragonfly freebsd netbsd openbsd rumprun
package http

import (
	"net"

	"github.com/valyala/tcplisten"
)

var cfg = &tcplistenConfig{
	DeferAccept: true,
	FastOpen:    true,
	Backlog:     2048,
}

// Unix listener with special TCP options
var listen = cfg.NewListener
var fallbackListen = net.Listen
