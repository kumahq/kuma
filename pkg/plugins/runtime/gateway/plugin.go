package gateway

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator"
	generator_core "github.com/kumahq/kuma/pkg/xds/generator/core"
	generator_secrets "github.com/kumahq/kuma/pkg/xds/generator/secrets"
	"github.com/kumahq/kuma/pkg/xds/template"
)

func init() {
	core_plugins.Register(metadata.PluginName, &plugin{})
}

var (
	log = core.Log.WithName("plugin").WithName("runtime").WithName("gateway")
)

type plugin struct{}

var _ core_plugins.BootstrapPlugin = &plugin{}

func (p *plugin) BeforeBootstrap(context *core_plugins.MutablePluginContext, config core_plugins.PluginConfig) error {
	if context.Config().Environment == config_core.KubernetesEnvironment {
		mesh_k8s.RegisterK8sGatewayTypes()

		if context.Config().Experimental.GatewayAPI {
			mesh_k8s.RegisterK8sGatewayAPITypes()
		}
	}
	return nil
}

func (p *plugin) AfterBootstrap(context *core_plugins.MutablePluginContext, config core_plugins.PluginConfig) error {
	// Insert our resolver before the default so that we can intercept
	// builtin gateway dataplanes.
	generator.DefaultTemplateResolver = template.SequentialResolver(
		TemplateResolver{},
		generator.DefaultTemplateResolver,
	)

	generator.RegisterProfile(metadata.ProfileGatewayProxy, NewProxyProfile(context.Config().Multizone.Zone.Name))

	log.Info("registered gateway plugin")
	return nil
}

func (p *plugin) Name() core_plugins.PluginName {
	return metadata.PluginName
}

func (p *plugin) Order() int {
	// It has to go before Environment is prepared, so we have resources registered in K8S schema
	return core_plugins.EnvironmentPreparingOrder - 1
}

// NewProxyProfile returns a new resource generator profile for builtin
// gateway dataplanes.
func NewProxyProfile(zone string) generator_core.ResourceGenerator {
	return generator_core.CompositeResourceGenerator{
		generator.AdminProxyGenerator{},
		generator.PrometheusEndpointGenerator{},
		generator.TracingProxyGenerator{},
		generator.TransparentProxyGenerator{},
		generator.DNSGenerator{},
		Generator{
			FilterChainGenerators: FilterChainGenerators{
				FilterChainGenerators: map[mesh_proto.MeshGateway_Listener_Protocol]FilterChainGenerator{
					mesh_proto.MeshGateway_Listener_HTTP:  &HTTPFilterChainGenerator{},
					mesh_proto.MeshGateway_Listener_HTTPS: &HTTPSFilterChainGenerator{},
					mesh_proto.MeshGateway_Listener_TCP:   &TCPFilterChainGenerator{},
				}},
			ClusterGenerator: ClusterGenerator{
				Zone: zone,
			},
			Zone: zone,
		},
		generator_secrets.Generator{},
	}
}
