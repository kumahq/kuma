package memory

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/events"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

var _ core_plugins.ResourceStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Memory, &plugin{})
}

func (p *plugin) NewResourceStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_store.ResourceStore, error) {
	return NewStore(), nil
}

func (p *plugin) Migrate(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_plugins.DbVersion, error) {
	return 0, errors.New("migrations are not supported for Memory resource store")
}

func (p *plugin) EventListener(context core_plugins.PluginContext, writer events.Emitter) error {
	context.ResourceStore().(*memoryStore).SetEventWriter(writer)
	return nil
}
