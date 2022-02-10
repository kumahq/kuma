package generator

import (
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	util_envoy "github.com/kumahq/kuma/pkg/util/envoy"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
	"github.com/kumahq/kuma/pkg/xds/generator/modifications"
	"github.com/kumahq/kuma/pkg/xds/template"
)

type ProxyTemplateGenerator struct {
	ProxyTemplate *mesh_proto.ProxyTemplate
}

func (g *ProxyTemplateGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
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
	if err := modifications.Apply(resources, g.ProxyTemplate.GetConf().GetModifications(), proxy.APIVersion); err != nil {
		return nil, errors.Wrap(err, "could not apply modifications")
	}
	return resources, nil
}

// OriginProxyTemplateRaw is a marker to indicate by which ProxyGenerator resources were generated.
const OriginProxyTemplateRaw = "proxy-template-raw"

type ProxyTemplateRawSource struct {
	Resources []*mesh_proto.ProxyTemplateRawResource
}

func (s *ProxyTemplateRawSource) Generate(_ xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	for i, r := range s.Resources {
		res, err := util_envoy.ResourceFromYaml(r.Resource)
		if err != nil {
			return nil, fmt.Errorf("raw.resources[%d]{name=%q}.resource: %s", i, r.Name, err)
		}

		resources.Add(&model.Resource{
			Name:     r.Name,
			Origin:   OriginProxyTemplateRaw,
			Resource: res,
		})
	}
	return resources, nil
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

func NewDefaultProxyProfile() ResourceGenerator {
	return CompositeResourceGenerator{
		AdminProxyGenerator{},
		PrometheusEndpointGenerator{},
		SecretsProxyGenerator{},
		TransparentProxyGenerator{},
		InboundProxyGenerator{},
		OutboundProxyGenerator{},
		DirectAccessProxyGenerator{},
		TracingProxyGenerator{},
		ProbeProxyGenerator{},
		DNSGenerator{},
	}
}

func NewEgressProxyProfile() ResourceGenerator {
	return CompositeResourceGenerator{
		AdminProxyGenerator{},
		SecretsProxyGenerator{},
		egress.Generator{
			Generators: []egress.ZoneEgressGenerator{
				&egress.InternalServicesGenerator{},
				&egress.ExternalServicesGenerator{},
			},
		},
	}
}

// DefaultTemplateResolver is the default template resolver that xDS
// generators fall back to if they are otherwise unable to determine which
// ProxyTemplate resource to apply. Plugins may modify this variable.
var DefaultTemplateResolver template.ProxyTemplateResolver = &template.StaticProxyTemplateResolver{
	Template: &mesh_proto.ProxyTemplate{
		Conf: &mesh_proto.ProxyTemplate_Conf{
			Imports: []string{core_mesh.ProfileDefaultProxy},
		},
	},
}

var predefinedProfiles = make(map[string]ResourceGenerator)

func init() {
	RegisterProfile(core_mesh.ProfileDefaultProxy, NewDefaultProxyProfile())
	RegisterProfile(IngressProxy, CompositeResourceGenerator{AdminProxyGenerator{}, IngressGenerator{}})
	RegisterProfile(egress.EgressProxy, NewEgressProxyProfile())
}

func RegisterProfile(profileName string, generator ResourceGenerator) {
	predefinedProfiles[profileName] = generator
}
