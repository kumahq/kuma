package memory

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
)

var (
	log                                  = core.Log.WithName("plugins").WithName("resources").WithName("memory")
	_   core_plugins.ResourceStorePlugin = &plugin{}
)

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Memory, &plugin{})
}

func (p *plugin) NewResourceStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_store.ResourceStore, core_store.Transactions, error) {
	log.Info("kuma-cp runs with an in-memory database and its state isn't preserved between restarts. Keep in mind that an in-memory database cannot be used with multiple instances of the control plane.")
	return NewStore(), core_store.NoTransactions{}, nil
}

func (p *plugin) Migrate(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_plugins.DbVersion, error) {
	return 0, errors.New("migrations are not supported for Memory resource store")
}

func (p *plugin) EventListener(context core_plugins.PluginContext, writer events.Emitter) error {
	context.ResourceStore().DefaultResourceStore().(*memoryStore).SetEventWriter(writer)
	return nil
}
