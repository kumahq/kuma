package runtime

import (
	"context"

	"github.com/Kong/kuma/pkg/core/datasource"

	"github.com/pkg/errors"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/core"
	core_ca "github.com/Kong/kuma/pkg/core/ca"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

// BuilderContext provides access to Builder's interim state.
type BuilderContext interface {
	ComponentManager() component.Manager
	ResourceStore() core_store.ResourceStore
	XdsContext() core_xds.XdsContext
	Config() kuma_cp.Config
	SecretManager() secret_manager.SecretManager
	DataSourceLoader() datasource.Loader
	Extensions() context.Context
}

var _ BuilderContext = &Builder{}

// Builder represents a multi-step initialization process.
type Builder struct {
	cfg kuma_cp.Config
	cm  component.Manager
	rs  core_store.ResourceStore
	rm  core_manager.ResourceManager
	rom core_manager.ReadOnlyResourceManager
	sm  secret_manager.SecretManager
	cam core_ca.CaManagers
	xds core_xds.XdsContext
	dsl datasource.Loader
	ext context.Context
}

func BuilderFor(cfg kuma_cp.Config) *Builder {
	return &Builder{
		cfg: cfg,
		ext: context.Background(),
		cam: core_ca.CaManagers{},
	}
}

func (b *Builder) WithComponentManager(cm component.Manager) *Builder {
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

func (b *Builder) WithReadOnlyResourceManager(rom core_manager.ReadOnlyResourceManager) *Builder {
	b.rom = rom
	return b
}

func (b *Builder) WithSecretManager(sm secret_manager.SecretManager) *Builder {
	b.sm = sm
	return b
}

func (b *Builder) WithCaManagers(cam core_ca.CaManagers) *Builder {
	b.cam = cam
	return b
}

func (b *Builder) WithCaManager(name string, cam core_ca.CaManager) *Builder {
	b.cam[name] = cam
	return b
}

func (b *Builder) WithDataSourceLoader(loader datasource.Loader) *Builder {
	b.dsl = loader
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
	if b.rom == nil {
		return nil, errors.Errorf("ReadOnlyResourceManager has not been configured")
	}
	if b.sm == nil {
		return nil, errors.Errorf("SecretManager has not been configured")
	}
	if b.xds == nil {
		return nil, errors.Errorf("xDS Context has not been configured")
	}
	if b.dsl == nil {
		return nil, errors.Errorf("DataSourceLoader has not been configured")
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
			rom: b.rom,
			sm:  b.sm,
			cam: b.cam,
			xds: b.xds,
			ext: b.ext,
		},
		Manager: b.cm,
	}, nil
}

func (b *Builder) ComponentManager() component.Manager {
	return b.cm
}
func (b *Builder) ResourceStore() core_store.ResourceStore {
	return b.rs
}
func (b *Builder) SecretManager() secret_manager.SecretManager {
	return b.sm
}
func (b *Builder) ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager {
	return b.rom
}
func (b *Builder) CaManagers() core_ca.CaManagers {
	return b.cam
}
func (b *Builder) XdsContext() core_xds.XdsContext {
	return b.xds
}
func (b *Builder) Config() kuma_cp.Config {
	return b.cfg
}
func (b *Builder) DataSourceLoader() datasource.Loader {
	return b.dsl
}
func (b *Builder) Extensions() context.Context {
	return b.ext
}
