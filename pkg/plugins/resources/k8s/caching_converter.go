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

func NewCachingConverter(expirationTime time.Duration, systemNamespace string) k8s_common.Converter {
	return &cachingConverter{
		SimpleConverter: SimpleConverter{
			KubeFactory: &SimpleKubeFactory{
				KubeTypes: k8s_registry.Global(),
			},
			SystemNamespace: systemNamespace,
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
		if err := out.SetSpec(v.(core_model.ResourceSpec)); err != nil {
			return err
		}
		// Only the spec is cached; the labels are derived on every read, enforcement
		// included, so a hit yields the same label set as the miss did.
		out.SetMeta(newMetaAdapter(obj, c.SystemNamespace, out.Descriptor(), out.GetSpec()))
		return nil
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
	out.SetMeta(newMetaAdapter(obj, c.SystemNamespace, out.Descriptor(), out.GetSpec()))
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
