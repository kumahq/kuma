package setup

import (
	"context"
	"time"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/api-server/authn"
	"github.com/kumahq/kuma/pkg/api-server/customization"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/dp-server/server"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/insights/globalinsight"
	"github.com/kumahq/kuma/pkg/intercp/client"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	reconcile_v2 "github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	kds_server_v2 "github.com/kumahq/kuma/pkg/kds/v2/server"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/plugins/resources/postgres/config"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_runtime "github.com/kumahq/kuma/pkg/xds/runtime"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

type testRuntimeContext struct {
	runtime.RuntimeInfo
	rom                      manager.ReadOnlyResourceManager
	rm                       manager.ResourceManager
	cfg                      kuma_cp.Config
	components               []component.Component
	metrics                  core_metrics.Metrics
	pgxConfigCustomizationFn config.PgxConfigCustomization
	tenants                  multitenant.Tenants
	eventBus                 events.EventBus
}

func (t *testRuntimeContext) DataSourceLoader() datasource.Loader {
	panic("implement me")
}

func (t *testRuntimeContext) ResourceStore() store.ResourceStore {
	panic("implement me")
}

func (t *testRuntimeContext) SecretStore() secret_store.SecretStore {
	panic("implement me")
}

func (t *testRuntimeContext) ConfigStore() store.ResourceStore {
	panic("implement me")
}

func (t *testRuntimeContext) GlobalInsightService() globalinsight.GlobalInsightService {
	panic("implement me")
}

func (t *testRuntimeContext) CaManagers() core_ca.Managers {
	panic("implement me")
}

func (t *testRuntimeContext) ConfigManager() config_manager.ConfigManager {
	panic("implement me")
}

func (t *testRuntimeContext) LeaderInfo() component.LeaderInfo {
	panic("implement me")
}

func (t *testRuntimeContext) LookupIP() lookup.LookupIPFunc {
	panic("implement me")
}

func (t *testRuntimeContext) EnvoyAdminClient() admin.EnvoyAdminClient {
	panic("implement me")
}

func (t *testRuntimeContext) APIInstaller() customization.APIInstaller {
	panic("implement me")
}

func (t *testRuntimeContext) XDS() xds_runtime.XDSRuntimeContext {
	panic("implement me")
}

func (t *testRuntimeContext) CAProvider() secrets.CaProvider {
	panic("implement me")
}

func (t *testRuntimeContext) DpServer() *server.DpServer {
	panic("implement me")
}

func (t *testRuntimeContext) KDSContext() *kds_context.Context {
	panic("implement me")
}

func (t *testRuntimeContext) APIServerAuthenticator() authn.Authenticator {
	panic("implement me")
}

func (t *testRuntimeContext) ResourceValidators() runtime.ResourceValidators {
	panic("implement me")
}

func (t *testRuntimeContext) Access() runtime.Access {
	panic("implement me")
}

func (t *testRuntimeContext) AppContext() context.Context {
	panic("implement me")
}

func (t *testRuntimeContext) ExtraReportsFn() runtime.ExtraReportsFn {
	panic("implement me")
}

func (t *testRuntimeContext) TokenIssuers() builtin.TokenIssuers {
	panic("implement me")
}

func (t *testRuntimeContext) MeshCache() *mesh.Cache {
	panic("implement me")
}

func (t *testRuntimeContext) InterCPClientPool() *client.Pool {
	panic("implement me")
}

func (t *testRuntimeContext) Start(i <-chan struct{}) error {
	panic("implement me")
}

func (t *testRuntimeContext) Config() kuma_cp.Config {
	return t.cfg
}

func (t *testRuntimeContext) ReadOnlyResourceManager() manager.ReadOnlyResourceManager {
	return t.rom
}

func (t *testRuntimeContext) ResourceManager() manager.ResourceManager {
	return t.rm
}

func (t *testRuntimeContext) Metrics() core_metrics.Metrics {
	return t.metrics
}

func (t *testRuntimeContext) PgxConfigCustomizationFn() config.PgxConfigCustomization {
	return t.pgxConfigCustomizationFn
}

func (t *testRuntimeContext) Tenants() multitenant.Tenants {
	return t.tenants
}

func (t *testRuntimeContext) Transactions() store.Transactions {
	return store.NoTransactions{}
}

func (t *testRuntimeContext) EventBus() events.EventBus {
	return t.eventBus
}

func (t *testRuntimeContext) Extensions() context.Context {
	return context.Background()
}

func (t *testRuntimeContext) APIWebServiceCustomize() func(*restful.WebService) error {
	return func(*restful.WebService) error { return nil }
}

func (t *testRuntimeContext) Ready() bool {
	for _, c := range t.components {
		if rc, ok := c.(component.ReadyComponent); ok && !rc.Ready() {
			return false
		}
	}
	return true
}

func (t *testRuntimeContext) Add(c ...component.Component) error {
	t.components = append(t.components, c...)
	return nil
}

type KdsServerBuilder struct {
	rt             *testRuntimeContext
	providedMapper reconcile_v2.ResourceMapper
	providedFilter reconcile_v2.ResourceFilter
	providedTypes  []model.ResourceType
}

func NewKdsServerBuilder(store store.ResourceStore) *KdsServerBuilder {
	metrics, err := core_metrics.NewMetrics("Global")
	if err != nil {
		panic(err)
	}
	rm := manager.NewResourceManager(store)
	eventBus, err := events.NewEventBus(20, metrics)
	if err != nil {
		panic(err)
	}
	cfg := kuma_cp.DefaultConfig()
	cfg.Mode = config_core.Global
	runtimeInfo := runtime.NewRuntimeInfo("global-cp", cfg.Mode)
	runtimeInfo.SetClusterInfo("my-cluster", time.Now())
	rt := &testRuntimeContext{
		RuntimeInfo:              runtimeInfo,
		rom:                      rm,
		rm:                       rm,
		cfg:                      cfg,
		metrics:                  metrics,
		tenants:                  multitenant.SingleTenant,
		pgxConfigCustomizationFn: config.NoopPgxConfigCustomizationFn,
		eventBus:                 eventBus,
	}
	return &KdsServerBuilder{
		rt:             rt,
		providedMapper: reconcile_v2.NoopResourceMapper,
		providedTypes:  registry.Global().ObjectTypes(model.HasKdsEnabled()),
		providedFilter: reconcile_v2.Any,
	}
}

func (b *KdsServerBuilder) AsZone(name string) *KdsServerBuilder {
	b.rt.cfg.Multizone.Zone.Name = name
	b.rt.cfg.Mode = config_core.Zone
	b.rt.RuntimeInfo = runtime.NewRuntimeInfo("zone-cp", b.rt.cfg.Mode)
	return b
}

func (b *KdsServerBuilder) WithKdsContext(kctx *kds_context.Context) *KdsServerBuilder {
	b.providedMapper = kctx.GlobalResourceMapper
	b.providedFilter = kctx.GlobalProvidedFilter
	return b
}

func (b *KdsServerBuilder) WithTypes(types []model.ResourceType) *KdsServerBuilder {
	b.providedTypes = types
	return b
}

func (b *KdsServerBuilder) Delta() (kds_server_v2.Server, error) {
	return kds_server_v2.New(core.Log.WithName("kds-delta").WithName(b.rt.GetMode()), b.rt, b.providedTypes, b.rt.Config().Multizone.Zone.Name, 100*time.Millisecond, b.providedFilter, b.providedMapper, 1*time.Second)
}
