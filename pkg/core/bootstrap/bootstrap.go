package bootstrap

import (
	"context"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"net"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/api-server/customization"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
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
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
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
	"github.com/kumahq/kuma/pkg/envoy/admin/access"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/intercp"
	"github.com/kumahq/kuma/pkg/intercp/catalog"
	"github.com/kumahq/kuma/pkg/intercp/envoyadmin"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	"github.com/kumahq/kuma/pkg/metrics"
	metrics_store "github.com/kumahq/kuma/pkg/metrics/store"
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
	if err := initializeMetrics(builder); err != nil {
		return nil, err
	}
	for _, plugin := range core_plugins.Plugins().BootstrapPlugins() {
		if err := plugin.BeforeBootstrap(builder, cfg); err != nil {
			return nil, errors.Wrapf(err, "failed to run beforeBootstrap plugin:'%s'", plugin.Name())
		}
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
	// we add Secret store to unified ResourceStore so global<->zone synchronizer can use unified interface
	builder.WithResourceStore(core_store.NewCustomizableResourceStore(builder.ResourceStore(), map[core_model.ResourceType]core_store.ResourceStore{
		system.SecretType:       builder.SecretStore(),
		system.GlobalSecretType: builder.SecretStore(),
		system.ConfigType:       builder.ConfigStore(),
	}))

	initializeConfigManager(builder)

	builder.WithResourceValidators(core_runtime.ResourceValidators{
		Dataplane: dataplane.NewMembershipValidator(),
		Mesh:      mesh_managers.NewMeshValidator(builder.CaManagers(), builder.ResourceStore()),
	})

	if err := initializeResourceManager(cfg, builder); err != nil {
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
	builder.WithDpServer(server.NewDpServer(*cfg.DpServer, builder.Metrics()))
	builder.WithKDSContext(kds_context.DefaultContext(appCtx, builder.ResourceManager(), cfg.Multizone.Zone.Name))
	builder.WithInterCPClientPool(intercp.DefaultClientPool())

	if cfg.Mode == config_core.Global {
		kdsEnvoyAdminClient := admin.NewKDSEnvoyAdminClient(
			builder.KDSContext().EnvoyAdminRPCs,
			cfg.Store.Type == store.KubernetesStore,
		)
		forwardingClient := envoyadmin.NewForwardingEnvoyAdminClient(
			builder.ReadOnlyResourceManager(),
			catalog.NewConfigCatalog(builder.ResourceManager()),
			builder.GetInstanceId(),
			intercp.PooledEnvoyAdminClientFn(builder.InterCPClientPool()),
			kdsEnvoyAdminClient,
		)
		builder.WithEnvoyAdminClient(forwardingClient)
	} else {
		builder.WithEnvoyAdminClient(admin.NewEnvoyAdminClient(
			builder.ResourceManager(),
			builder.CaManagers(),
			builder.Config().GetEnvoyAdminPort(),
		))
	}

	xdsCtx, err := xds_runtime.Default(builder) //nolint:contextcheck
	if err != nil {
		return nil, err
	}
	builder.WithXDS(xdsCtx)

	// The setting should be removed, and there is no easy way to set it without breaking most of the code
	mesh_proto.EnableLocalhostInboundClusters = builder.Config().Defaults.EnableLocalhostInboundClusters

	builder.WithAccess(core_runtime.Access{
		ResourceAccess:       resources_access.NewAdminResourceAccess(builder.Config().Access.Static.AdminResources),
		DataplaneTokenAccess: tokens_access.NewStaticGenerateDataplaneTokenAccess(builder.Config().Access.Static.GenerateDPToken),
		ZoneTokenAccess:      zone_access.NewStaticZoneTokenAccess(builder.Config().Access.Static.GenerateZoneToken),
		EnvoyAdminAccess: access.NewStaticEnvoyAdminAccess(
			builder.Config().Access.Static.ViewConfigDump,
			builder.Config().Access.Static.ViewStats,
			builder.Config().Access.Static.ViewClusters,
		),
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
	zoneName := ""
	switch builder.Config().Mode {
	case config_core.Zone:
		zoneName = builder.Config().Multizone.Zone.Name
	case config_core.Global:
		zoneName = "Global"
	case config_core.Standalone:
		zoneName = "Standalone"
	}
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
		switch cfg.Store.Postgres.DriverName {
		case postgres.DriverNamePgx:
			pluginName = core_plugins.Pgx
		case postgres.DriverNamePq:
			pluginName = core_plugins.Postgres
		default:
			return errors.Errorf("unknown driver name %s", cfg.Store.Postgres.DriverName)
		}
		pluginConfig = cfg.Store.Postgres
	default:
		return errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	plugin, err := core_plugins.Plugins().ResourceStore(pluginName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve store %s plugin", pluginName)
	}

	rs, err := plugin.NewResourceStore(builder, pluginConfig)
	if err != nil {
		return err
	}
	builder.WithResourceStore(rs)
	eventBus := events.NewEventBus()
	if err := plugin.EventListener(builder, eventBus); err != nil {
		return err
	}
	builder.WithEventReaderFactory(eventBus)

	paginationStore := core_store.NewPaginationStore(rs)
	meteredStore, err := metrics_store.NewMeteredStore(paginationStore, builder.Metrics())
	if err != nil {
		return err
	}

	builder.WithResourceStore(meteredStore)
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
			cfg.Store.UnsafeDelete,
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
		dataplane.NewDataplaneManager(builder.ResourceStore(), builder.Config().Multizone.Zone.Name, builder.ResourceValidators().Dataplane),
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
	switch cfg.Mode {
	case config_core.Zone:
		secretValidator = secret_manager.ValidateDelete(func(ctx context.Context, secretName string, secretMesh string) error { return nil })
	default:
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
		cachedManager, err := core_manager.NewCachedManager(customizableManager, builder.Config().Store.Cache.ExpirationTime.Duration, builder.Metrics())
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
	meshContextBuilder := xds_context.NewMeshContextBuilder(
		builder.ReadOnlyResourceManager(),
		xds_server.MeshResourceTypes(xds_server.HashMeshExcludedResources),
		builder.LookupIP(),
		builder.Config().Multizone.Zone.Name,
		vips.NewPersistence(builder.ReadOnlyResourceManager(), builder.ConfigManager()),
		builder.Config().DNSServer.Domain,
		builder.Config().DNSServer.ServiceVipPort,
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
	issuers := builtin.TokenIssuers{
		ZoneIngressToken: builtin.NewZoneIngressTokenIssuer(builder.ResourceManager()),
	}
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
