package lookup

import (
	"github.com/kumahq/kuma/pkg/core"
	"net"
	"sync"
	"time"
)

type cacheRecord struct {
	ips          []net.IP
	creationTime time.Time
}

func CachedLookupIP(f LookupIPFunc, ttl time.Duration) LookupIPFunc {
	cache := map[string]*cacheRecord{}
	var rwmux sync.RWMutex
	return func(host string) ([]net.IP, error) {
		rwmux.RLock()
		r, ok := cache[host]
		rwmux.RUnlock()

		if ok && r.creationTime.Add(ttl).After(core.Now()) {
			return r.ips, nil
		}

		ips, err := f(host)
		if err != nil {
			return nil, err
		}

		rwmux.Lock()
		cache[host] = &cacheRecord{ips: ips, creationTime: core.Now()}
		rwmux.Unlock()

		return ips, nil
	}
}
