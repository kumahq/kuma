package bootstrap

import (
	"context"
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
	"github.com/kumahq/kuma/pkg/dp-server/server"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/envoy/admin/access"
	"github.com/kumahq/kuma/pkg/events"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	"github.com/kumahq/kuma/pkg/metrics"
	metrics_store "github.com/kumahq/kuma/pkg/metrics/store"
	tokens_access "github.com/kumahq/kuma/pkg/tokens/builtin/access"
	zone_access "github.com/kumahq/kuma/pkg/tokens/builtin/zone/access"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

var log = core.Log.WithName("bootstrap")

func buildRuntime(appCtx context.Context, cfg kuma_cp.Config) (core_runtime.Runtime, error) {
	if err := autoconfigure(&cfg); err != nil {
		return nil, err
	}
	runtime, err := core_runtime.NewRuntime(appCtx, cfg)
	if err != nil {
		return nil, err
	}
	if err := core_runtime.ApplyOpts(runtime, initializeMetrics); err != nil {
		return nil, err
	}
	for _, plugin := range core_plugins.Plugins().BootstrapPlugins() {
		if err := plugin.BeforeBootstrap(runtime, cfg); err != nil {
			return nil, errors.Wrapf(err, "failed to run beforeBootstrap plugin:'%s'", plugin.Name())
		}
	}
	leaderInfoComponent := &component.LeaderInfoComponent{}
	err = core_runtime.ApplyOpts(runtime,
		initializeResourceStore,
		initializeSecretStore,
		initializeConfigStore,
		initializeResourceStore,
		initializeSecretStore,
		initializeConfigStore)
	if err != nil {
		return nil, err
	}
	err = core_runtime.ApplyOpts(runtime,
		// we add Secret store to unified ResourceStore so global<->zone synchronizer can use unified interface
		core_runtime.WithResourceStore(core_store.NewCustomizableResourceStore(runtime.ResourceStore(), map[core_model.ResourceType]core_store.ResourceStore{
			system.SecretType:       runtime.SecretStore(),
			system.GlobalSecretType: runtime.SecretStore(),
			system.ConfigType:       runtime.ConfigStore(),
		})),
		core_runtime.WithConfigManager(config_manager.NewConfigManager(runtime.ConfigStore())),
		core_runtime.WithResourceValidators(core_runtime.ResourceValidators{
			Dataplane: dataplane.NewMembershipValidator(),
			Mesh:      mesh_managers.NewMeshValidator(runtime.CaManagers(), runtime.ResourceStore()),
		}),
		initializeResourceManager,
		core_runtime.WithDataSourceLoader(datasource.NewDataSourceLoader(runtime.ReadOnlyResourceManager())),
		initializeCaManagers,
		core_runtime.WithLeaderInfo(leaderInfoComponent),
		core_runtime.WithLookupIP(lookup.CachedLookupIP(net.LookupIP, cfg.General.DNSCacheTTL)),
		core_runtime.WithAPIManager(customization.NewAPIList()),
		initializeCaProvider,
		core_runtime.WithDpServer(server.NewDpServer(*cfg.DpServer, runtime.Metrics())),
		core_runtime.WithKDSContext(kds_context.DefaultContext(appCtx, runtime.ResourceManager(), cfg.Multizone.Zone.Name)),
		intializeEnvoyAdminClient,
		core_runtime.WithAccess(core_runtime.Access{
			ResourceAccess:       resources_access.NewAdminResourceAccess(runtime.Config().Access.Static.AdminResources),
			DataplaneTokenAccess: tokens_access.NewStaticGenerateDataplaneTokenAccess(runtime.Config().Access.Static.GenerateDPToken),
			ZoneTokenAccess:      zone_access.NewStaticZoneTokenAccess(runtime.Config().Access.Static.GenerateZoneToken),
			EnvoyAdminAccess: access.NewStaticEnvoyAdminAccess(
				runtime.Config().Access.Static.ViewConfigDump,
				runtime.Config().Access.Static.ViewStats,
				runtime.Config().Access.Static.ViewClusters,
			),
		}),
		initializeAPIServerAuthenticator,
	)
	if err != nil {
		return nil, err
	}

	// The setting should be removed, and there is no easy way to set it without breaking most of the code
	mesh_proto.EnableLocalhostInboundClusters = runtime.Config().Defaults.EnableLocalhostInboundClusters

	for _, plugin := range core_plugins.Plugins().BootstrapPlugins() {
		if err := plugin.AfterBootstrap(runtime, cfg); err != nil {
			return nil, errors.Wrapf(err, "failed to run afterBootstrap plugin:'%s'", plugin.Name())
		}
	}
	if err := core_runtime.ValidateRuntime(runtime); err != nil {
		return nil, err
	}

	if err := runtime.Add(leaderInfoComponent); err != nil {
		return nil, err
	}

	for name, plugin := range core_plugins.Plugins().RuntimePlugins() {
		if err := plugin.Customize(runtime); err != nil {
			return nil, errors.Wrapf(err, "failed to configure runtime plugin:'%s'", name)
		}
	}

	logWarnings(runtime.Config())

	return runtime, nil
}

func intializeEnvoyAdminClient(r core_runtime.Runtime) error {
	var adminClient admin.EnvoyAdminClient
	if r.Config().Mode == config_core.Global {
		adminClient = admin.NewKDSEnvoyAdminClient(
			r.KDSContext().EnvoyAdminRPCs,
			r.Config().Store.Type == store.KubernetesStore,
		)
	} else {
		adminClient = admin.NewEnvoyAdminClient(
			r.ResourceManager(),
			r.CaManagers(),
			r.Config().GetEnvoyAdminPort(),
		)
	}
	return core_runtime.WithEnvoyAdminClient(adminClient)(r)
}

func initializeCaProvider(r core_runtime.Runtime) error {
	caProvider, err := secrets.NewCaProvider(r.CaManagers(), r.Metrics())
	if err != nil {
		return err
	}
	return core_runtime.WithCAProvider(caProvider)(r)
}

func logWarnings(config kuma_cp.Config) {
	if config.ApiServer.Authn.LocalhostIsAdmin {
		log.Info("WARNING: you can access Control Plane API as admin by sending requests from the same machine where Control Plane runs. To increase security, it is recommended to extract admin credentials and set KUMA_API_SERVER_AUTHN_LOCALHOST_IS_ADMIN to false.")
	}
}

func initializeMetrics(r core_runtime.Runtime) error {
	zoneName := ""
	switch r.Config().Mode {
	case config_core.Zone:
		zoneName = r.Config().Multizone.Zone.Name
	case config_core.Global:
		zoneName = "Global"
	case config_core.Standalone:
		zoneName = "Standalone"
	}
	m, err := metrics.NewMetrics(zoneName)
	if err != nil {
		return err
	}
	return core_runtime.WithMetrics(m)(r)
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

func initializeResourceStore(r core_runtime.Runtime) error {
	cfg := r.Config()
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

	rs, err := plugin.NewResourceStore(r, pluginConfig)
	if err != nil {
		return err
	}
	eventBus := events.NewEventBus()
	if err := plugin.EventListener(r, eventBus); err != nil {
		return err
	}

	paginationStore := core_store.NewPaginationStore(rs)
	meteredStore, err := metrics_store.NewMeteredStore(paginationStore, r.Metrics())
	if err != nil {
		return err
	}

	if err := core_runtime.WithEventReaderFactory(eventBus)(r); err != nil {
		return err
	}
	if err := core_runtime.WithResourceStore(meteredStore)(r); err != nil {
		return err
	}
	return nil
}

func initializeSecretStore(r core_runtime.Runtime) error {
	var pluginName core_plugins.PluginName
	var pluginConfig core_plugins.PluginConfig
	cfg := r.Config()
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
	ss, err := plugin.NewSecretStore(r, pluginConfig)
	if err != nil {
		return err
	}
	return core_runtime.WithSecretStore(ss)(r)
}

func initializeConfigStore(r core_runtime.Runtime) error {
	cfg := r.Config()
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
	cs, err := plugin.NewConfigStore(r, pluginConfig)
	if err != nil {
		return err
	}
	return core_runtime.WithConfigStore(cs)(r)
}

func initializeCaManagers(r core_runtime.Runtime) error {
	for pluginName, caPlugin := range core_plugins.Plugins().CaPlugins() {
		caManager, err := caPlugin.NewCaManager(r, nil)
		if err != nil {
			return errors.Wrapf(err, "could not create CA manager for plugin %q", pluginName)
		}
		if err = core_runtime.WithCaManager(string(pluginName), caManager)(r); err != nil {
			return errors.Wrapf(err, "could not create CA manager for plugin %q", pluginName)
		}
	}
	return nil
}

func initializeAPIServerAuthenticator(r core_runtime.Runtime) error {
	authnType := r.Config().ApiServer.Authn.Type
	plugin, ok := core_plugins.Plugins().AuthnAPIServer()[core_plugins.PluginName(authnType)]
	if !ok {
		return errors.Errorf("there is not implementation of authn named %s", authnType)
	}
	authenticator, err := plugin.NewAuthenticator(r)
	if err != nil {
		return errors.Wrapf(err, "could not initiate authenticator %s", authnType)
	}
	return core_runtime.WithAPIServerAuthenticator(authenticator)(r)
}

func initializeResourceManager(r core_runtime.Runtime) error {
	defaultManager := core_manager.NewResourceManager(r.ResourceStore())
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, nil)
	cfg := r.Config()

	customizableManager.Customize(
		mesh.MeshType,
		mesh_managers.NewMeshManager(
			r.ResourceStore(),
			customizableManager,
			r.CaManagers(),
			registry.Global(),
			r.ResourceValidators().Mesh,
			cfg.Store.UnsafeDelete,
		),
	)

	rateLimitValidator := ratelimit_managers.RateLimitValidator{
		Store: r.ResourceStore(),
	}
	customizableManager.Customize(
		mesh.RateLimitType,
		ratelimit_managers.NewRateLimitManager(r.ResourceStore(), rateLimitValidator),
	)

	externalServiceValidator := externalservice_managers.ExternalServiceValidator{
		Store: r.ResourceStore(),
	}
	customizableManager.Customize(
		mesh.ExternalServiceType,
		externalservice_managers.NewExternalServiceManager(r.ResourceStore(), externalServiceValidator),
	)

	customizableManager.Customize(
		mesh.DataplaneType,
		dataplane.NewDataplaneManager(r.ResourceStore(), cfg.Multizone.Zone.Name, r.ResourceValidators().Dataplane),
	)

	customizableManager.Customize(
		mesh.DataplaneInsightType,
		dataplaneinsight.NewDataplaneInsightManager(r.ResourceStore(), cfg.Metrics.Dataplane),
	)

	customizableManager.Customize(
		system.ZoneType,
		zone.NewZoneManager(r.ResourceStore(), zone.Validator{Store: r.ResourceStore()}, cfg.Store.UnsafeDelete),
	)

	customizableManager.Customize(
		system.ZoneInsightType,
		zoneinsight.NewZoneInsightManager(r.ResourceStore(), cfg.Metrics.Zone),
	)

	customizableManager.Customize(
		mesh.ZoneIngressInsightType,
		zoneingressinsight.NewZoneIngressInsightManager(r.ResourceStore(), cfg.Metrics.Dataplane),
	)

	customizableManager.Customize(
		mesh.ZoneEgressInsightType,
		zoneegressinsight.NewZoneEgressInsightManager(r.ResourceStore(), cfg.Metrics.Dataplane),
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
		secretValidator = secret_manager.NewSecretValidator(r.CaManagers(), r.ResourceStore())
	}

	customizableManager.Customize(
		system.SecretType,
		secret_manager.NewSecretManager(r.SecretStore(), cipher, secretValidator, cfg.Store.UnsafeDelete),
	)

	customizableManager.Customize(
		system.GlobalSecretType,
		secret_manager.NewGlobalSecretManager(r.SecretStore(), cipher),
	)

	if err := core_runtime.WithResourceManager(customizableManager)(r); err != nil {
		return err
	}

	opt := core_runtime.WithReadOnlyResourceManager(customizableManager)
	if cfg.Store.Cache.Enabled {
		cachedManager, err := core_manager.NewCachedManager(customizableManager, cfg.Store.Cache.ExpirationTime, r.Metrics())
		if err != nil {
			return err
		}
		opt = core_runtime.WithReadOnlyResourceManager(cachedManager)
	}
	return opt(r)
}
