package generator

import (
	"fmt"
	"github.com/Kong/kuma/pkg/xds/generator/modifications"
	"github.com/pkg/errors"

	kuma_mesh "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	util_envoy "github.com/Kong/kuma/pkg/util/envoy"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
)

type ProxyTemplateGenerator struct {
	ProxyTemplate *kuma_mesh.ProxyTemplate
}

func (g *ProxyTemplateGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := &model.ResourceSet{}
	for i, name := range g.ProxyTemplate.GetConf().GetImports() {
		generator := &ProxyTemplateProfileSource{ProfileName: name}
		if rs, err := generator.Generate(ctx, proxy); err != nil {
			return nil, fmt.Errorf("imports[%d]{name=%q}: %s", i, name, err)
		} else {
			resources.AddSet(rs)
		}
	}
	generator := &ProxyTemplateRawSource{Resources: g.ProxyTemplate.GetConf().GetResources()}
	if rs, err := generator.Generate(ctx, proxy); err != nil {
		return nil, fmt.Errorf("resources: %s", err)
	} else {
		resources.AddSet(rs)
	}
	if err := modifications.Apply(resources, g.ProxyTemplate.GetConf().GetModifications()); err != nil {
		return nil, errors.Wrap(err, "could not apply modifications")
	}
	return resources, nil
}

type ProxyTemplateRawSource struct {
	Resources []*kuma_mesh.ProxyTemplateRawResource
}

func (s *ProxyTemplateRawSource) Generate(_ xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := &model.ResourceSet{}
	for i, r := range s.Resources {
		res, err := util_envoy.ResourceFromYaml(r.Resource)
		if err != nil {
			return nil, fmt.Errorf("raw.resources[%d]{name=%q}.resource: %s", i, r.Name, err)
		}

		resources.Add(&model.Resource{
			Name:     r.Name,
			Version:  r.Version,
			Resource: res,
		})
	}
	return resources, nil
}

var predefinedProfiles = make(map[string]ResourceGenerator)

func NewDefaultProxyProfile() ResourceGenerator {
	return CompositeResourceGenerator{PrometheusEndpointGenerator{}, TransparentProxyGenerator{}, InboundProxyGenerator{}, OutboundProxyGenerator{}, DirectAccessProxyGenerator{}}
}

func init() {
	predefinedProfiles[mesh_core.ProfileDefaultProxy] = NewDefaultProxyProfile()
	predefinedProfiles[IngressProxy] = &IngressGenerator{}
}

type ProxyTemplateProfileSource struct {
	ProfileName string
}

func (s *ProxyTemplateProfileSource) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	g, ok := predefinedProfiles[s.ProfileName]
	if !ok {
		return nil, fmt.Errorf("profile{name=%q}: unknown profile", s.ProfileName)
	}
	return g.Generate(ctx, proxy)
}
