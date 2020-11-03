package k8s

import (
	"github.com/pkg/errors"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	k8s_runtime "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

var _ core_plugins.ResourceStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewResourceStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_store.ResourceStore, error) {
	mgr, ok := k8s_runtime.FromManagerContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	if err := mesh_k8s.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrap(err, "could not add to scheme")
	}
	converter, ok := k8s_runtime.FromResourceConverterContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s resource converter hasn't been configured")
	}
	return NewStore(mgr.GetClient(), mgr.GetScheme(), converter)
}

func (p *plugin) Migrate(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_plugins.DbVersion, error) {
	return 0, errors.New("migrations are not supported for Kubernetes resource store")
}
