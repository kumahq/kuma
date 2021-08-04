package k8s

import (
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
)

var _ core_plugins.SecretStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewSecretStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (secret_store.SecretStore, error) {
	mgr, ok := k8s_extensions.FromManagerContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	client, ok := k8s_extensions.FromSecretClientContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("secret client hasn't been configured")
	}
	return NewStore(client, client, pc.Config().Store.Kubernetes.SystemNamespace)
}
