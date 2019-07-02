package k8s

import (
	"github.com/pkg/errors"
	"reflect"

	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"

	kube_ctrl "sigs.k8s.io/controller-runtime"
)

var _ core_plugins.ResourceStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewResourceStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_store.ResourceStore, error) {
	mgr, ok := pc.ComponentManager().(kube_ctrl.Manager)
	if !ok {
		return nil, errors.Errorf("Component Manager has a wrong type: expected=%q got=%q", reflect.TypeOf(kube_ctrl.Manager(nil)), reflect.TypeOf(pc.ComponentManager()))
	}
	mesh_k8s.AddToScheme(mgr.GetScheme())
	return NewStore(mgr.GetClient())
}
