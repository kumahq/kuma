package httpclient

import (
	"context"
	"net"
	"net/http"
	"time"
)

// NewUDS returns an http.Client that dials the given Unix domain socket.
func NewUDS(socketPath string, dialTimeout, clientTimeout time.Duration) http.Client {
	dialer := &net.Dialer{Timeout: dialTimeout}
	return http.Client{
		Timeout: clientTimeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, "unix", socketPath)
			},
		},
	}
}

// NewTCPOrUDS returns a UDS client when socketPath is non-empty,
// otherwise a plain TCP client with the given timeout.
func NewTCPOrUDS(socketPath string, dialTimeout, clientTimeout time.Duration) http.Client {
	if socketPath != "" {
		return NewUDS(socketPath, dialTimeout, clientTimeout)
	}
	return http.Client{Timeout: clientTimeout}
}
