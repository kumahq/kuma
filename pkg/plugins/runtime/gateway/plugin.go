package gateway

import (
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/template"
)

// OriginGateway marks xDS resources generated by this plugin.
const OriginGateway = "gateway"

const PluginName core_plugins.PluginName = "gateway"

func init() {
	core_plugins.Register(PluginName, &plugin{})
}

var (
	log = core.Log.WithName("plugin").WithName("runtime").WithName("gateway")
)

type plugin struct{}

var _ core_plugins.BootstrapPlugin = &plugin{}

func (p *plugin) BeforeBootstrap(context *core_plugins.MutablePluginContext, config core_plugins.PluginConfig) error {
	if !context.Config().Experimental.Gateway {
		log.V(1).Info("gateway plugin is disabled")
		return nil
	}

	register.RegisterGatewayTypes()
	if context.Config().Environment == config_core.KubernetesEnvironment {
		mesh_k8s.RegisterK8SGatewayTypes()
	}
	return nil
}

func (p *plugin) AfterBootstrap(context *core_plugins.MutablePluginContext, config core_plugins.PluginConfig) error {
	if !context.Config().Experimental.Gateway {
		log.V(1).Info("gateway plugin is disabled")
		return nil
	}

	// Insert our resolver before the default so that we can intercept
	// builtin gateway dataplanes.
	generator.DefaultTemplateResolver = template.SequentialResolver(
		TemplateResolver{},
		generator.DefaultTemplateResolver,
	)

	generator.RegisterProfile(ProfileGatewayProxy, NewProxyProfile(context.Config().Multizone.Zone.Name, context.DataSourceLoader()))

	log.Info("registered gateway plugin")
	return nil
}

func (p *plugin) Name() core_plugins.PluginName {
	return PluginName
}

func (p *plugin) Order() int {
	// It has to go before Environment is prepared, so we have resources registered in K8S schema
	return core_plugins.EnvironmentPreparingOrder - 1
}

// ProfileGatewayProxy is the name of the gateway proxy template profile.
const ProfileGatewayProxy = "gateway-proxy"

// NewProxyProfile returns a new resource generator profile for builtin
// gateway dataplanes.
func NewProxyProfile(zone string, dataSourceLoader datasource.Loader) generator.ResourceGenerator {
	return generator.CompositeResourceGenerator{
		generator.AdminProxyGenerator{},
		generator.PrometheusEndpointGenerator{},
		generator.SecretsProxyGenerator{},
		generator.TracingProxyGenerator{},
		generator.TransparentProxyGenerator{},
		generator.DNSGenerator{},

		Generator{
			ListenerGenerator: ListenerGenerator{},
			Generators: []GatewayHostGenerator{
				// The order here matters because generators can
				// depend on state created by a previous generator.
				&HTTPFilterChainGenerator{},
				&HTTPSFilterChainGenerator{
					DataSourceLoader: dataSourceLoader,
				},
				&RouteConfigurationGenerator{},
				&GatewayRouteGenerator{},
				&ConnectionPolicyGenerator{},
				&ClusterGenerator{
					DataSourceLoader: dataSourceLoader,
					Zone:             zone,
				},
				&RouteTableGenerator{},
			},
		},
	}
}
