package dns

import (
	"net"
	"sync"
	"time"
)

type LookupIPFunc func(string) ([]net.IP, error)

type cacheRecord struct {
	ips          []net.IP
	creationTime time.Time
}

func MakeCaching(f LookupIPFunc, ttl time.Duration) LookupIPFunc {
	cache := map[string]*cacheRecord{}
	var rwmux sync.RWMutex
	return func(host string) ([]net.IP, error) {
		rwmux.RLock()
		r, ok := cache[host]
		rwmux.RUnlock()

		if ok && r.creationTime.Add(ttl).After(time.Now()) {
			return r.ips, nil
		}

		ips, err := f(host)
		if err != nil {
			return nil, err
		}

		rwmux.Lock()
		cache[host] = &cacheRecord{ips: ips, creationTime: time.Now()}
		rwmux.Unlock()

		return ips, nil
	}
}
