package defaults

import (
	"time"
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
	// Gateway
	DefaultGatewayIdleTimeout           = 5 * time.Minute
	DefaultGatewayStreamIdleTimeout     = 5 * time.Second
	DefaultGatewayRequestHeadersTimeout = 500 * time.Millisecond
)
