package k8s

import (
	"github.com/pkg/errors"

	config_store "github.com/Kong/kuma/pkg/core/config/store"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	k8s_runtime "github.com/Kong/kuma/pkg/runtime/k8s"

	kube_core "k8s.io/api/core/v1"
)

var _ core_plugins.ConfigStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewConfigStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (config_store.ConfigStore, error) {
	mgr, ok := k8s_runtime.FromManagerContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	return NewStore(mgr.GetClient(), pc.Config().Store.Kubernetes.SystemNamespace)
}
