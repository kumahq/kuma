package bootstrap

import (
	"context"
	"net"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/api-server/customization"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/access"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	"github.com/kumahq/kuma/pkg/core/managers/apis/dataplaneinsight"
	externalservice_managers "github.com/kumahq/kuma/pkg/core/managers/apis/external_service"
	mesh_managers "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	ratelimit_managers "github.com/kumahq/kuma/pkg/core/managers/apis/ratelimit"
	"github.com/kumahq/kuma/pkg/core/managers/apis/zone"
	"github.com/kumahq/kuma/pkg/core/managers/apis/zoneegressinsight"
	"github.com/kumahq/kuma/pkg/core/managers/apis/zoneingressinsight"
	"github.com/kumahq/kuma/pkg/core/managers/apis/zoneinsight"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	resources_access "github.com/kumahq/kuma/pkg/core/resources/access"
	core_apis "github.com/kumahq/kuma/pkg/core/resources/apis"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	runtime_reports "github.com/kumahq/kuma/pkg/core/runtime/reports"
	secret_cipher "github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/dp-server/server"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	envoyadmin_access "github.com/kumahq/kuma/pkg/envoy/admin/access"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/insights/globalinsight"
	"github.com/kumahq/kuma/pkg/intercp"
	"github.com/kumahq/kuma/pkg/intercp/catalog"
	"github.com/kumahq/kuma/pkg/intercp/envoyadmin"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	kds_envoyadmin "github.com/kumahq/kuma/pkg/kds/envoyadmin"
	"github.com/kumahq/kuma/pkg/metrics"
	metrics_store "github.com/kumahq/kuma/pkg/metrics/store"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/plugins/policies"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/graph"
	"github.com/kumahq/kuma/pkg/plugins/resources/postgres/config"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	tokens_access "github.com/kumahq/kuma/pkg/tokens/builtin/access"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	zone2 "github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	zone_access "github.com/kumahq/kuma/pkg/tokens/builtin/zone/access"
	mesh_cache "github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_runtime "github.com/kumahq/kuma/pkg/xds/runtime"
	"github.com/kumahq/kuma/pkg/xds/secrets"
	xds_server "github.com/kumahq/kuma/pkg/xds/server"
)

var log = core.Log.WithName("bootstrap")

