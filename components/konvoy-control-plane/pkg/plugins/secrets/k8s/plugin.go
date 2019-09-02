package k8s

import (
	"reflect"

	"github.com/pkg/errors"

	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"
	secret_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/store"

	kube_core "k8s.io/api/core/v1"
	kube_ctrl "sigs.k8s.io/controller-runtime"
)

var _ core_plugins.SecretStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewSecretStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (secret_store.SecretStore, error) {
	mgr, ok := pc.ComponentManager().(kube_ctrl.Manager)
	if !ok {
		return nil, errors.Errorf("Component Manager has a wrong type: expected=%q got=%q", reflect.TypeOf(kube_ctrl.Manager(nil)), reflect.TypeOf(pc.ComponentManager()))
	}
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	return NewStore(mgr.GetClient())
}
