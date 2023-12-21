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
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"
	kds_server_v2 "github.com/kumahq/kuma/pkg/kds/v2/server"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/plugins/resources/postgres/config"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
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

func (t *testRuntimeContext) Add(c ...component.Component) error {
	t.components = append(t.components, c...)
	return nil
}

func StartServer(store store.ResourceStore, clusterID string, providedTypes []model.ResourceType, providedFilter reconcile.ResourceFilter, providedMapper reconcile.ResourceMapper) (kds_server.Server, error) {
	metrics, err := core_metrics.NewMetrics("Global")
	if err != nil {
		return nil, err
	}
	rm := manager.NewResourceManager(store)
	rt := &testRuntimeContext{
		rom:                      rm,
		rm:                       rm,
		cfg:                      kuma_cp.DefaultConfig(),
		metrics:                  metrics,
		tenants:                  multitenant.SingleTenant,
		pgxConfigCustomizationFn: config.NoopPgxConfigCustomizationFn,
	}
	return kds_server.New(core.Log.WithName("kds").WithName(clusterID), rt, providedTypes, clusterID, 100*time.Millisecond, providedFilter, providedMapper, 1*time.Second)
}

func StartDeltaServer(store store.ResourceStore, mode config_core.CpMode, name string, providedTypes []model.ResourceType, providedFilter reconcile.ResourceFilter, providedMapper reconcile.ResourceMapper) (kds_server_v2.Server, error) {
	metrics, err := core_metrics.NewMetrics("Global")
	if err != nil {
		return nil, err
	}
	eventBus, err := events.NewEventBus(20, metrics)
	if err != nil {
		return nil, err
	}
	rm := manager.NewResourceManager(store)
	rt := &testRuntimeContext{
		RuntimeInfo: &test_runtime.TestRuntimeInfo{
			InstanceId: "my-instance",
			ClusterId:  name,
			Mode:       mode,
		},
		rom:                      rm,
		rm:                       rm,
		cfg:                      kuma_cp.DefaultConfig(),
		metrics:                  metrics,
		tenants:                  multitenant.SingleTenant,
		pgxConfigCustomizationFn: config.NoopPgxConfigCustomizationFn,
		eventBus:                 eventBus,
	}
	return kds_server_v2.New(core.Log.WithName("kds-delta").WithName(name), rt, providedTypes, name, 100*time.Millisecond, providedFilter, providedMapper, 1*time.Second)
}
