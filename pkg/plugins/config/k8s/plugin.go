package k8s

import (
	"github.com/pkg/errors"

	core_store "github.com/kumahq/kuma/pkg/core/resources/store"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"

	kube_core "k8s.io/api/core/v1"
)

var _ core_plugins.ConfigStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewConfigStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_store.ResourceStore, error) {
	mgr, ok := k8s_extensions.FromManagerContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	if err := kube_core.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", kube_core.SchemeGroupVersion)
	}
	converter, ok := k8s_extensions.FromResourceConverterContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s resource converter hasn't been configured")
	}
	return NewStore(mgr.GetClient(), pc.Config().Store.Kubernetes.SystemNamespace, mgr.GetScheme(), converter)
}