func buildRuntime(appCtx context.Context, cfg kuma_cp.Config) (core_runtime.Runtime, error) {
	if err := autoconfigure(&cfg); err != nil {
		return nil, err
	}
	builder, err := core_runtime.BuilderFor(appCtx, cfg)
	if err != nil {
		return nil, err
	}
	core_plugins.Init(cfg.CoreResources.Enabled, core_apis.NameToModule)
	core_plugins.Init(cfg.Policies.Enabled, policies.NameToModule)
	builder.WithMultitenancy(multitenant.SingleTenant)
	builder.WithPgxConfigCustomizationFn(config.NoopPgxConfigCustomizationFn)
	for _, plugin := range core_plugins.Plugins().BootstrapPlugins() {
		if err := plugin.BeforeBootstrap(builder, cfg); err != nil {
			return nil, errors.Wrapf(err, "failed to run beforeBootstrap plugin:'%s'", plugin.Name())
		}
	}
	if err := initializeMetrics(builder); err != nil {
		return nil, err
	}
	if err := initializeResourceStore(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeSecretStore(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeConfigStore(cfg, builder); err != nil {
		return nil, err
	}

	initializeGlobalInsightService(cfg, builder)

	// we add Secret store to unified ResourceStore so global<->zone synchronizer can use unified interface
	builder.ResourceStore().Customize(system.SecretType, builder.SecretStore())
	builder.ResourceStore().Customize(system.GlobalSecretType, builder.SecretStore())
	builder.ResourceStore().Customize(system.ConfigType, builder.ConfigStore())

	initializeConfigManager(builder)

	builder.WithResourceValidators(core_runtime.ResourceValidators{
		Dataplane: dataplane.NewMembershipValidator(),
		Mesh:      mesh_managers.NewMeshValidator(builder.CaManagers(), builder.ResourceStore()),
	})

	if err := initializeResourceManager(cfg, builder); err != nil { //nolint:contextcheck
		return nil, err
	}

	builder.WithDataSourceLoader(datasource.NewDataSourceLoader(builder.ReadOnlyResourceManager()))

	if err := initializeCaManagers(builder); err != nil {
		return nil, err
	}

	leaderInfoComponent := &component.LeaderInfoComponent{}
	builder.WithLeaderInfo(leaderInfoComponent)

	builder.WithLookupIP(lookup.CachedLookupIP(net.LookupIP, cfg.General.DNSCacheTTL.Duration))
	builder.WithAPIManager(customization.NewAPIList())
	caProvider, err := secrets.NewCaProvider(builder.CaManagers(), builder.Metrics())
	if err != nil {
		return nil, err
	}
	builder.WithCAProvider(caProvider)
	builder.WithDpServer(server.NewDpServer(*cfg.DpServer, builder.Metrics(), func(writer http.ResponseWriter, request *http.Request) bool {
		return true
	}))
	resourceManager := builder.ResourceManager()
	kdsContext := kds_context.DefaultContext(appCtx, resourceManager, cfg)
	builder.WithKDSContext(kdsContext)
	builder.WithInterCPClientPool(intercp.DefaultClientPool())

	if cfg.Mode == config_core.Global {
		kdsEnvoyAdminClient := kds_envoyadmin.NewClient(
			builder.KDSContext().EnvoyAdminRPCs,
			builder.ReadOnlyResourceManager(),
		)
		forwardingClient := envoyadmin.NewForwardingEnvoyAdminClient(
			builder.ReadOnlyResourceManager(),
			catalog.NewConfigCatalog(resourceManager),
			builder.GetInstanceId(),
			intercp.PooledEnvoyAdminClientFn(builder.InterCPClientPool()),
			kdsEnvoyAdminClient,
		)
		builder.WithEnvoyAdminClient(forwardingClient)
	} else {
		builder.WithEnvoyAdminClient(admin.NewEnvoyAdminClient(
			resourceManager,
			builder.CaManagers(),
			builder.Config().GetEnvoyAdminPort(),
		))
	}

	xdsCtx, err := xds_runtime.WithDefaults(builder) //nolint:contextcheck
	if err != nil {
		return nil, err
	}
	builder.WithXDS(xdsCtx)

	builder.WithAccess(core_runtime.Access{
		ResourceAccess:       resources_access.NewAdminResourceAccess(builder.Config().Access.Static.AdminResources),
		DataplaneTokenAccess: tokens_access.NewStaticGenerateDataplaneTokenAccess(builder.Config().Access.Static.GenerateDPToken),
		ZoneTokenAccess:      zone_access.NewStaticZoneTokenAccess(builder.Config().Access.Static.GenerateZoneToken),
		EnvoyAdminAccess: envoyadmin_access.NewStaticEnvoyAdminAccess(
			builder.Config().Access.Static.ViewConfigDump,
			builder.Config().Access.Static.ViewStats,
			builder.Config().Access.Static.ViewClusters,
		),
		ControlPlaneMetadataAccess: access.NewStaticControlPlaneMetadataAccess(builder.Config().Access.Static.ControlPlaneMetadata),
	})

	if err := initializeAPIServerAuthenticator(builder); err != nil {
		return nil, err
	}

	initializeTokenIssuers(builder)

	if err := initializeMeshCache(builder); err != nil {
		return nil, err
	}

	for _, plugin := range core_plugins.Plugins().BootstrapPlugins() {
		if err := plugin.AfterBootstrap(builder, cfg); err != nil {
			return nil, errors.Wrapf(err, "failed to run afterBootstrap plugin:'%s'", plugin.Name())
		}
	}
	rt, err := builder.Build()
	if err != nil {
		return nil, err
	}

	if err := rt.Add(leaderInfoComponent); err != nil {
		return nil, err
	}

	for name, plugin := range core_plugins.Plugins().RuntimePlugins() {
		if err := plugin.Customize(rt); err != nil {
			return nil, errors.Wrapf(err, "failed to configure runtime plugin:'%s'", name)
		}
	}

	logWarnings(rt.Config())

	return rt, nil
}

func logWarnings(config kuma_cp.Config) {
	if config.ApiServer.Authn.LocalhostIsAdmin {
		log.Info("WARNING: you can access Control Plane API as admin by sending requests from the same machine where Control Plane runs. To increase security, it is recommended to extract admin credentials and set KUMA_API_SERVER_AUTHN_LOCALHOST_IS_ADMIN to false.")
	}
}

func initializeMetrics(builder *core_runtime.Builder) error {
	if builder.Metrics() != nil {
		// do not configure if it was already configured in BeforeBootstrap
		return nil
	}
	zoneName := metrics.ZoneNameOrMode(builder.Config().Mode, builder.Config().Multizone.Zone.Name)
	metrics, err := metrics.NewMetrics(zoneName)
	if err != nil {
		return err
	}
	builder.WithMetrics(metrics)
	return nil
}

func Bootstrap(appCtx context.Context, cfg kuma_cp.Config) (core_runtime.Runtime, error) {
	runtime, err := buildRuntime(appCtx, cfg)
	if err != nil {
		return nil, err
	}

	if err = startReporter(runtime); err != nil { //nolint:contextcheck
		return nil, err
	}

	return runtime, nil
}

func startReporter(runtime core_runtime.Runtime) error {
	return runtime.Add(component.ComponentFunc(func(stop <-chan struct{}) error {
		runtime_reports.Init(runtime, runtime.Config(), runtime.ExtraReportsFn())
		<-stop
		return nil
	}))
}

func initializeResourceStore(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	var pluginConfig core_plugins.PluginConfig
	switch cfg.Store.Type {
	case store.KubernetesStore:
		pluginName = core_plugins.Kubernetes
		pluginConfig = nil
	case store.MemoryStore:
		pluginName = core_plugins.Memory
		pluginConfig = nil
	case store.PostgresStore:
		pluginName = core_plugins.Postgres
		pluginConfig = cfg.Store.Postgres
	default:
		return errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	plugin, err := core_plugins.Plugins().ResourceStore(pluginName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve store %s plugin", pluginName)
	}

	rs, transactions, err := plugin.NewResourceStore(builder, pluginConfig)
	if err != nil {
		return err
	}
	builder.WithResourceStore(core_store.NewCustomizableResourceStore(rs))
	builder.WithTransactions(transactions)
	eventBus, err := events.NewEventBus(cfg.EventBus.BufferSize, builder.Metrics())
	if err != nil {
		return err
	}
	if err := plugin.EventListener(builder, eventBus); err != nil {
		return err
	}
	builder.WithEventBus(eventBus)

	paginationStore := core_store.NewPaginationStore(rs)
	meteredStore, err := metrics_store.NewMeteredStore(paginationStore, builder.Metrics())
	if err != nil {
		return err
	}

	builder.WithResourceStore(core_store.NewCustomizableResourceStore(meteredStore))
	return nil
}

func initializeSecretStore(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	var pluginConfig core_plugins.PluginConfig
	switch cfg.Store.Type {
	case store.KubernetesStore:
		pluginName = core_plugins.Kubernetes
	case store.MemoryStore, store.PostgresStore:
		pluginName = core_plugins.Universal
	default:
		return errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	plugin, err := core_plugins.Plugins().SecretStore(pluginName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve secret store %s plugin", pluginName)
	}
	if ss, err := plugin.NewSecretStore(builder, pluginConfig); err != nil {
		return err
	} else {
		builder.WithSecretStore(ss)
		return nil
	}
}

func initializeConfigStore(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	var pluginConfig core_plugins.PluginConfig
	switch cfg.Store.Type {
	case store.KubernetesStore:
		pluginName = core_plugins.Kubernetes
	case store.MemoryStore, store.PostgresStore:
		pluginName = core_plugins.Universal
	default:
		return errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	plugin, err := core_plugins.Plugins().ConfigStore(pluginName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve secret store %s plugin", pluginName)
	}
	if cs, err := plugin.NewConfigStore(builder, pluginConfig); err != nil {
		return err
	} else {
		builder.WithConfigStore(cs)
		return nil
	}
}

func initializeGlobalInsightService(cfg kuma_cp.Config, builder *core_runtime.Builder) {
	globalInsightService := globalinsight.NewDefaultGlobalInsightService(builder.ResourceStore())
	if cfg.Store.Cache.Enabled {
		globalInsightService = globalinsight.NewCachedGlobalInsightService(
			globalInsightService,
			builder.Tenants(),
			cfg.Store.Cache.ExpirationTime.Duration,
		)
	}

	builder.WithGlobalInsightService(globalInsightService)
}

func initializeCaManagers(builder *core_runtime.Builder) error {
	for pluginName, caPlugin := range core_plugins.Plugins().CaPlugins() {
		caManager, err := caPlugin.NewCaManager(builder, nil)
		if err != nil {
			return errors.Wrapf(err, "could not create CA manager for plugin %q", pluginName)
		}
		builder.WithCaManager(string(pluginName), caManager)
	}
	return nil
}

func initializeAPIServerAuthenticator(builder *core_runtime.Builder) error {
	authnType := builder.Config().ApiServer.Authn.Type
	plugin, ok := core_plugins.Plugins().AuthnAPIServer()[core_plugins.PluginName(authnType)]
	if !ok {
		return errors.Errorf("there is not implementation of authn named %s", authnType)
	}
	authenticator, err := plugin.NewAuthenticator(builder)
	if err != nil {
		return errors.Wrapf(err, "could not initiate authenticator %s", authnType)
	}
	builder.WithAPIServerAuthenticator(authenticator)
	return nil
}

func initializeResourceManager(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, nil)

	customizableManager.Customize(
		mesh.MeshType,
		mesh_managers.NewMeshManager(
			builder.ResourceStore(),
			customizableManager,
			builder.CaManagers(),
			registry.Global(),
			builder.ResourceValidators().Mesh,
			builder.Extensions(),
			cfg,
		),
	)

	rateLimitValidator := ratelimit_managers.RateLimitValidator{
		Store: builder.ResourceStore(),
	}
	customizableManager.Customize(
		mesh.RateLimitType,
		ratelimit_managers.NewRateLimitManager(builder.ResourceStore(), rateLimitValidator),
	)

	externalServiceValidator := externalservice_managers.ExternalServiceValidator{
		Store: builder.ResourceStore(),
	}
	customizableManager.Customize(
		mesh.ExternalServiceType,
		externalservice_managers.NewExternalServiceManager(builder.ResourceStore(), externalServiceValidator),
	)

	customizableManager.Customize(
		mesh.DataplaneType,
		dataplane.NewDataplaneManager(
			builder.ResourceStore(),
			builder.Config().Multizone.Zone.Name,
			builder.Config().Mode,
			builder.Config().Environment == config_core.KubernetesEnvironment,
			builder.Config().Store.Kubernetes.SystemNamespace,
			builder.ResourceValidators().Dataplane,
		),
	)

	customizableManager.Customize(
		mesh.DataplaneInsightType,
		dataplaneinsight.NewDataplaneInsightManager(builder.ResourceStore(), builder.Config().Metrics.Dataplane),
	)

	customizableManager.Customize(
		system.ZoneType,
		zone.NewZoneManager(builder.ResourceStore(), zone.Validator{Store: builder.ResourceStore()}, builder.Config().Store.UnsafeDelete),
	)

	customizableManager.Customize(
		system.ZoneInsightType,
		zoneinsight.NewZoneInsightManager(builder.ResourceStore(), builder.Config().Metrics.Zone),
	)

	customizableManager.Customize(
		mesh.ZoneIngressInsightType,
		zoneingressinsight.NewZoneIngressInsightManager(builder.ResourceStore(), builder.Config().Metrics.Dataplane),
	)

	customizableManager.Customize(
		mesh.ZoneEgressInsightType,
		zoneegressinsight.NewZoneEgressInsightManager(builder.ResourceStore(), builder.Config().Metrics.Dataplane),
	)

	var cipher secret_cipher.Cipher
	switch cfg.Store.Type {
	case store.KubernetesStore:
		cipher = secret_cipher.None() // deliberately turn encryption off on Kubernetes
	case store.MemoryStore, store.PostgresStore:
		cipher = secret_cipher.TODO() // get back to encryption in universal case
	default:
		return errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	var secretValidator secret_manager.SecretValidator
	if cfg.IsFederatedZoneCP() {
		secretValidator = secret_manager.ValidateDelete(func(ctx context.Context, secretName string, secretMesh string) error { return nil })
	} else {
		secretValidator = secret_manager.NewSecretValidator(builder.CaManagers(), builder.ResourceStore())
	}

	customizableManager.Customize(
		system.SecretType,
		secret_manager.NewSecretManager(builder.SecretStore(), cipher, secretValidator, cfg.Store.UnsafeDelete),
	)

	customizableManager.Customize(
		system.GlobalSecretType,
		secret_manager.NewGlobalSecretManager(builder.SecretStore(), cipher),
	)

	builder.WithResourceManager(customizableManager)

	if builder.Config().Store.Cache.Enabled {
		cachedManager, err := core_manager.NewCachedManager(
			customizableManager,
			builder.Config().Store.Cache.ExpirationTime.Duration,
			builder.Metrics(),
			builder.Tenants(),
		)
		if err != nil {
			return err
		}
		builder.WithReadOnlyResourceManager(cachedManager)
	} else {
		builder.WithReadOnlyResourceManager(customizableManager)
	}
	return nil
}

func initializeConfigManager(builder *core_runtime.Builder) {
	builder.WithConfigManager(config_manager.NewConfigManager(builder.ConfigStore()))
}

func initializeMeshCache(builder *core_runtime.Builder) error {
	rsGraphBuilder := xds_context.AnyToAnyReachableServicesGraphBuilder
	if builder.Config().Experimental.AutoReachableServices {
		rsGraphBuilder = graph.Builder
	}
	meshContextBuilder := xds_context.NewMeshContextBuilder(
		builder.ReadOnlyResourceManager(),
		xds_server.MeshResourceTypes(),
		builder.LookupIP(),
		builder.Config().Multizone.Zone.Name,
		vips.NewPersistence(builder.ReadOnlyResourceManager(), builder.ConfigManager(), builder.Config().Experimental.UseTagFirstVirtualOutboundModel),
		builder.Config().DNSServer.Domain,
		builder.Config().DNSServer.ServiceVipPort,
		rsGraphBuilder,
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

func initializeTokenIssuers(builder *core_runtime.Builder) {
	issuers := builtin.TokenIssuers{}
	if builder.Config().DpServer.Authn.DpProxy.DpToken.EnableIssuer {
		issuers.DataplaneToken = builtin.NewDataplaneTokenIssuer(builder.ResourceManager())
	} else {
		issuers.DataplaneToken = issuer.DisabledIssuer{}
	}
	if builder.Config().DpServer.Authn.ZoneProxy.ZoneToken.EnableIssuer {
		issuers.ZoneToken = builtin.NewZoneTokenIssuer(builder.ResourceManager())
	} else {
		issuers.ZoneToken = zone2.DisabledIssuer{}
	}
	builder.WithTokenIssuers(issuers)
}
