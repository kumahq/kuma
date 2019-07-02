package bootstrap

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
)

func Bootstrap(cfg config.Config) (core_runtime.Runtime, error) {
	// 1. init Runtime Builder with passed-in configuration
	builder := core_runtime.BuilderFor(cfg)

	// 2. initialize environment-specific parts, e.g. ComponentManager

	// delegate this task to a "bootstrap" plugin (for now, always k8s)
	bp, err := core_plugins.Plugins().Bootstrap(core_plugins.Kubernetes)
	if err != nil {
		return nil, err
	}
	if err := bp.Bootstrap(builder, nil); err != nil {
		return nil, err
	}

	// 3. initialize ResourceStore

	// delegate this task to a "resource-store" plugin (for now, always k8s)
	rsp, err := core_plugins.Plugins().ResourceStore(core_plugins.Kubernetes)
	if err != nil {
		return nil, err
	}
	if rs, err := rsp.NewResourceStore(builder, nil); err != nil {
		return nil, err
	} else {
		builder.WithResourceStore(rs)
	}

	// 4. initialize Service Discovery

	// delegate this task to a "discovery" plugin (for now, always k8s)
	dp, err := core_plugins.Plugins().Discovery(core_plugins.Kubernetes)
	if err != nil {
		return nil, err
	}
	if ds, err := dp.NewDiscoverySource(builder, nil); err != nil {
		return nil, err
	} else {
		builder.AddDiscoverySource(ds)
	}

	// 5. initialize xDS

	builder.WithXdsContext(core_xds.NewXdsContext())

	// finally, build Runtime
	return builder.Build()
}
