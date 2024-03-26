package server

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/go-logr/logr"
)

func StartServer(log logr.Logger, server *http.Server, ready *atomic.Bool, errChan chan error) error {
	l := log.WithValues("tls", server.TLSConfig != nil, "interface", server.Addr)
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	if server.TLSConfig != nil {
		listener = tls.NewListener(listener, server.TLSConfig)
	}

	l.Info("starting server")
	go func() {
		ready.Store(true)
		if err := server.Serve(listener); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				l.Info("shutting down server")
			} else {
				l.Error(err, "could not start server")
				errChan <- err
			}
		}
	}()
	return nil
}
