package k8s

import (
	"github.com/pkg/errors"

	core_discovery "github.com/Kong/kuma/pkg/core/discovery"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"
)

var _ core_plugins.DiscoveryPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewDiscoverySource(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_discovery.DiscoverySource, error) {
	mgr, ok := k8s_runtime.FromManagerContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	return NewDiscoverySource(mgr, pc.Config().Store.Kubernetes.SystemNamespace)
}
