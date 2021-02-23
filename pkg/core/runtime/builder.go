package runtime

import (
	"context"
	"fmt"
	"os"

	"github.com/kumahq/kuma/pkg/envoy/admin"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"

	api_server "github.com/kumahq/kuma/pkg/api-server/customization"
	dp_server "github.com/kumahq/kuma/pkg/dp-server/server"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/events"

	"github.com/kumahq/kuma/pkg/dns/resolver"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/metrics"
)

// BuilderContext provides access to Builder's interim state.
type BuilderContext interface {
	ComponentManager() component.Manager
	ResourceStore() core_store.ResourceStore
	SecretStore() store.SecretStore
	ConfigStore() core_store.ResourceStore
	ResourceManager() core_manager.ResourceManager
	Config() kuma_cp.Config
	DataSourceLoader() datasource.Loader
	Extensions() context.Context
	DNSResolver() resolver.DNSResolver
	ConfigManager() config_manager.ConfigManager
	LeaderInfo() component.LeaderInfo
	Metrics() metrics.Metrics
	EventReaderFactory() events.ListenerFactory
	APIManager() api_server.APIManager
	XDSHooks() *xds_hooks.Hooks
	DpServer() *dp_server.DpServer
	KDSContext() *kds_context.Context
}

var _ BuilderContext = &Builder{}

// Builder represents a multi-step initialization process.
type Builder struct {
	cfg      kuma_cp.Config
	cm       component.Manager
	rs       core_store.ResourceStore
	ss       store.SecretStore
	cs       core_store.ResourceStore
	rm       core_manager.ResourceManager
	rom      core_manager.ReadOnlyResourceManager
	cam      core_ca.Managers
	dsl      datasource.Loader
	ext      context.Context
	dns      resolver.DNSResolver
	configm  config_manager.ConfigManager
	leadInfo component.LeaderInfo
	lif      lookup.LookupIPFunc
	eac      admin.EnvoyAdminClient
	metrics  metrics.Metrics
	erf      events.ListenerFactory
	apim     api_server.APIManager
	xdsh     *xds_hooks.Hooks
	dps      *dp_server.DpServer
	kdsctx   *kds_context.Context
	closeCh  <-chan struct{}
	*runtimeInfo
}

func BuilderFor(cfg kuma_cp.Config, closeCh <-chan struct{}) (*Builder, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "could not get hostname")
	}
	suffix := core.NewUUID()[0:4]
	return &Builder{
		cfg: cfg,
		ext: context.Background(),
		cam: core_ca.Managers{},
		runtimeInfo: &runtimeInfo{
			instanceId: fmt.Sprintf("%s-%s", hostname, suffix),
		},
		closeCh: closeCh,
	}, nil
}

func (b *Builder) WithComponentManager(cm component.Manager) *Builder {
	b.cm = cm
	return b
}

func (b *Builder) WithResourceStore(rs core_store.ResourceStore) *Builder {
	b.rs = rs
	return b
}

func (b *Builder) WithSecretStore(ss store.SecretStore) *Builder {
	b.ss = ss
	return b
}

func (b *Builder) WithConfigStore(cs core_store.ResourceStore) *Builder {
	b.cs = cs
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

func (b *Builder) WithExtensions(ext context.Context) *Builder {
	b.ext = ext
	return b
}

func (b *Builder) WithExtension(key interface{}, value interface{}) *Builder {
	b.ext = context.WithValue(b.ext, key, value)
	return b
}

func (b *Builder) WithDNSResolver(dns resolver.DNSResolver) *Builder {
	b.dns = dns
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

func (b *Builder) WithLookupIP(lif lookup.LookupIPFunc) *Builder {
	b.lif = lif
	return b
}

func (b *Builder) WithEnvoyAdminClient(eac admin.EnvoyAdminClient) *Builder {
	b.eac = eac
	return b
}

func (b *Builder) WithMetrics(metrics metrics.Metrics) *Builder {
	b.metrics = metrics
	return b
}

func (b *Builder) WithEventReaderFactory(erf events.ListenerFactory) *Builder {
	b.erf = erf
	return b
}

func (b *Builder) WithAPIManager(apim api_server.APIManager) *Builder {
	b.apim = apim
	return b
}

func (b *Builder) WithXDSHooks(xdsh *xds_hooks.Hooks) *Builder {
	b.xdsh = xdsh
	return b
}

func (b *Builder) WithDpServer(dps *dp_server.DpServer) *Builder {
	b.dps = dps
	return b
}

func (b *Builder) WithKDSContext(kdsctx *kds_context.Context) *Builder {
	b.kdsctx = kdsctx
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
	if b.lif == nil {
		return nil, errors.Errorf("LookupIP func has not been configured")
	}
	if b.eac == nil {
		return nil, errors.Errorf("EnvoyAdminClient has not been configured")
	}
	if b.metrics == nil {
		return nil, errors.Errorf("Metrics has not been configured")
	}
	if b.erf == nil {
		return nil, errors.Errorf("EventReaderFactory has not been configured")
	}
	if b.apim == nil {
		return nil, errors.Errorf("APIManager has not been configured")
	}
	if b.xdsh == nil {
		return nil, errors.Errorf("XDSHooks has not been configured")
	}
	if b.dps == nil {
		return nil, errors.Errorf("DpServer has not been configured")
	}
	if b.kdsctx == nil {
		return nil, errors.Errorf("KDSContext has not been configured")
	}
	return &runtime{
		RuntimeInfo: b.runtimeInfo,
		RuntimeContext: &runtimeContext{
			cfg:      b.cfg,
			rm:       b.rm,
			rom:      b.rom,
			rs:       b.rs,
			ss:       b.ss,
			cam:      b.cam,
			dsl:      b.dsl,
			ext:      b.ext,
			dns:      b.dns,
			configm:  b.configm,
			leadInfo: b.leadInfo,
			lif:      b.lif,
			eac:      b.eac,
			metrics:  b.metrics,
			erf:      b.erf,
			apim:     b.apim,
			xdsh:     b.xdsh,
			dps:      b.dps,
			kdsctx:   b.kdsctx,
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
func (b *Builder) SecretStore() store.SecretStore {
	return b.ss
}
func (b *Builder) ConfigStore() core_store.ResourceStore {
	return b.cs
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
func (b *Builder) Config() kuma_cp.Config {
	return b.cfg
}
func (b *Builder) DataSourceLoader() datasource.Loader {
	return b.dsl
}
func (b *Builder) Extensions() context.Context {
	return b.ext
}
func (b *Builder) DNSResolver() resolver.DNSResolver {
	return b.dns
}
func (b *Builder) ConfigManager() config_manager.ConfigManager {
	return b.configm
}
func (b *Builder) LeaderInfo() component.LeaderInfo {
	return b.leadInfo
}
func (b *Builder) LookupIP() lookup.LookupIPFunc {
	return b.lif
}
func (b *Builder) Metrics() metrics.Metrics {
	return b.metrics
}
func (b *Builder) EventReaderFactory() events.ListenerFactory {
	return b.erf
}
func (b *Builder) APIManager() api_server.APIManager {
	return b.apim
}
func (b *Builder) XDSHooks() *xds_hooks.Hooks {
	return b.xdsh
}
func (b *Builder) DpServer() *dp_server.DpServer {
	return b.dps
}
func (b *Builder) KDSContext() *kds_context.Context {
	return b.kdsctx
}
func (b *Builder) CloseCh() <-chan struct{} {
	return b.closeCh
}
