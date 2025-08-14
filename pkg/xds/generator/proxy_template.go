package generator

import (
	"context"
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_generator "github.com/kumahq/kuma/pkg/core/resources/apis/core/generator"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/generator"
	gateway_metadata "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	util_envoy "github.com/kumahq/kuma/pkg/util/envoy"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator/core"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
	generator_secrets "github.com/kumahq/kuma/pkg/xds/generator/secrets"
	"github.com/kumahq/kuma/pkg/xds/template"
)

type ProxyTemplateGenerator struct {
	ProxyTemplate *mesh_proto.ProxyTemplate
}

func (g *ProxyTemplateGenerator) Generate(ctx context.Context, xdsCtx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	for i, name := range g.ProxyTemplate.GetConf().GetImports() {
		generator := &ProxyTemplateProfileSource{ProfileName: name}
		if rs, err := generator.Generate(ctx, resources, xdsCtx, proxy); err != nil {
			return nil, fmt.Errorf("imports[%d]{name=%q}: %s", i, name, err)
		} else {
			resources.AddSet(rs)
		}
	}
	generator := &ProxyTemplateRawSource{Resources: g.ProxyTemplate.GetConf().GetResources()}
	if rs, err := generator.Generate(xdsCtx, proxy); err != nil {
		return nil, fmt.Errorf("resources: %s", err)
	} else {
		resources.AddSet(rs)
	}
	return resources, nil
}

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
			Origin:   metadata.OriginProxyTemplateRaw,
			Resource: res,
		})
	}
	return resources, nil
}

type ProxyTemplateProfileSource struct {
	ProfileName string
}

func (s *ProxyTemplateProfileSource) Generate(ctx context.Context, rs *model.ResourceSet, xdsCtx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	g, ok := predefinedProfiles[s.ProfileName]
	if !ok {
		return nil, fmt.Errorf("profile{name=%q}: unknown profile", s.ProfileName)
	}
	return g.Generate(ctx, rs, xdsCtx, proxy)
}

func NewDefaultProxyProfile() core.ResourceGenerator {
	return core.CompositeResourceGenerator{
		AdminProxyGenerator{},
		PrometheusEndpointGenerator{},
		TransparentProxyGenerator{},
		InboundProxyGenerator{},
		OutboundProxyGenerator{},
		DirectAccessProxyGenerator{},
		TracingProxyGenerator{},
		ProbeProxyGenerator{},
		DNSGenerator{},
		generator.NewGenerator(),
		generator_secrets.Generator{},
		core_generator.NewGenerator(),
	}
}

func NewEgressProxyProfile() core.ResourceGenerator {
	return core.CompositeResourceGenerator{
		AdminProxyGenerator{},
		egress.Generator{
			ZoneEgressGenerators: []egress.ZoneEgressGenerator{
				&egress.InternalServicesGenerator{},
				&egress.ExternalServicesGenerator{},
			},
			SecretGenerator: &generator_secrets.Generator{},
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

var predefinedProfiles = make(map[string]core.ResourceGenerator)

func init() {
	RegisterProfile(core_mesh.ProfileDefaultProxy, NewDefaultProxyProfile())
	RegisterProfile(IngressProxy, core.CompositeResourceGenerator{AdminProxyGenerator{}, IngressGenerator{}})
	RegisterProfile(egress.EgressProxy, NewEgressProxyProfile())
	// we register this so that kumactl does not fail validation of profiles registered by plugins (only "gateway-proxy" for now)
	// a proper solution for this is to rewrite as a custom ResourceManager
	// TODO: https://github.com/kumahq/kuma/issues/5144
	RegisterProfile(gateway_metadata.ProfileGatewayProxy, NewFailingProfile())
}

type FailingResourceGenerator struct{}

func (c FailingResourceGenerator) Generate(context.Context, *model.ResourceSet, xds_context.Context, *model.Proxy) (*model.ResourceSet, error) {
	panic("generator for this resource should not be called")
}

func NewFailingProfile() core.ResourceGenerator {
	return FailingResourceGenerator{}
}

func RegisterProfile(profileName string, generator core.ResourceGenerator) {
	predefinedProfiles[profileName] = generator
	core_mesh.AvailableProfiles[profileName] = struct{}{}
}
