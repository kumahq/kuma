package k8s

import (
	"github.com/pkg/errors"
	"reflect"

	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"

	kube_ctrl "sigs.k8s.io/controller-runtime"
)

var _ core_plugins.DiscoveryPlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewDiscoverySource(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_discovery.DiscoverySource, error) {
	mgr, ok := pc.ComponentManager().(kube_ctrl.Manager)
	if !ok {
		return nil, errors.Errorf("Component Manager has a wrong type: expected=%q got=%q", reflect.TypeOf(kube_ctrl.Manager(nil)), reflect.TypeOf(pc.ComponentManager()))
	}
	return NewDiscoverySource(mgr)
}
