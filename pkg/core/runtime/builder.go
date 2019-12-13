package runtime

import (
	"context"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/core"
	builtin_ca "github.com/Kong/kuma/pkg/core/ca/builtin"
	provided_ca "github.com/Kong/kuma/pkg/core/ca/provided"
	core_discovery "github.com/Kong/kuma/pkg/core/discovery"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	"github.com/pkg/errors"
)

// BuilderContext provides access to Builder's interim state.
type BuilderContext interface {
	ComponentManager() ComponentManager
	ResourceStore() core_store.ResourceStore
	XdsContext() core_xds.XdsContext
	Config() kuma_cp.Config
	Extensions() context.Context
}

var _ BuilderContext = &Builder{}

// Builder represents a multi-step initialization process.
type Builder struct {
	cfg kuma_cp.Config
	cm  ComponentManager
	rs  core_store.ResourceStore
	rm  core_manager.ResourceManager
	sm  secret_manager.SecretManager
	bcm builtin_ca.BuiltinCaManager
	pcm provided_ca.ProvidedCaManager
	dss []core_discovery.DiscoverySource
	xds core_xds.XdsContext
	ext context.Context
}

func BuilderFor(cfg kuma_cp.Config) *Builder {
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

func (b *Builder) WithBuiltinCaManager(bcm builtin_ca.BuiltinCaManager) *Builder {
	b.bcm = bcm
	return b
}

func (b *Builder) WithProvidedCaManager(pcm provided_ca.ProvidedCaManager) *Builder {
	b.pcm = pcm
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
	if b.bcm == nil {
		return nil, errors.Errorf("BuiltinCaManager has not been configured")
	}
	if b.pcm == nil {
		return nil, errors.Errorf("ProvidedCaManager has not been configured")
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
			bcm: b.bcm,
			pcm: b.pcm,
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
func (b *Builder) SecretManager() secret_manager.SecretManager {
	return b.sm
}
func (b *Builder) BuiltinCaManager() builtin_ca.BuiltinCaManager {
	return b.bcm
}
func (b *Builder) ProvidedCaManager() provided_ca.ProvidedCaManager {
	return b.pcm
}
func (b *Builder) XdsContext() core_xds.XdsContext {
	return b.xds
}
func (b *Builder) Config() kuma_cp.Config {
	return b.cfg
}
func (b *Builder) Extensions() context.Context {
	return b.ext
}
