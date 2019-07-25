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

func initializeDiscovery(cfg konvoy_cp.Config, builder *core_runtime.Builder) error {
	if cfg.Environment == konvoy_cp.KubernetesEnvironmentType {
		plugin, err := core_plugins.Plugins().Discovery(core_plugins.Kubernetes)
		if err != nil {
			return nil
		}
		if source, err := plugin.NewDiscoverySource(builder, nil); err != nil {
			return nil
		} else {
			builder.AddDiscoverySource(source)
		}
	}
	return nil
}

func initializeBootstrap(cfg konvoy_cp.Config, builder *core_runtime.Builder) error {
	var plugin core_plugins.BootstrapPlugin
	switch cfg.Environment {
	case konvoy_cp.KubernetesEnvironmentType:
		bp, err := core_plugins.Plugins().Bootstrap(core_plugins.Kubernetes)
		if err != nil {
			return errors.Wrap(err, "could not retrieve boostrap Kubernetes plugin")
		}
		plugin = bp
	case konvoy_cp.StandaloneEnvironmentType:
		bp, err := core_plugins.Plugins().Bootstrap(core_plugins.Standalone)
		if err != nil {
			return errors.Wrap(err, "could not retrieve boostrap Standalone plugin")
		}
		plugin = bp
	default:
		return errors.Errorf("unknown environment type %s", cfg.Environment)
	}
	if err := plugin.Bootstrap(builder, nil); err != nil {
		return err
	}
	return nil
}

func initializeResourceStore(cfg konvoy_cp.Config, builder *core_runtime.Builder) error {
	var plugin core_plugins.ResourceStorePlugin
	switch cfg.Store.Type {
	case store.KubernetesStoreType:
		rsp, err := core_plugins.Plugins().ResourceStore(core_plugins.Kubernetes)
		if err != nil {
			return errors.Wrap(err, "could not retrieve store Kubernetes plugin")
		}
		plugin = rsp
	case store.MemoryStoreType:
		rsp, err := core_plugins.Plugins().ResourceStore(core_plugins.Memory)
		if err != nil {
			return errors.Wrap(err, "could not retrieve store Memory plugin")
		}
		plugin = rsp
	case store.PostgresStoreType:
		rsp, err := core_plugins.Plugins().ResourceStore(core_plugins.Postgres)
		if err != nil {
			return errors.Wrap(err, "could not retrieve store Postgres plugin")
		}
		plugin = rsp
	default:
		return errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	if rs, err := plugin.NewResourceStore(builder, nil); err != nil {
		return err
	} else {
		builder.WithResourceStore(rs)
		return nil
	}
}

func initializeXds(builder *core_runtime.Builder) {
	builder.WithXdsContext(core_xds.NewXdsContext())
}
