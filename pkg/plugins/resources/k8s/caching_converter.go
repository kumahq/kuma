package k8s

import (
	"maps"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	k8s_common "github.com/kumahq/kuma/v2/pkg/plugins/common/k8s"
	k8s_model "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/v2/pkg/plugins/resources/k8s/native/pkg/registry"
)

var _ k8s_common.Converter = &cachingConverter{}

// According to the profile, a huge amount of time is spent on marshaling of json objects.
// That's why having a cache on this execution path gives a big performance boost in Kubernetes.
type cachingConverter struct {
	SimpleConverter
	cache *cache.Cache
}

// cachedEntry is the value type stored in cachingConverter.cache. It bundles
// the two version-stable results of converting a Kubernetes object:
//   - spec: the unmarshaled protobuf spec
//   - labels: the precomputed KubernetesMetaAdapter.GetLabels() output
//
// Status is intentionally excluded: existing tests assert that mutations to
// status surface on subsequent cache hits, so it is fetched fresh every call.
type cachedEntry struct {
	spec   core_model.ResourceSpec
	labels map[string]string
}

func NewCachingConverter(expirationTime time.Duration) k8s_common.Converter {
	return &cachingConverter{
		SimpleConverter: SimpleConverter{
			KubeFactory: &SimpleKubeFactory{
				KubeTypes: k8s_registry.Global(),
			},
		},
		cache: cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
	}
}

func (c *cachingConverter) ToCoreResource(obj k8s_model.KubernetesObject, out core_model.Resource) error {
	key := strings.Join([]string{
		obj.GetNamespace(),
		obj.GetName(),
		obj.GetResourceVersion(),
		obj.GetObjectKind().GroupVersionKind().String(),
	}, ":")
	if v, ok := c.cache.Get(key); ok {
		entry := v.(cachedEntry)
		// Pre-populate cachedLabels with a fresh clone so the adapter's
		// GetLabels skips the annotation-merge work, while keeping the cached
		// map isolated from downstream consumers that mutate labels in place
		// (e.g. removeDisplayNameLabel in the ServiceInsight endpoints).
		out.SetMeta(&KubernetesMetaAdapter{
			ObjectMeta:   *obj.GetObjectMeta(),
			Mesh:         obj.GetMesh(),
			cachedLabels: maps.Clone(entry.labels),
		})
		if err := out.SetSpec(entry.spec); err != nil {
			return err
		}
		// Status is not cached (see cachedEntry comment); fetch from obj.
		if out.Descriptor().HasStatus {
			status, err := obj.GetStatus()
			if err != nil {
				return err
			}
			if err := out.SetStatus(status); err != nil {
				return err
			}
		}
		return nil
	}
	adapter := &KubernetesMetaAdapter{ObjectMeta: *obj.GetObjectMeta(), Mesh: obj.GetMesh()}
	out.SetMeta(adapter)
	spec, err := obj.GetSpec()
	if err != nil {
		return err
	}
	if err := out.SetSpec(spec); err != nil {
		return err
	}
	if out.Descriptor().HasStatus {
		status, err := obj.GetStatus()
		if err != nil {
			return err
		}
		if err := out.SetStatus(status); err != nil {
			return err
		}
	}
	if obj.GetResourceVersion() != "" {
		// an absence of the ResourceVersion means we decode 'obj' from webhook request,
		// all webhooks use SimpleConverter, so this is not supposed to happen
		// Clone the materialized labels so the cache holds a copy that is not
		// shared with the adapter we just handed back to the caller. Without
		// this, downstream in-place mutations (e.g. removeDisplayNameLabel)
		// would corrupt the cached entry on the very first access.
		c.cache.SetDefault(key, cachedEntry{
			spec:   out.GetSpec(),
			labels: maps.Clone(adapter.GetLabels()),
		})
	}
	return nil
}

// ToCoreList overrides SimpleConverter.ToCoreList so the per-item conversion
// dispatches to cachingConverter.ToCoreResource. Without this override, Go
// method resolution on the embedded SimpleConverter binds c.ToCoreResource to
// the inner type, bypassing the cache for every list read.
func (c *cachingConverter) ToCoreList(in k8s_model.KubernetesList, out core_model.ResourceList, predicate k8s_common.ConverterPredicate) error {
	for _, o := range in.GetItems() {
		r := out.NewItem()
		if err := c.ToCoreResource(o, r); err != nil {
			return err
		}
		if predicate(r) {
			_ = out.AddItem(r)
		}
	}
	return nil
}
