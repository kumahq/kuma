package k8s

import (
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

<<<<<<< HEAD
func NewCachingConverter(expirationTime time.Duration) k8s_common.Converter {
=======
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

func NewCachingConverter(expirationTime time.Duration, systemNamespace string) k8s_common.Converter {
>>>>>>> 462d97e142 (fix(k8s): enforce control-plane-owned labels on read (backport of #18271) (#18281))
	return &cachingConverter{
		SimpleConverter: SimpleConverter{
			KubeFactory: &SimpleKubeFactory{
				KubeTypes: k8s_registry.Global(),
			},
		},
		SystemNamespace: systemNamespace,
		cache:           cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
	}
}

func (c *cachingConverter) ToCoreResource(obj k8s_model.KubernetesObject, out core_model.Resource) error {
	out.SetMeta(&KubernetesMetaAdapter{ObjectMeta: *obj.GetObjectMeta(), Mesh: obj.GetMesh()})
	key := strings.Join([]string{
		obj.GetNamespace(),
		obj.GetName(),
		obj.GetResourceVersion(),
		obj.GetObjectKind().GroupVersionKind().String(),
	}, ":")
	if v, ok := c.cache.Get(key); ok {
<<<<<<< HEAD
		return out.SetSpec(v.(core_model.ResourceSpec))
=======
		entry := v.(cachedEntry)
		// Reuse the labels computed on the miss - enforcement included - as a fresh
		// clone, so the cached map stays isolated from downstream consumers that
		// mutate labels in place (e.g. removeDisplayNameLabel in the ServiceInsight
		// endpoints).
		out.SetMeta(newMetaAdapterWithLabels(obj, maps.Clone(entry.labels)))
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
>>>>>>> 462d97e142 (fix(k8s): enforce control-plane-owned labels on read (backport of #18271) (#18281))
	}
	spec, err := obj.GetSpec()
	if err != nil {
		return err
	}
	// SetSpec first, then derive labels from out.GetSpec(): a stored object with an
	// omitted spec yields a typed-nil here, and SetSpec normalizes it to an empty
	// spec. Deriving from the raw value would call policy methods on a nil pointer.
	if err := out.SetSpec(spec); err != nil {
		return err
	}
	adapter := newMetaAdapter(obj, c.SystemNamespace, out.Descriptor(), out.GetSpec())
	out.SetMeta(adapter)
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
		c.cache.SetDefault(key, out.GetSpec())
	}
	return nil
}
