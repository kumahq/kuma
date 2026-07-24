package generator

import (
	"context"
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_generator "github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/generator"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/v3/pkg/core/xds"
	policies_generator "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/generator"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/core"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/egress"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
	generator_secrets "github.com/kumahq/kuma/v3/pkg/xds/generator/secrets"
	"github.com/kumahq/kuma/v3/pkg/xds/template"
)

type ProxyTemplateGenerator struct {
	ProxyTemplate *mesh_proto.ProxyTemplate
}

func (g *ProxyTemplateGenerator) Generate(ctx context.Context, xdsCtx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	if len(g.ProxyTemplate.GetConf().GetResources()) > 0 || len(g.ProxyTemplate.GetConf().GetModifications()) > 0 {
		return nil, fmt.Errorf("ProxyTemplate.Conf.Resources and ProxyTemplate.Conf.Modifications are no longer applied; use MeshProxyPatch instead")
	}
	resources := model.NewResourceSet()
	for i, name := range g.ProxyTemplate.GetConf().GetImports() {
		generator := &ProxyTemplateProfileSource{ProfileName: name}
		if rs, err := generator.Generate(ctx, resources, xdsCtx, proxy); err != nil {
			return nil, fmt.Errorf("imports[%d]{name=%q}: %s", i, name, err)
		} else {
			resources.AddSet(rs)
		}
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
		TransparentProxyGenerator{},
		InboundProxyGenerator{},
		DirectAccessProxyGenerator{},
		ProbeProxyGenerator{},
		DNSGenerator{},
		ZoneProxyListenerGenerator{},
		policies_generator.NewGenerator(),
		generator_secrets.Generator{},
		core_generator.NewGenerator(),
	}
}

func NewEgressProxyProfile() core.ResourceGenerator {
	return core.CompositeResourceGenerator{
		AdminProxyGenerator{},
		egress.Generator{
			SecretGenerator: &generator_secrets.Generator{},
			PolicyGenerator: policies_generator.NewGenerator(),
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
	RegisterProfile(metadata.ProxyTemplateProfileIngressProxy, core.CompositeResourceGenerator{AdminProxyGenerator{}, IngressGenerator{}})
	RegisterProfile(metadata.ProxyTemplateProfileEgressProxy, NewEgressProxyProfile())
}

func RegisterProfile(profileName string, generator core.ResourceGenerator) {
	predefinedProfiles[profileName] = generator
	core_mesh.AvailableProfiles[profileName] = struct{}{}
}
