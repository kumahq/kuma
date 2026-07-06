package matchers

import (
	"hash/fnv"

	"github.com/goburrow/cache"
	"github.com/prometheus/client_golang/prometheus"

	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
)

var _ core_plugins.PolicyMatchingCacheAccessor = &PolicyMatchingCache{}

// PolicyMatchingCache is a bounded LRU cache for TypedMatchingPolicies; safe for concurrent use.
type PolicyMatchingCache struct {
	c      cache.Cache
	metric *prometheus.CounterVec
}

func NewPolicyMatchingCache(metric *prometheus.CounterVec, maxSize int) *PolicyMatchingCache {
	c := cache.New(
		cache.WithMaximumSize(maxSize),
	)
	return &PolicyMatchingCache{c: c, metric: metric}
}

func (p *PolicyMatchingCache) GetIfPresent(key string) (core_xds.TypedMatchingPolicies, bool) {
	v, ok := p.c.GetIfPresent(key)
	if !ok {
		p.metric.WithLabelValues("miss").Inc()
		return core_xds.TypedMatchingPolicies{}, false
	}
	p.metric.WithLabelValues("hit").Inc()
	return v.(core_xds.TypedMatchingPolicies), true
}

func (p *PolicyMatchingCache) Put(key string, value core_xds.TypedMatchingPolicies) {
	p.c.Put(key, value)
}

func BuildCacheKey(rType string, cfg *core_plugins.MatchedPoliciesConfig, dpp *core_mesh.DataplaneResource) string {
	h := fnv.New128a()
	_, _ = h.Write([]byte(rType))
	if cfg.IncludeShadow {
		_, _ = h.Write([]byte{1})
	} else {
		_, _ = h.Write([]byte{0})
	}
	_, _ = h.Write(dpp.Hash())
	_, _ = h.Write([]byte(cfg.PolicyMatchingHash))
	return string(h.Sum(nil))
}
