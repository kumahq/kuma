package runtime

import (
	"context"

	konvoy_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-cp"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/manager"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	secret_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/manager"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	"github.com/pkg/errors"
)

// BuilderContext provides access to Builder's interim state.
type BuilderContext interface {
	ComponentManager() ComponentManager
	ResourceStore() core_store.ResourceStore
	XdsContext() core_xds.XdsContext
	Config() konvoy_cp.Config
	Extensions() context.Context
}

var _ BuilderContext = &Builder{}

// Builder represents a multi-step initialization process.
type Builder struct {
	cfg konvoy_cp.Config
	cm  ComponentManager
	rs  core_store.ResourceStore
	rm  core_manager.ResourceManager
	sm  secret_manager.SecretManager
	dss []core_discovery.DiscoverySource
	xds core_xds.XdsContext
	ext context.Context
}

func BuilderFor(cfg konvoy_cp.Config) *Builder {
	return &Builder{cfg: cfg, ext: context.Background()}
}

func (b *Builder) WithComponentManager(cm ComponentManager) *Builder {
	b.cm = cm
	return b
}

func (b *Builder) WithResourceStore(rs core_store.ResourceStore) *Builder {
	b.rs = rs
	return b
}

func (b *Builder) WithResourceManager(rm core_manager.ResourceManager) *Builder {
	b.rm = rm
	return b
}

func (b *Builder) WithSecretManager(sm secret_manager.SecretManager) *Builder {
	b.sm = sm
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

func (b *Builder) WithExtensions(ext context.Context) *Builder {
	b.ext = ext
	return b
}

func (b *Builder) Build() (Runtime, error) {
	if b.cm == nil {
		return nil, errors.Errorf("ComponentManager has not been configured")
	}
	if b.rs == nil {
		return nil, errors.Errorf("ResourceStore has not been configured")
	}
	if b.rm == nil {
		return nil, errors.Errorf("ResourceManager has not been configured")
	}
	if b.sm == nil {
		return nil, errors.Errorf("SecretManager has not been configured")
	}
	// todo(jakubdyszkiewicz) restore when we've got store based discovery source
	//if len(b.dss) == 0 {
	//	return nil, errors.Errorf("DiscoverySources have not been configured")
	//}
	if b.xds == nil {
		return nil, errors.Errorf("xDS Context has not been configured")
	}
	if b.ext == nil {
		return nil, errors.Errorf("Extensions have been misconfigured")
	}
	return &runtime{
		RuntimeInfo: &runtimeInfo{
			instanceId: core.NewUUID(),
		},
		RuntimeContext: &runtimeContext{
			cfg: b.cfg,
			rm:  b.rm,
			sm:  b.sm,
			dss: b.dss,
			xds: b.xds,
			ext: b.ext,
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
func (b *Builder) Config() konvoy_cp.Config {
	return b.cfg
}
func (b *Builder) Extensions() context.Context {
	return b.ext
}
