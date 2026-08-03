package test

import (
	"context"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

// TCPProxy forwards TCP connections to a target address and can be stopped and restarted.
type TCPProxy struct {
	target string

	mu         sync.Mutex
	listenAddr string
	listener   net.Listener
	conns      map[net.Conn]struct{}
}

// NewTCPProxy starts a TCP proxy that forwards connections to target.
func NewTCPProxy(target string) (*TCPProxy, error) {
	proxy := &TCPProxy{
		target:     target,
		listenAddr: "127.0.0.1:0",
		conns:      map[net.Conn]struct{}{},
	}
	if err := proxy.Start(); err != nil {
		return nil, err
	}
	return proxy, nil
}

// Host returns the proxy listener host.
func (p *TCPProxy) Host() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	host, _, err := net.SplitHostPort(p.listenAddr)
	if err != nil {
		return ""
	}
	return host
}

// Port returns the proxy listener port.
func (p *TCPProxy) Port() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, port, err := net.SplitHostPort(p.listenAddr)
	if err != nil {
		return 0
	}
	parsed, err := strconv.Atoi(port)
	if err != nil {
		return 0
	}
	return parsed
}

// Start starts the proxy listener if it is not already running.
func (p *TCPProxy) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.listener != nil {
		return nil
	}
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", p.listenAddr)
	if err != nil {
		return err
	}
	p.listenAddr = listener.Addr().String()
	p.listener = listener
	go p.accept(listener)
	return nil
}

// Stop stops the proxy listener and closes active proxied connections.
func (p *TCPProxy) Stop() error {
	p.mu.Lock()
	listener := p.listener
	p.listener = nil
	conns := make([]net.Conn, 0, len(p.conns))
	for conn := range p.conns {
		conns = append(conns, conn)
	}
	p.conns = map[net.Conn]struct{}{}
	p.mu.Unlock()

	for _, conn := range conns {
		_ = conn.Close()
	}
	if listener == nil {
		return nil
	}
	return listener.Close()
}

func (p *TCPProxy) accept(listener net.Listener) {
	for {
		source, err := listener.Accept()
		if err != nil {
			return
		}
		go p.forward(source)
	}
}

func (p *TCPProxy) forward(source net.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	target, err := (&net.Dialer{}).DialContext(ctx, "tcp", p.target)
	if err != nil {
		_ = source.Close()
		return
	}

	if !p.track(source) {
		_ = source.Close()
		_ = target.Close()
		return
	}
	if !p.track(target) {
		p.untrack(source)
		_ = source.Close()
		_ = target.Close()
		return
	}
	defer func() {
		p.untrack(source)
		p.untrack(target)
		_ = source.Close()
		_ = target.Close()
	}()

	doneCh := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(target, source)
		doneCh <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(source, target)
		doneCh <- struct{}{}
	}()

	<-doneCh
	_ = source.Close()
	_ = target.Close()
	<-doneCh
}

func (p *TCPProxy) track(conn net.Conn) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.listener == nil {
		_ = conn.Close()
		return false
	}
	p.conns[conn] = struct{}{}
	return true
}

func (p *TCPProxy) untrack(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.conns, conn)
}
