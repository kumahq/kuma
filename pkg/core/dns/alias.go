package dns

import (
	"net"
	"time"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
)

const (
	DefaultTTL = 10 * time.Second
)

var (
	LookupIP = lookup.CachedLookupIP(net.LookupIP, DefaultTTL)
)
