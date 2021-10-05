package gateway

import (
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/template"
)

// OriginGateway marks xDS resources generated by this plugin.
const OriginGateway = "gateway"

var (
	log = core.Log.WithName("plugin").WithName("runtime").WithName("gateway")
)

type plugin struct{}

var _ core_plugins.RuntimePlugin = &plugin{}

func (p *plugin) Customize(rt core_runtime.Runtime) error {
	// Insert our resolver before the default so that we can intercept
	// builtin gateway dataplanes.
	generator.DefaultTemplateResolver = template.SequentialResolver(
		TemplateResolver{},
		generator.DefaultTemplateResolver,
	)

	generator.RegisterProfile(ProfileGatewayProxy, NewProxyProfile(rt.ReadOnlyResourceManager()))

	// TODO(jpeach) As new gateway resources are added, register them here.

	log.Info("registered gateway plugin")
	return nil
}

// ProfileGatewayProxy is the name of the gateway proxy template profile.
const ProfileGatewayProxy = "gateway-proxy"

// NewProxyProfile returns a new resource generator profile for builtin
// gateway dataplanes.
func NewProxyProfile(manager manager.ReadOnlyResourceManager) generator.ResourceGenerator {
	return generator.CompositeResourceGenerator{
		generator.AdminProxyGenerator{},
		generator.PrometheusEndpointGenerator{},
		generator.TracingProxyGenerator{},
		generator.TransparentProxyGenerator{},
		generator.DNSGenerator{},

		Generator{
			ResourceManager: manager,
			Generators: []GatewayHostGenerator{
				// The order here matters because generators can
				// depend on state created by a previous generator.
				&ListenerGenerator{},
				&RouteConfigurationGenerator{},
				&GatewayRouteGenerator{},
				&RouteTableGenerator{},
			},
		},
	}
}
