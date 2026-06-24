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

// defaultPolicyMatchingCacheSize caps the number of cached entries. Each entry
// stores one TypedMatchingPolicies per (policy-type × dataplane) pair, so this
// bounds worst-case memory to O(policyTypes × dataplanes).
const defaultPolicyMatchingCacheSize = 10_000

var _ core_plugins.PolicyMatchingCacheAccessor = &PolicyMatchingCache{}

// PolicyMatchingCache is a bounded LRU cache for TypedMatchingPolicies results.
// It is safe for concurrent use.
type PolicyMatchingCache struct {
	c cache.Cache
}

// NewPolicyMatchingCache creates a PolicyMatchingCache wired to the given Prometheus metric.
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

// BuildCacheKey returns a compact cache key for a MatchedPolicies call.
// The key encodes: resource type, shadow flag, dataplane identity, and the
// policy-matching hash (which covers all matching-relevant mesh resources
// but excludes the Dataplane roster).
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
