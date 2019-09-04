package generator

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/envoy"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"
	"github.com/pkg/errors"

	konvoy_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
)

type Profiles map[string]ResourceGenerator

func PredefinedProfiles(resourcesFactory envoy.EnvoyResourcesFactory) Profiles {
	profiles := Profiles{}
	profiles[template.ProfileDefaultProxy] = CompositeResourceGenerator{
		TransparentProxyGenerator{},
		&InboundProxyGenerator{
			ResourcesFactory: resourcesFactory,
		},
	}
	return profiles
}

func (p Profiles) GetProfile(source *konvoy_mesh.ProxyTemplateProfileSource) (ResourceGenerator, error) {
	res, ok := p[source.Name]
	if !ok {
		return nil, errors.Errorf("profile{name=%q}: unknown profile", source.Name)
	}
	return res, nil
}