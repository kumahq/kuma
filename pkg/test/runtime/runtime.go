package runtime

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/api-server/customization"
	kuma_cp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	dp_server "github.com/kumahq/kuma/v2/pkg/config/dp-server"
	"github.com/kumahq/kuma/v2/pkg/core/access"
	config_manager "github.com/kumahq/kuma/v2/pkg/core/config/manager"
	"github.com/kumahq/kuma/v2/pkg/core/datasource"
	"github.com/kumahq/kuma/v2/pkg/core/managers/apis/dataplane"
	mesh_managers "github.com/kumahq/kuma/v2/pkg/core/managers/apis/mesh"
	resources_access "github.com/kumahq/kuma/v2/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	core_runtime "github.com/kumahq/kuma/v2/pkg/core/runtime"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	secret_cipher "github.com/kumahq/kuma/v2/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/v2/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/v2/pkg/core/secrets/store"
	"github.com/kumahq/kuma/v2/pkg/dns/vips"
	"github.com/kumahq/kuma/v2/pkg/dp-server/server"
	envoyadmin_access "github.com/kumahq/kuma/v2/pkg/envoy/admin/access"
	"github.com/kumahq/kuma/v2/pkg/events"
	"github.com/kumahq/kuma/v2/pkg/intercp"
	kds_context "github.com/kumahq/kuma/v2/pkg/kds/context"
	"github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/multitenant"
	"github.com/kumahq/kuma/v2/pkg/plugins/authn/api-server/certs"
	"github.com/kumahq/kuma/v2/pkg/plugins/ca/builtin"
	leader_memory "github.com/kumahq/kuma/v2/pkg/plugins/leader/memory"
	resources_memory "github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/postgres/config"
	tokens_builtin "github.com/kumahq/kuma/v2/pkg/tokens/builtin"
	tokens_access "github.com/kumahq/kuma/v2/pkg/tokens/builtin/access"
	mesh_cache "github.com/kumahq/kuma/v2/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	xds_runtime "github.com/kumahq/kuma/v2/pkg/xds/runtime"
	"github.com/kumahq/kuma/v2/pkg/xds/secrets"
	xds_server "github.com/kumahq/kuma/v2/pkg/xds/server"
)

var _ core_runtime.RuntimeInfo = &TestRuntimeInfo{}

type TestRuntimeInfo struct {
	InstanceId string
	ClusterId  string
	StartTime  time.Time
	Mode       config_core.CpMode
}

func (i *TestRuntimeInfo) GetMode() config_core.CpMode {
	return i.Mode
}

func (i *TestRuntimeInfo) GetInstanceId() string {
	return i.InstanceId
}

func (i *TestRuntimeInfo) SetClusterId(clusterId string) {
	i.ClusterId = clusterId
}

func (i *TestRuntimeInfo) GetClusterId() string {
	return i.ClusterId
}

func (i *TestRuntimeInfo) GetStartTime() time.Time {
	return i.StartTime
}

func BuilderFor(appCtx context.Context, cfg kuma_cp.Config) (*core_runtime.Builder, error) {
	if cfg.DpServer.Authn.DpProxy.Type == "" {
		cfg.DpServer.Authn.DpProxy.Type = dp_server.DpServerAuthDpToken
	}
	if cfg.DpServer.Authn.ZoneProxy.Type == "" {
		cfg.DpServer.Authn.ZoneProxy.Type = dp_server.DpServerAuthZoneToken
	}
	builder, err := core_runtime.BuilderFor(appCtx, cfg)
	if err != nil {
		return nil, err
	}

	builder.
		WithComponentManager(component.NewManager(leader_memory.NewAlwaysLeaderElector())).
		WithResourceStore(store.NewCustomizableResourceStore(store.NewPaginationStore(resources_memory.NewStore()))).
		WithTransactions(store.NoTransactions{}).
		WithSecretStore(secret_store.NewSecretStore(builder.ResourceStore())).
		WithResourceValidators(core_runtime.ResourceValidators{
			Dataplane: dataplane.NewMembershipValidator(),
			Mesh:      mesh_managers.NewMeshValidator(builder.CaManagers(), builder.ResourceStore()),
		})

	rm := newResourceManager(builder) //nolint:contextcheck
	builder.WithResourceManager(rm).
		WithReadOnlyResourceManager(rm)

	metrics, _ := metrics.NewMetrics("Zone")
	builder.WithMetrics(metrics)

	builder.WithDataSourceLoader(datasource.NewDataSourceLoader(builder.ResourceManager()))
	builder.WithCaManager("builtin", builtin.NewBuiltinCaManager(builder.ResourceManager()))
	builder.WithLeaderInfo(&component.LeaderInfoComponent{})
	builder.WithLookupIP(func(s string) ([]net.IP, error) {
		return nil, errors.New("LookupIP not set, set one in your test to resolve things")
	})
	builder.WithEnvoyAdminClient(&DummyEnvoyAdminClient{})
	eventBus, err := events.NewEventBus(10, metrics)
	if err != nil {
		return nil, err
	}
	builder.WithEventBus(eventBus)
	builder.WithAPIManager(customization.NewAPIList())
	xdsCtx, err := xds_runtime.WithDefaults(builder) //nolint:contextcheck
	if err != nil {
		return nil, err
	}
	builder.WithXDS(xdsCtx)
	builder.WithDpServer(server.NewDpServer(*cfg.DpServer, metrics, func(writer http.ResponseWriter, request *http.Request) bool {
		return true
	}))
	builder.WithKDSContext(kds_context.DefaultContext(appCtx, builder.ResourceManager(), cfg))
	caProvider, err := secrets.NewCaProvider(builder.CaManagers(), metrics)
	if err != nil {
		return nil, err
	}
	builder.WithCAProvider(caProvider)
	builder.WithAPIServerAuthenticator(certs.ClientCertAuthenticator)
	builder.WithAccess(core_runtime.Access{
		ResourceAccess:       resources_access.NewAdminResourceAccess(builder.Config().Access.Static.AdminResources),
		DataplaneTokenAccess: tokens_access.NewStaticGenerateDataplaneTokenAccess(builder.Config().Access.Static.GenerateDPToken),
		EnvoyAdminAccess: envoyadmin_access.NewStaticEnvoyAdminAccess(
			builder.Config().Access.Static.ViewConfigDump,
			builder.Config().Access.Static.ViewStats,
			builder.Config().Access.Static.ViewClusters,
		),
		ControlPlaneMetadataAccess: access.NewStaticControlPlaneMetadataAccess(builder.Config().Access.Static.ControlPlaneMetadata),
	})
	builder.WithTokenIssuers(tokens_builtin.TokenIssuers{
		DataplaneToken: tokens_builtin.NewDataplaneTokenIssuer(builder.ResourceManager()),
		ZoneToken:      tokens_builtin.NewZoneTokenIssuer(builder.ResourceManager()),
	})
	builder.WithInterCPClientPool(intercp.DefaultClientPool())
	builder.WithMultitenancy(multitenant.SingleTenant)
	builder.WithPgxConfigCustomizationFn(config.NoopPgxConfigCustomizationFn)

	initializeConfigManager(builder)

	err = initializeMeshCache(builder)
	if err != nil {
		return nil, err
	}

	return builder, nil
}

