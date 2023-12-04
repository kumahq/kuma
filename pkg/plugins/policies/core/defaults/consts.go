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
	DefaultMaxStreamDuration     = 0
	DefaultMaxConnectionDuration = 0
)
