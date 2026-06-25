package matchers

import (
	"encoding/base64"
	"hash/fnv"

	"github.com/goburrow/cache"
	"github.com/prometheus/client_golang/prometheus"

	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	util_cache "github.com/kumahq/kuma/v3/pkg/util/cache"
)

// bounds worst-case memory to O(policyTypes × dataplanes)
const defaultPolicyMatchingCacheSize = 10_000

var _ core_plugins.PolicyMatchingCacheAccessor = &PolicyMatchingCache{}

// PolicyMatchingCache is a bounded LRU cache for TypedMatchingPolicies; safe for concurrent use.
type PolicyMatchingCache struct {
	c cache.Cache
}

func NewPolicyMatchingCache(metric *prometheus.CounterVec) *PolicyMatchingCache {
	c := cache.New(
		cache.WithMaximumSize(defaultPolicyMatchingCacheSize),
		cache.WithStatsCounter(&util_cache.PrometheusStatsCounter{Metric: metric}),
	)
	return &PolicyMatchingCache{c: c}
}

func (p *PolicyMatchingCache) GetIfPresent(key string) (core_xds.TypedMatchingPolicies, bool) {
	v, ok := p.c.GetIfPresent(key)
	if !ok {
		return core_xds.TypedMatchingPolicies{}, false
	}
	return v.(core_xds.TypedMatchingPolicies), true
}

func (p *PolicyMatchingCache) Put(key string, value core_xds.TypedMatchingPolicies) {
	p.c.Put(key, value)
}

func BuildCacheKey(rType string, includeShadow bool, dpp *core_mesh.DataplaneResource, policyMatchingHash string) string {
	h := fnv.New128a()
	_, _ = h.Write([]byte(rType))
	if includeShadow {
		_, _ = h.Write([]byte{1})
	} else {
		_, _ = h.Write([]byte{0})
	}
	_, _ = h.Write(dpp.Hash())
	_, _ = h.Write([]byte(policyMatchingHash))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
