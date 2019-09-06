package k8s

import (
	"github.com/pkg/errors"

	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"
	secret_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/store"
	k8s_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/runtime/k8s"

	kube_core "k8s.io/api/core/v1"
)

var _ core_plugins.SecretStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewSecretStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (secret_store.SecretStore, error) {
	mgr, ok := k8s_runtime.FromManagerContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	return NewStore(mgr.GetClient(), pc.Config().Store.Kubernetes.SystemNamespace)
}
