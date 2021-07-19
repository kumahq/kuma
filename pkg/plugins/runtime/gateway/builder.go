package gateway

import (
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

// ResourceBuilder is an interface commonly implemented by complex Envoy
// configuration element builders.
type ResourceBuilder interface {
	Build() (envoy.NamedResource, error)
}

// BuildResourceSet is an adaptor that triggers the resource builder,
// b, to build its resource. If the builder is successful, the result is
// wrapped in a ResourceSet.
func BuildResourceSet(b ResourceBuilder) (*xds.ResourceSet, error) {
	resource, err := b.Build()
	if err != nil {
		return nil, err
	}

	set := xds.NewResourceSet()
	set.Add(&xds.Resource{
		Name:     resource.GetName(),
		Origin:   OriginGateway,
		Resource: resource,
	})

	return set, nil
}
