package envoy

import (
	"time"
)

// Timeouts carries the timeouts applied to a cluster and its filter chain.
// A zero value means the timeout is left unset, which is why every accessor
// tolerates a nil receiver.
type Timeouts struct {
	// Connect is the time to establish a connection to the upstream.
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
	// HttpMaxStreamDuration is the maximum lifetime of a stream.
	HttpMaxStreamDuration time.Duration
}

func (t *Timeouts) ConnectOrDefault(defaultConnectTimeout time.Duration) time.Duration {
	if t == nil || t.Connect == 0 {
		return defaultConnectTimeout
	}
	return t.Connect
}

func (t *Timeouts) GetTcpIdle() time.Duration {
	if t == nil {
		return 0
	}
	return t.TcpIdle
}

func (t *Timeouts) GetHttpIdle() time.Duration {
	if t == nil {
		return 0
	}
	return t.HttpIdle
}

func (t *Timeouts) GetHttpStreamIdle() time.Duration {
	if t == nil {
		return 0
	}
	return t.HttpStreamIdle
}

func (t *Timeouts) GetHttpMaxStreamDuration() time.Duration {
	if t == nil {
		return 0
	}
	return t.HttpMaxStreamDuration
}
