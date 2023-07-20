package etcd

import (
	"github.com/kumahq/kuma/pkg/config/plugins/resources/etcd"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/plugins/resources/etcd/event"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
)

var (
	log                                  = core.Log.WithName("plugins").WithName("resources").WithName("etcd")
	_   core_plugins.ResourceStorePlugin = &plugin{}
)

const etcdKeyPrefix = "kuma"

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Etcd, &plugin{})
}

func (p *plugin) NewResourceStore(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_store.ResourceStore, error) {
	log.Info("kuma-cp runs with an in-memory database and its state isn't preserved between restarts. Keep in mind that an in-memory database cannot be used with multiple instances of the control plane.")
	cfg, ok := config.(*etcd.EtcdConfig)
	if !ok {
		return nil, errors.New("invalid type of the config. Passed config should be a PostgresStoreConfig")
	}
	return newEtcdStore(etcdKeyPrefix, pc.Metrics(), cfg)
}

func (p *plugin) Migrate(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_plugins.DbVersion, error) {
	return 0, errors.New("migrations are not supported for Memory resource store")
}

func (p *plugin) EventListener(pc core_plugins.PluginContext, writer events.Emitter) error {
	etcdListener := event.NewListener(etcdKeyPrefix, *pc.ResourceStore().(*EtcdStore).client, writer)
	return pc.ComponentManager().Add(component.NewResilientComponent(core.Log.WithName("etcd-event-listener-component"), etcdListener))
}
