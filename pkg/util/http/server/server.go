package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/go-logr/logr"
)

func StartServer(log logr.Logger, server *http.Server, ready *atomic.Bool, errChan chan error) error {
	ready.Store(false)
	listener, err := NewListener(server)
	if err != nil {
		return err
	}
	ServeServer(log, server, listener, ready, errChan)
	return nil
}

func NewListener(server *http.Server) (net.Listener, error) {
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", server.Addr)
	if err != nil {
		return nil, err
	}
	if server.TLSConfig != nil {
		listener = tls.NewListener(listener, server.TLSConfig)
	}
	return listener, nil
}

func ServeServer(log logr.Logger, server *http.Server, listener net.Listener, ready *atomic.Bool, errChan chan error) <-chan struct{} {
	l := log.WithValues("tls", server.TLSConfig != nil, "interface", server.Addr)
	l.Info("starting server")
	ready.Store(true)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer ready.Store(false)
		if err := server.Serve(listener); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				l.Info("shutting down server")
			} else {
				l.Error(err, "could not start server")
				errChan <- err
			}
		}
	}()
	return done
}
