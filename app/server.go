package app

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/thatique/kuade/app/listener"
)

type Server struct {
	server http.Server
}

func NewDefaultServer() *Server {
	return &Server{
		server: http.Server{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

func (srv *Server) ListenAndServe(addr string, h http.Handler) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	ln, err := listener.NewListener(host, port)
	if err != nil {
		return err
	}

	srv.server.Handler = h
	return srv.server.Serve(ln)
}

func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.server.Shutdown(ctx)
}