func initializeConfigManager(builder *core_runtime.Builder) {
	configm := config_manager.NewConfigManager(builder.ResourceStore())
	builder.WithConfigManager(configm)
}

func newResourceManager(builder *core_runtime.Builder) core_manager.CustomizableResourceManager {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
	meshManager := mesh_managers.NewMeshManager(
		builder.ResourceStore(),
		customizableManager,
		builder.CaManagers(),
		registry.Global(),
		builder.ResourceValidators().Mesh,
		builder.Extensions(),
		builder.Config(),
	)
	customManagers[core_mesh.MeshType] = meshManager

	secretManager := secret_manager.NewSecretManager(builder.SecretStore(), secret_cipher.None(), nil, builder.Config().Store.UnsafeDelete)
	customManagers[system.SecretType] = secretManager
	return customizableManager
}

func initializeMeshCache(builder *core_runtime.Builder) error {
	meshContextBuilder := xds_context.NewMeshContextBuilder(
		builder.ReadOnlyResourceManager(),
		xds_server.MeshResourceTypes(),
		builder.LookupIP(),
		builder.Config().Multizone.Zone.Name,
		vips.NewPersistence(builder.ReadOnlyResourceManager(), builder.ConfigManager(), builder.Config().Experimental.UseTagFirstVirtualOutboundModel),
		builder.Config().DNSServer.Domain,
		builder.Config().DNSServer.ServiceVipPort,
		xds_context.AnyToAnyReachableServicesGraphBuilder,
		builder.Config().Experimental.SkipPersistedVIPs,
	)

	meshSnapshotCache, err := mesh_cache.NewCache(
		builder.Config().Store.Cache.ExpirationTime.Duration,
		meshContextBuilder,
		builder.Metrics(),
	)
	if err != nil {
		return err
	}

	builder.WithMeshCache(meshSnapshotCache)
	return nil
}

type DummyEnvoyAdminClient struct {
	PostQuitCalled   *int
	ConfigDumpCalled int
	StatsCalled      int
	ClustersCalled   int
}

func (d *DummyEnvoyAdminClient) Stats(ctx context.Context, proxy core_model.ResourceWithAddress, format v1alpha1.AdminOutputFormat) ([]byte, error) {
	d.StatsCalled++
	if format == v1alpha1.AdminOutputFormat_JSON {
		return []byte("{\"server.live\": 1}\n"), nil
	}
	return []byte("server.live: 1\n"), nil
}

func (d *DummyEnvoyAdminClient) Clusters(ctx context.Context, proxy core_model.ResourceWithAddress, format v1alpha1.AdminOutputFormat) ([]byte, error) {
	d.ClustersCalled++
	if format == v1alpha1.AdminOutputFormat_JSON {
		return []byte("{\"kuma\": \"envoy:admin\"}\n"), nil
	}
	return []byte("kuma:envoy:admin\n"), nil
}

func (d *DummyEnvoyAdminClient) GenerateAPIToken(dp *core_mesh.DataplaneResource) (string, error) {
	return "token", nil
}

func (d *DummyEnvoyAdminClient) PostQuit(ctx context.Context, dataplane *core_mesh.DataplaneResource) error {
	if d.PostQuitCalled != nil {
		*d.PostQuitCalled++
	}

	return nil
}

func (d *DummyEnvoyAdminClient) ConfigDump(ctx context.Context, proxy core_model.ResourceWithAddress, includeEds bool) ([]byte, error) {
	out := map[string]string{
		"envoyAdminAddress": proxy.AdminAddress(9901),
	}
	d.ConfigDumpCalled++
	if includeEds {
		out["eds"] = "eds"
	}
	return json.Marshal(out)
}
