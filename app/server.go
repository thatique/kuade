package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/thatique/kuade/app/listener"
)

type Server struct {
	net    string // either tcp or unix
	server http.Server
}

func NewDefaultServer() *Server {
	return &Server{
		net: "tcp",
		server: http.Server{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

// SetNet set net iface for server to bind. Default is set to tcp.
func (srv *Server) SetNet(net string) error {
	switch net {
	case "unix", "tcp", "":
		srv.net = net
		return nil
	default:
		return errors.New("invalid net iface")
	}
}

func (srv *Server) ListenAndServe(addr string, h http.Handler) error {
	ln, err := listener.NewListener(srv.net, addr)
	if err != nil {
		return err
	}

	srv.server.Handler = h
	return srv.server.Serve(ln)
}

func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.server.Shutdown(ctx)
}
