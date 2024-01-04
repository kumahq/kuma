package runtime

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	api_server "github.com/kumahq/kuma/pkg/api-server/customization"
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
	dp_server "github.com/kumahq/kuma/pkg/dp-server/server"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/insights/globalinsight"
	"github.com/kumahq/kuma/pkg/intercp/client"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/plugins/resources/postgres/config"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_runtime "github.com/kumahq/kuma/pkg/xds/runtime"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

// BuilderContext provides access to Builder's interim state.
type BuilderContext interface {
	ComponentManager() component.Manager
	ResourceStore() core_store.CustomizableResourceStore
	Transactions() core_store.Transactions
	SecretStore() store.SecretStore
	ConfigStore() core_store.ResourceStore
	ResourceManager() core_manager.CustomizableResourceManager
	Config() kuma_cp.Config
	DataSourceLoader() datasource.Loader
	Extensions() context.Context
	ConfigManager() config_manager.ConfigManager
	LeaderInfo() component.LeaderInfo
	Metrics() metrics.Metrics
	EventBus() events.EventBus
	APIManager() api_server.APIManager
	CAProvider() secrets.CaProvider
	DpServer() *dp_server.DpServer
	ResourceValidators() ResourceValidators
	KDSContext() *kds_context.Context
	APIServerAuthenticator() authn.Authenticator
	Access() Access
	TokenIssuers() builtin.TokenIssuers
	MeshCache() *mesh.Cache
	InterCPClientPool() *client.Pool
	PgxConfigCustomizationFn() config.PgxConfigCustomization
	Tenants() multitenant.Tenants
}

var _ BuilderContext = &Builder{}

// Builder represents a multi-step initialization process.
type Builder struct {
	cfg            kuma_cp.Config
	cm             component.Manager
	rs             core_store.CustomizableResourceStore
	ss             store.SecretStore
	cs             core_store.ResourceStore
	txs            core_store.Transactions
	rm             core_manager.CustomizableResourceManager
	rom            core_manager.ReadOnlyResourceManager
	gis            globalinsight.GlobalInsightService
	cam            core_ca.Managers
	dsl            datasource.Loader
	ext            context.Context
	configm        config_manager.ConfigManager
	leadInfo       component.LeaderInfo
	lif            lookup.LookupIPFunc
	eac            admin.EnvoyAdminClient
	metrics        metrics.Metrics
	erf            events.EventBus
	apim           api_server.APIManager
	xds            xds_runtime.XDSRuntimeContext
	cap            secrets.CaProvider
	dps            *dp_server.DpServer
	kdsctx         *kds_context.Context
	rv             ResourceValidators
	au             authn.Authenticator
	acc            Access
	appCtx         context.Context
	extraReportsFn ExtraReportsFn
	tokenIssuers   builtin.TokenIssuers
	meshCache      *mesh.Cache
	interCpPool    *client.Pool
	*runtimeInfo
	pgxConfigCustomizationFn config.PgxConfigCustomization
	tenants                  multitenant.Tenants
	apiWebServiceCustomize   []func(*restful.WebService) error
}

func BuilderFor(appCtx context.Context, cfg kuma_cp.Config) (*Builder, error) {
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
			startTime:  time.Now(),
			mode:       cfg.Mode,
		},
		appCtx: appCtx,
	}, nil
}

func (b *Builder) WithComponentManager(cm component.Manager) *Builder {
	b.cm = cm
	return b
}

func (b *Builder) WithResourceStore(rs core_store.CustomizableResourceStore) *Builder {
	b.rs = rs
	return b
}

