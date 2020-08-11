package dns

import (
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"net"
	"time"
)

const (
	DefaultTTL = 10 * time.Second
)

var (
	LookupIP = lookup.CachedLookupIP(net.LookupIP, DefaultTTL)
)
