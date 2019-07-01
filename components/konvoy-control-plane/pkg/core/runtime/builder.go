package runtime

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	"github.com/pkg/errors"
)

// BuilderContext provides access to Builder's interim state.
type BuilderContext interface {
	ComponentManager() ComponentManager
	ResourceStore() core_store.ResourceStore
	XdsContext() core_xds.XdsContext
}

var _ BuilderContext = &Builder{}

// Builder represents a multi-step initialization process.
type Builder struct {
	cfg config.Config
	cm  ComponentManager
	rs  core_store.ResourceStore
	dss []core_discovery.DiscoverySource
	xds core_xds.XdsContext
}

func BuilderFor(cfg config.Config) *Builder {
	return &Builder{cfg: cfg}
}

func (b *Builder) WithComponentManager(cm ComponentManager) *Builder {
	b.cm = cm
	return b
}

func (b *Builder) WithResourceStore(rs core_store.ResourceStore) *Builder {
	b.rs = rs
	return b
}

func (b *Builder) AddDiscoverySource(ds core_discovery.DiscoverySource) *Builder {
	b.dss = append(b.dss, ds)
	return b
}

func (b *Builder) WithXdsContext(xds core_xds.XdsContext) *Builder {
	b.xds = xds
	return b
}

func (b *Builder) Build() (Runtime, error) {
	if b.cm == nil {
		return nil, errors.Errorf("ComponentManager has not been configured")
	}
	if b.rs == nil {
		return nil, errors.Errorf("ResourceStore has not been configured")
	}
	if len(b.dss) == 0 {
		return nil, errors.Errorf("DiscoverySources have not been configured")
	}
	if b.xds == nil {
		return nil, errors.Errorf("xDS Context has not been configured")
	}
	return &runtime{
		RuntimeContext: &runtimeContext{
			cfg: b.cfg,
			rs:  b.rs,
			dss: b.dss,
			xds: b.xds,
		},
		ComponentManager: b.cm,
	}, nil
}

func (b *Builder) ComponentManager() ComponentManager {
	return b.cm
}

func (b *Builder) ResourceStore() core_store.ResourceStore {
	return b.rs
}
func (b *Builder) XdsContext() core_xds.XdsContext {
	return b.xds
}