func (b *Builder) WithTransactions(txs core_store.Transactions) *Builder {
	b.txs = txs
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

func (b *Builder) WithGlobalInsightService(gis globalinsight.GlobalInsightService) *Builder {
	b.gis = gis
	return b
}

func (b *Builder) WithResourceManager(rm core_manager.CustomizableResourceManager) *Builder {
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

func (b *Builder) WithEventBus(erf events.EventBus) *Builder {
	b.erf = erf
	return b
}

func (b *Builder) WithAPIManager(apim api_server.APIManager) *Builder {
	b.apim = apim
	return b
}

func (b *Builder) WithCAProvider(cap secrets.CaProvider) *Builder {
	b.cap = cap
	return b
}

func (b *Builder) WithDpServer(dps *dp_server.DpServer) *Builder {
	b.dps = dps
	return b
}

func (b *Builder) WithResourceValidators(rv ResourceValidators) *Builder {
	b.rv = rv
	return b
}

func (b *Builder) WithKDSContext(kdsctx *kds_context.Context) *Builder {
	b.kdsctx = kdsctx
	return b
}

func (b *Builder) WithXDS(xds xds_runtime.XDSRuntimeContext) *Builder {
	b.xds = xds
	return b
}

func (b *Builder) WithAPIServerAuthenticator(au authn.Authenticator) *Builder {
	b.au = au
	return b
}

func (b *Builder) WithAccess(acc Access) *Builder {
	b.acc = acc
	return b
}

func (b *Builder) WithExtraReportsFn(fn ExtraReportsFn) *Builder {
	b.extraReportsFn = fn
	return b
}

func (b *Builder) WithTokenIssuers(tokenIssuers builtin.TokenIssuers) *Builder {
	b.tokenIssuers = tokenIssuers
	return b
}

func (b *Builder) WithMeshCache(meshCache *mesh.Cache) *Builder {
	b.meshCache = meshCache
	return b
}

func (b *Builder) WithInterCPClientPool(interCpPool *client.Pool) *Builder {
	b.interCpPool = interCpPool
	return b
}

func (b *Builder) WithMultitenancy(tenants multitenant.Tenants) *Builder {
	b.tenants = tenants
	return b
}

func (b *Builder) WithPgxConfigCustomizationFn(pgxConfigCustomizationFn config.PgxConfigCustomization) *Builder {
	b.pgxConfigCustomizationFn = pgxConfigCustomizationFn
	return b
}

func (b *Builder) WithAPIWebServiceCustomize(customize func(*restful.WebService) error) *Builder {
	b.apiWebServiceCustomize = append(b.apiWebServiceCustomize, customize)
	return b
}

func (b *Builder) Build() (Runtime, error) {
	if b.cm == nil {
		return nil, errors.Errorf("ComponentManager has not been configured")
	}
	if b.rs == nil {
		return nil, errors.Errorf("ResourceStore has not been configured")
	}
	if b.txs == nil {
		return nil, errors.Errorf("Transactions has not been configured")
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
	if b.xds == (xds_runtime.XDSRuntimeContext{}) {
		return nil, errors.New("xds is not configured")
	}
	if b.cap == nil {
		return nil, errors.Errorf("CAProvider has not been configured")
	}
	if b.dps == nil {
		return nil, errors.Errorf("DpServer has not been configured")
	}
	if b.kdsctx == nil {
		return nil, errors.Errorf("KDSContext has not been configured")
	}
	if b.rv == (ResourceValidators{}) {
		return nil, errors.Errorf("ResourceValidators have not been configured")
	}
	if b.au == nil {
		return nil, errors.Errorf("API Server Authenticator has not been configured")
	}
	if b.acc == (Access{}) {
		return nil, errors.Errorf("Access has not been configured")
	}
	if b.tokenIssuers == (builtin.TokenIssuers{}) {
		return nil, errors.Errorf("TokenIssuers has not been configured")
	}
	if b.meshCache == nil {
		return nil, errors.Errorf("MeshCache has not been configured")
	}
	if b.interCpPool == nil {
		return nil, errors.Errorf("InterCP client pool has not been configured")
	}
	if b.pgxConfigCustomizationFn == nil {
		return nil, errors.Errorf("PgxConfigCustomizationFn has not been configured")
	}
	if b.tenants == nil {
		return nil, errors.Errorf("Tenants has not been configured")
	}
	return &runtime{
		RuntimeInfo: b.runtimeInfo,
		RuntimeContext: &runtimeContext{
			cfg:                      b.cfg,
			rm:                       b.rm,
			rom:                      b.rom,
			rs:                       b.rs,
			txs:                      b.txs,
			ss:                       b.ss,
			cam:                      b.cam,
			gis:                      b.gis,
			dsl:                      b.dsl,
			ext:                      b.ext,
			configm:                  b.configm,
			leadInfo:                 b.leadInfo,
			lif:                      b.lif,
			eac:                      b.eac,
			metrics:                  b.metrics,
			erf:                      b.erf,
			apim:                     b.apim,
			xds:                      b.xds,
			cap:                      b.cap,
			dps:                      b.dps,
			kdsctx:                   b.kdsctx,
			rv:                       b.rv,
			au:                       b.au,
			acc:                      b.acc,
			appCtx:                   b.appCtx,
			extraReportsFn:           b.extraReportsFn,
			tokenIssuers:             b.tokenIssuers,
			meshCache:                b.meshCache,
			interCpPool:              b.interCpPool,
			pgxConfigCustomizationFn: b.pgxConfigCustomizationFn,
			tenants:                  b.tenants,
			apiWebServiceCustomize:   b.apiWebServiceCustomize,
		},
		Manager: b.cm,
	}, nil
}

func (b *Builder) ComponentManager() component.Manager {
	return b.cm
}

func (b *Builder) ResourceStore() core_store.CustomizableResourceStore {
	return b.rs
}

func (b *Builder) Transactions() core_store.Transactions {
	return b.txs
}

func (b *Builder) SecretStore() store.SecretStore {
	return b.ss
}

func (b *Builder) ConfigStore() core_store.ResourceStore {
	return b.cs
}

func (b *Builder) GlobalInsightService() globalinsight.GlobalInsightService {
	return b.gis
}

func (b *Builder) ResourceManager() core_manager.CustomizableResourceManager {
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

func (b *Builder) EventBus() events.EventBus {
	return b.erf
}

func (b *Builder) APIManager() api_server.APIManager {
	return b.apim
}

func (b *Builder) CAProvider() secrets.CaProvider {
	return b.cap
}

func (b *Builder) DpServer() *dp_server.DpServer {
	return b.dps
}

func (b *Builder) KDSContext() *kds_context.Context {
	return b.kdsctx
}

func (b *Builder) ResourceValidators() ResourceValidators {
	return b.rv
}

func (b *Builder) APIServerAuthenticator() authn.Authenticator {
	return b.au
}

func (b *Builder) Access() Access {
	return b.acc
}

func (b *Builder) AppCtx() context.Context {
	return b.appCtx
}

func (b *Builder) ExtraReportsFn() ExtraReportsFn {
	return b.extraReportsFn
}

func (b *Builder) TokenIssuers() builtin.TokenIssuers {
	return b.tokenIssuers
}

func (b *Builder) EnvoyAdminClient() admin.EnvoyAdminClient {
	return b.eac
}

func (b *Builder) MeshCache() *mesh.Cache {
	return b.meshCache
}

func (b *Builder) InterCPClientPool() *client.Pool {
	return b.interCpPool
}

func (b *Builder) XDS() xds_runtime.XDSRuntimeContext {
	return b.xds
}

func (b *Builder) PgxConfigCustomizationFn() config.PgxConfigCustomization {
	return b.pgxConfigCustomizationFn
}

func (b *Builder) Tenants() multitenant.Tenants {
	return b.tenants
}

func (b *Builder) APIWebServiceCustomize() []func(*restful.WebService) error {
	return b.apiWebServiceCustomize
}
