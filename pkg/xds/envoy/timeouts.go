package envoy

import (
	"time"
)

// Timeouts carries the timeouts applied to a cluster and its filter chain.
// Only Connect treats zero as "not set", see ConnectOrDefault. The idle
// timeouts are always written, so a zero value reaches Envoy as an explicit
// `0s`, which disables that timeout rather than leaving it at Envoy's default.
type Timeouts struct {
	// Connect is the time to establish a connection to the upstream. Zero
	// falls back to the default passed to ConnectOrDefault.
	Connect time.Duration
	// TcpIdle is the period without bytes sent or received on either the
	// upstream or the downstream connection.
	TcpIdle time.Duration
	// HttpIdle is the time after which a connection with no active streams
	// is terminated.
	HttpIdle time.Duration
	// HttpStreamIdle is the time a stream is allowed to exist with no
	// upstream or downstream activity.
	HttpStreamIdle time.Duration
}

func (t Timeouts) ConnectOrDefault(defaultConnectTimeout time.Duration) time.Duration {
	if t.Connect == 0 {
		return defaultConnectTimeout
	}
	return t.Connect
}
