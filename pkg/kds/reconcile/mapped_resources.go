package reconcile

import (
	"hash/fnv"
	"slices"
	"strconv"
	"strings"
	"sync"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/kds"
)

// mappedResource is one memoized resource. The result is computed once even if
// several zones ask for it at the same moment, which they do, because the KDS
// watchdogs align their flushes.
type mappedResource struct {
	once sync.Once
	res  envoy_types.Resource
	err  error
}

// mappedResources memoizes the result of mapping and marshaling resources of a
// single type, for a single tenant and set of KDS features. That work is the
// expensive part of generating a snapshot and it does not depend on which zone
// the snapshot is for, so zones sharing those attributes share the result
// rather than each repeating it. Entries fill lazily, so a resource no zone
// asks for is never mapped.
type mappedResources struct {
	fingerprint uint64

	mu    sync.RWMutex
	byKey map[model.ResourceKey]*mappedResource
}

func (m *mappedResources) loadOrCompute(
	key model.ResourceKey,
	compute func() (envoy_types.Resource, error),
) (envoy_types.Resource, error) {
	m.mu.RLock()
	entry, ok := m.byKey[key]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		if entry, ok = m.byKey[key]; !ok {
			entry = &mappedResource{}
			m.byKey[key] = entry
		}
		m.mu.Unlock()
	}

	entry.once.Do(func() {
		entry.res, entry.err = compute()
	})

	if entry.err != nil {
		m.forget(key, entry)
		return nil, entry.err
	}
	return entry.res, nil
}

// forget drops a failed entry so that a later generation retries it instead of
// serving the failure for as long as the resources stay unchanged.
func (m *mappedResources) forget(key model.ResourceKey, entry *mappedResource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if current, ok := m.byKey[key]; ok && current == entry {
		delete(m.byKey, key)
	}
}

type mappedResourcesKey struct {
	typ      model.ResourceType
	tenant   string
	features string
}

type mappedResourcesCache struct {
	mu      sync.Mutex
	entries map[mappedResourcesKey]*mappedResources
}

func newMappedResourcesCache() *mappedResourcesCache {
	return &mappedResourcesCache{entries: map[mappedResourcesKey]*mappedResources{}}
}

// entryFor returns the memo for a type, tenant and feature set, discarding what
// it held when the underlying resources have changed.
func (c *mappedResourcesCache) entryFor(
	typ model.ResourceType,
	tenant string,
	features kds.Features,
	rs model.ResourceList,
) *mappedResources {
	key := mappedResourcesKey{typ: typ, tenant: tenant, features: featuresKey(features)}
	fingerprint := fingerprintOf(rs)

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.entries[key]; ok && entry.fingerprint == fingerprint {
		return entry
	}
	entry := &mappedResources{
		fingerprint: fingerprint,
		byKey:       map[model.ResourceKey]*mappedResource{},
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

// fingerprintOf identifies the content of a resource list by the identity of the
// resources in it. Creation time is part of that identity because a version is
// not unique across a delete and recreate: stores reset it, so the same name at
// the same version can be a different object. It is order independent, because a
// store is not required to return a stable order, and it never marshals a
// resource, which is the work this cache exists to avoid.
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
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(strconv.FormatInt(meta.GetCreationTime().UnixNano(), 10)))
		acc ^= h.Sum64()
		count++
	}
	return acc ^ (count * 0x9e3779b97f4a7c15)
}
