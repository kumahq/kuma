package bootstrap

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-cp"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/core/resources/store"
	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	"github.com/pkg/errors"
)

func Bootstrap(cfg konvoy_cp.Config) (core_runtime.Runtime, error) {
	builder := core_runtime.BuilderFor(cfg)
	if err := initializeBootstrap(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeResourceStore(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeDiscovery(cfg, builder); err != nil {
		return nil, err
	}
	initializeXds(builder)
	return builder.Build()
}

func initializeBootstrap(cfg konvoy_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	switch cfg.Environment {
	case konvoy_cp.KubernetesEnvironment:
		pluginName = core_plugins.Kubernetes
	case konvoy_cp.UniversalEnvironment:
		pluginName = core_plugins.Universal
	default:
		return errors.Errorf("unknown environment type %s", cfg.Environment)
	}
	plugin, err := core_plugins.Plugins().Bootstrap(pluginName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve bootstrap %s plugin", pluginName)
	}
	if err := plugin.Bootstrap(builder, nil); err != nil {
		return err
	}
	return nil
}

func initializeResourceStore(cfg konvoy_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	var pluginConfig core_plugins.PluginConfig
	switch cfg.Store.Type {
	case store.KubernetesStore:
		pluginName = core_plugins.Kubernetes
		pluginConfig = nil
	case store.MemoryStore:
		pluginName = core_plugins.Memory
		pluginConfig = nil
	case store.PostgresStore:
		pluginName = core_plugins.Postgres
		pluginConfig = cfg.Store.Postgres
	default:
		return errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	plugin, err := core_plugins.Plugins().ResourceStore(pluginName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve store %s plugin", pluginName)
	}
	if rs, err := plugin.NewResourceStore(builder, pluginConfig); err != nil {
		return err
	} else {
		builder.WithResourceStore(rs)
		return nil
	}
}

func initializeDiscovery(cfg konvoy_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	switch cfg.Environment {
	case konvoy_cp.KubernetesEnvironment:
		pluginName = core_plugins.Kubernetes
	case konvoy_cp.UniversalEnvironment:
		pluginName = core_plugins.Universal
	default:
		return errors.Errorf("unknown environment type %s", cfg.Environment)
	}
	plugin, err := core_plugins.Plugins().Discovery(pluginName)
	if err != nil {
		return nil
	}
	if source, err := plugin.NewDiscoverySource(builder, nil); err != nil {
		return nil
	} else {
		builder.AddDiscoverySource(source)
	}
	return nil
}

func initializeXds(builder *core_runtime.Builder) {
	builder.WithXdsContext(core_xds.NewXdsContext())
}
