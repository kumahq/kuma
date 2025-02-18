package k8s

import (
	"github.com/pkg/errors"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	k8s_runtime "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	k8s_events "github.com/kumahq/kuma/pkg/plugins/resources/k8s/events"
)

var _ core_plugins.ResourceStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewResourceStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_store.ResourceStore, core_store.Transactions, error) {
	mgr, ok := k8s_runtime.FromManagerContext(pc.Extensions())
	if !ok {
		return nil, nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	converter, ok := k8s_runtime.FromResourceConverterContext(pc.Extensions())
	if !ok {
		return nil, nil, errors.Errorf("k8s resource converter hasn't been configured")
	}
	store, err := NewStore(mgr.GetClient(), mgr.GetScheme(), converter)
	return store, core_store.NoTransactions{}, err
}

func (p *plugin) Migrate(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_plugins.DbVersion, error) {
	return 0, errors.New("migrations are not supported for Kubernetes resource store")
}

func (p *plugin) EventListener(pc core_plugins.PluginContext, writer events.Emitter) error {
	mgr, ok := k8s_runtime.FromManagerContext(pc.Extensions())
	if !ok {
		return errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	if err := pc.ComponentManager().Add(k8s_events.NewListener(mgr, writer, pc.Config().Runtime.Kubernetes.WatchNamespaces, pc.Config().Store.Kubernetes.SystemNamespace)); err != nil {
		return err
	}
	return nil
}
