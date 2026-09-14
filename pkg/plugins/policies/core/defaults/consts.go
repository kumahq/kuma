package defaults

import (
	"time"

	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

// Timeouts
const (
	DefaultConnectTimeout        = 5 * time.Second
	DefaultIdleTimeout           = time.Hour
	DefaultStreamIdleTimeout     = 30 * time.Minute
	DefaultRequestTimeout        = 15 * time.Second
	DefaultRequestHeadersTimeout = 0
	DefaultMaxStreamDuration     = 0
	DefaultMaxConnectionDuration = 0
)

// Retries
const (
	DefaultRetryBaseInterval = 25 * time.Millisecond
)

// Timeouts are the values the control plane writes when no MeshTimeout selects
// the listener, cluster or route it is configuring. A zero value is written as
// an explicit `0s`, which disables that timeout in Envoy rather than leaving
// Envoy's own default in place.
type Timeouts struct {
	Connect               time.Duration
	Idle                  time.Duration
	Request               time.Duration
	StreamIdle            time.Duration
	RequestHeaders        time.Duration
	MaxStreamDuration     time.Duration
	MaxConnectionDuration time.Duration
}

// OutboundTimeouts is what a request gets on the side that sends it.
var OutboundTimeouts = Timeouts{
	Connect:               DefaultConnectTimeout,
	Idle:                  DefaultIdleTimeout,
	Request:               DefaultRequestTimeout,
	StreamIdle:            DefaultStreamIdleTimeout,
	RequestHeaders:        DefaultRequestHeadersTimeout,
	MaxStreamDuration:     DefaultMaxStreamDuration,
	MaxConnectionDuration: DefaultMaxConnectionDuration,
}

// InboundTimeouts is what a request gets on the side that receives it. The
// values are deliberately larger than the outbound ones, or disabled, so the
// receiving side never cuts a request the sending side still allows.
var InboundTimeouts = Timeouts{
	Connect:               inboundFactor * DefaultConnectTimeout,
	Idle:                  inboundFactor * DefaultIdleTimeout,
	Request:               0,
	StreamIdle:            inboundFactor * DefaultStreamIdleTimeout,
	RequestHeaders:        DefaultRequestHeadersTimeout,
	MaxStreamDuration:     DefaultMaxStreamDuration,
	MaxConnectionDuration: DefaultMaxConnectionDuration,
}

const inboundFactor = 2

// Envoy is the same set of timeouts in the shape the cluster and listener
// builders take, for the resources the control plane writes before any policy
// is applied.
func (t Timeouts) Envoy() envoy_common.Timeouts {
	return envoy_common.Timeouts{
		Connect:        t.Connect,
		TcpIdle:        t.Idle,
		HttpIdle:       t.Idle,
		HttpStreamIdle: t.StreamIdle,
	}
}
