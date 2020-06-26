package runtime

import (
	"context"

	"github.com/Kong/kuma/pkg/clusters/poller"

	"github.com/Kong/kuma/pkg/dns"

	"github.com/pkg/errors"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/core"
	core_ca "github.com/Kong/kuma/pkg/core/ca"
	config_manager "github.com/Kong/kuma/pkg/core/config/manager"
	"github.com/Kong/kuma/pkg/core/datasource"
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
	DNSResolver() dns.DNSResolver
	Clusters() poller.ClusterStatusPoller
	ConfigManager() config_manager.ConfigManager
	LeaderInfo() component.LeaderInfo
}

var _ BuilderContext = &Builder{}

// Builder represents a multi-step initialization process.
type Builder struct {
	cfg      kuma_cp.Config
	cm       component.Manager
	rs       core_store.ResourceStore
	rm       core_manager.ResourceManager
	rom      core_manager.ReadOnlyResourceManager
	sm       secret_manager.SecretManager
	cam      core_ca.Managers
	xds      core_xds.XdsContext
	dsl      datasource.Loader
	ext      context.Context
	dns      dns.DNSResolver
	clusters poller.ClusterStatusPoller
	configm  config_manager.ConfigManager
	leadInfo component.LeaderInfo
	*runtimeInfo
}

func BuilderFor(cfg kuma_cp.Config) *Builder {
	return &Builder{
		cfg: cfg,
		ext: context.Background(),
		cam: core_ca.Managers{},
		runtimeInfo: &runtimeInfo{
			instanceId: core.NewUUID(),
		},
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

func (b *Builder) WithCaManagers(cam core_ca.Managers) *Builder {
	b.cam = cam
	return b
}

func (b *Builder) WithCaManager(name string, cam core_ca.Manager) *Builder {
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

func (b *Builder) WithDNSResolver(dns dns.DNSResolver) *Builder {
	b.dns = dns
	return b
}

func (b *Builder) WithClusters(clusters poller.ClusterStatusPoller) *Builder {
	b.clusters = clusters
	return b
}

func (b *Builder) WithConfigManager(configm config_manager.ConfigManager) *Builder {
	b.configm = configm
	return b
}

func (b *Builder) WithLeaderInfo(leadInfo component.LeaderInfo) *Builder {
	b.leadInfo = leadInfo
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
	if b.dns == nil {
		return nil, errors.Errorf("DNS has been misconfigured")
	}
	if b.leadInfo == nil {
		return nil, errors.Errorf("LeaderInfo has not been configured")
	}
	return &runtime{
		RuntimeInfo: b.runtimeInfo,
		RuntimeContext: &runtimeContext{
			cfg:      b.cfg,
			rm:       b.rm,
			rom:      b.rom,
			rs:       b.rs,
			sm:       b.sm,
			cam:      b.cam,
			xds:      b.xds,
			ext:      b.ext,
			dns:      b.dns,
			clusters: b.clusters,
			configm:  b.configm,
			leadInfo: b.leadInfo,
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
func (b *Builder) ResourceManager() core_manager.ResourceManager {
	return b.rm
}
func (b *Builder) ReadOnlyResourceManager() core_manager.ReadOnlyResourceManager {
	return b.rom
}
func (b *Builder) CaManagers() core_ca.Managers {
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
func (b *Builder) DNSResolver() dns.DNSResolver {
	return b.dns
}
func (b *Builder) Clusters() poller.ClusterStatusPoller {
	return b.clusters
}
func (b *Builder) ConfigManager() config_manager.ConfigManager {
	return b.configm
}
func (b *Builder) LeaderInfo() component.LeaderInfo {
	return b.leadInfo
}
