package reconcile

import (
	"hash/fnv"
	"slices"
	"strings"
	"sync"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/kds"
)

// mappedResources memoizes the result of mapping and marshaling resources of a
// single type for a single set of KDS features. That work is the expensive part
// of generating a snapshot and it does not depend on which zone the snapshot is
// for, so zones that negotiated the same features share it rather than each
// repeating it. Entries are filled lazily, so a resource no zone asks for is
// never mapped.
type mappedResources struct {
	fingerprint uint64

	mu    sync.RWMutex
	byKey map[model.ResourceKey]envoy_types.Resource
}

func (m *mappedResources) get(key model.ResourceKey) (envoy_types.Resource, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res, ok := m.byKey[key]
	return res, ok
}

func (m *mappedResources) put(key model.ResourceKey, res envoy_types.Resource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byKey[key] = res
}

type mappedResourcesKey struct {
	typ      model.ResourceType
	features string
}

type mappedResourcesCache struct {
	mu      sync.Mutex
	entries map[mappedResourcesKey]*mappedResources
}

func newMappedResourcesCache() *mappedResourcesCache {
	return &mappedResourcesCache{entries: map[mappedResourcesKey]*mappedResources{}}
}

// entryFor returns the memo for a type and feature set, discarding what it held
// when the underlying resources have changed.
func (c *mappedResourcesCache) entryFor(typ model.ResourceType, features kds.Features, rs model.ResourceList) *mappedResources {
	key := mappedResourcesKey{typ: typ, features: featuresKey(features)}
	fingerprint := fingerprintOf(rs)

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.entries[key]; ok && entry.fingerprint == fingerprint {
		return entry
	}
	entry := &mappedResources{
		fingerprint: fingerprint,
		byKey:       map[model.ResourceKey]envoy_types.Resource{},
	}
	c.entries[key] = entry
	return entry
}

func featuresKey(features kds.Features) string {
	enabled := make([]string, 0, len(features))
	for feature, on := range features {
		if on {
			enabled = append(enabled, feature)
		}
	}
	slices.Sort(enabled)
	return strings.Join(enabled, ",")
}

// fingerprintOf identifies the content of a resource list by the versions of the
// resources in it. It is order independent because a store is not required to
// return a stable order, and it never marshals a resource, which is the work
// this cache exists to avoid.
func fingerprintOf(rs model.ResourceList) uint64 {
	var acc uint64
	var count uint64
	for _, r := range rs.GetItems() {
		h := fnv.New64a()
		meta := r.GetMeta()
		_, _ = h.Write([]byte(meta.GetMesh()))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(meta.GetName()))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(meta.GetVersion()))
		acc ^= h.Sum64()
		count++
	}
	return acc ^ (count * 0x9e3779b97f4a7c15)
}
