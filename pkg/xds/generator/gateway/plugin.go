package gateway

import (
	"context"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_runtime "github.com/kumahq/kuma/v3/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	mesh_k8s "github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	xds_generator "github.com/kumahq/kuma/v3/pkg/xds/generator"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/template"
)

var (
	_ core_plugins.BootstrapPlugin = &plugin{}
	_ core_plugins.ProxyPlugin     = &plugin{}
)

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.PluginName(metadata.PluginName), &plugin{})
}

func (p *plugin) BeforeBootstrap(b *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	if b.Config().Environment == config_core.KubernetesEnvironment {
		mesh_k8s.RegisterK8sGatewayTypes()
		mesh_k8s.RegisterK8sGatewayAPITypes()
	}
	return nil
}

func (p *plugin) AfterBootstrap(_ *core_runtime.Builder, _ core_plugins.PluginConfig) error {
	xds_generator.DefaultTemplateResolver = template.SequentialResolver(
		TemplateResolver{},
		xds_generator.DefaultTemplateResolver,
	)
	log.Info("registered gateway plugin")
	return nil
}

func (p *plugin) Apply(ctx context.Context, meshContext xds_context.MeshContext, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil || !proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}

	SetGatewayListeners(proxy, GatewayListenerInfoFromProxy(ctx, meshContext, proxy))
	return nil
}

func (p *plugin) Name() core_plugins.PluginName {
	return core_plugins.PluginName(metadata.PluginName)
}

func (p *plugin) Order() int {
	return core_plugins.EnvironmentPreparingOrder - 1
}
