package k8s

import (
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	k8s_registry "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/patrickmn/go-cache"
)

var _ k8s_extensions.Converter = &cachingConverter{}

type cachingConverter struct {
	SimpleConverter
	cache *cache.Cache
}

func NewCachingConverter(expirationTime time.Duration) k8s_extensions.Converter {
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
	out.SetMeta(&KubernetesMetaAdapter{ObjectMeta: *obj.GetObjectMeta(), Mesh: obj.GetMesh()})
	key := strings.Join([]string{
		obj.GetNamespace(),
		obj.GetName(),
		obj.GetMesh(),
		obj.GetResourceVersion(),
		proto.MessageName(out.GetSpec()),
	}, ":")
	if v, ok := c.cache.Get(key); ok {
		return out.SetSpec(v.(core_model.ResourceSpec))
	}
	if err := util_proto.FromMap(obj.GetSpec(), out.GetSpec()); err != nil {
		return err
	}
	c.cache.SetDefault(key, out.GetSpec())
	return nil
}
