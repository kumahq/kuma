package globalloadbalancer

import (
	"errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/plugins/runtime/globalloadbalancer/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator"
	generator_core "github.com/kumahq/kuma/pkg/xds/generator/core"
	generator_secrets "github.com/kumahq/kuma/pkg/xds/generator/secrets"
	"github.com/kumahq/kuma/pkg/xds/template"
)

func init() {
	core_plugins.Register(metadata.PluginName, &plugin{})
}

var log = core.Log.WithName("plugin").WithName("runtime").WithName("global-load-balancer")

type plugin struct{}

var _ core_plugins.BootstrapPlugin = &plugin{}

func (p *plugin) BeforeBootstrap(context *core_plugins.MutablePluginContext, config core_plugins.PluginConfig) error {
	if context.Config().Environment == config_core.KubernetesEnvironment {
		return errors.New("kubernetes is unsupported")
	}

	return nil
}

func (p *plugin) AfterBootstrap(context *core_plugins.MutablePluginContext, config core_plugins.PluginConfig) error {
	// NOTE(nicoche): not sure if this is useful
	// Insert our resolver before the default so that we can intercept
	// builtin gateway dataplanes.
	generator.DefaultTemplateResolver = template.SequentialResolver(
		TemplateResolver{},
		generator.DefaultTemplateResolver,
	)

	generator.RegisterProfile(metadata.ProfileGlobalLoadBalancerProxy, NewProxyProfile(context.Config().Multizone.Zone.Name))

	log.Info("registered global-load-balancer plugin")
	return nil
}

func (p *plugin) Name() core_plugins.PluginName {
	return metadata.PluginName
}

func (p *plugin) Order() int {
	// NOTE(nicoche): this looks like a k8s setting. Take the same thing as the gateway.
	// It has to go before Environment is prepared, so we have resources registered in K8S schema
	return core_plugins.EnvironmentPreparingOrder - 1
}

func NewGenerator(zone string) Generator {
	return Generator{
		Zone: zone,
		FilterChainGenerators: FilterChainGenerators{
			FilterChainGenerators: map[mesh_proto.MeshGateway_Listener_Protocol]FilterChainGenerator{
				mesh_proto.MeshGateway_Listener_HTTP:  &HTTPFilterChainGenerator{},
				mesh_proto.MeshGateway_Listener_HTTPS: &HTTPSFilterChainGenerator{},
			},
		},
		ClusterGenerator: &ClusterGenerator{},
	}
}

// NewProxyProfile returns a new resource generator profile for
// global load balancer dataplanes.
func NewProxyProfile(zone string) generator_core.ResourceGenerator {
	return generator_core.CompositeResourceGenerator{
		generator.AdminProxyGenerator{},
		generator.PrometheusEndpointGenerator{},
		generator.TracingProxyGenerator{},
		generator.TransparentProxyGenerator{},
		generator.DNSGenerator{},
		NewGenerator(zone),
		generator_secrets.Generator{},
	}
}
