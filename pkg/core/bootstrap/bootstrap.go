package bootstrap

import (
	"context"
	"net"

	"github.com/kumahq/kuma/pkg/envoy/admin"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"

	"github.com/kumahq/kuma/pkg/api-server/customization"
	"github.com/kumahq/kuma/pkg/core/managers/apis/zone"
	"github.com/kumahq/kuma/pkg/dp-server/server"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/dns/resolver"

	metrics_store "github.com/kumahq/kuma/pkg/metrics/store"

	"github.com/kumahq/kuma/pkg/events"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	"github.com/kumahq/kuma/pkg/core/managers/apis/dataplaneinsight"
	mesh_managers "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/managers/apis/zoneinsight"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
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
	"github.com/kumahq/kuma/pkg/metrics"
)

func buildRuntime(cfg kuma_cp.Config, closeCh <-chan struct{}) (core_runtime.Runtime, error) {
	if err := autoconfigure(&cfg); err != nil {
		return nil, err
	}
	builder, err := core_runtime.BuilderFor(cfg, closeCh)
	if err != nil {
		return nil, err
	}
	if err := initializeMetrics(builder); err != nil {
		return nil, err
	}
	if err := initializeBeforeBootstrap(cfg, builder); err != nil {
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
	// we add Secret store to unified ResourceStore so global<->remote synchronizer can use unified interface
	builder.WithResourceStore(core_store.NewCustomizableResourceStore(builder.ResourceStore(), map[core_model.ResourceType]core_store.ResourceStore{
		system.SecretType: builder.SecretStore(),
		system.ConfigType: builder.ConfigStore(),
	}))

	if err := initializeConfigManager(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeDNSResolver(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeResourceManager(cfg, builder); err != nil {
		return nil, err
	}

	builder.WithDataSourceLoader(datasource.NewDataSourceLoader(builder.ReadOnlyResourceManager()))

	if err := initializeCaManagers(builder); err != nil {
		return nil, err
	}

	leaderInfoComponent := &component.LeaderInfoComponent{}
	builder.WithLeaderInfo(leaderInfoComponent)

	builder.WithLookupIP(lookup.CachedLookupIP(net.LookupIP, cfg.General.DNSCacheTTL))
	builder.WithEnvoyAdminClient(admin.NewEnvoyAdminClient(builder.ResourceManager(), builder.Config()))
	builder.WithAPIManager(customization.NewAPIList())
	builder.WithXDSHooks(&xds_hooks.Hooks{})
	builder.WithDpServer(server.NewDpServer(*cfg.DpServer, builder.Metrics()))
	builder.WithKDSContext(kds_context.DefaultContext(builder.ResourceManager(), cfg.Multizone.Remote.Zone))

	if err := initializeAfterBootstrap(cfg, builder); err != nil {
		return nil, err
	}

	rt, err := builder.Build()
	if err != nil {
		return nil, err
	}

	if err := rt.Add(leaderInfoComponent); err != nil {
		return nil, err
	}

	if err := customizeRuntime(rt); err != nil {
		return nil, err
	}

	return rt, nil
}

func initializeMetrics(builder *core_runtime.Builder) error {
	zone := ""
	switch builder.Config().Mode {
	case config_core.Remote:
		zone = builder.Config().Multizone.Remote.Zone
	case config_core.Global:
		zone = "Global"
	case config_core.Standalone:
		zone = "Standalone"
	}
	metrics, err := metrics.NewMetrics(zone)
	if err != nil {
		return err
	}
	builder.WithMetrics(metrics)
	return nil
}

func Bootstrap(cfg kuma_cp.Config, closeCh <-chan struct{}) (core_runtime.Runtime, error) {
	runtime, err := buildRuntime(cfg, closeCh)
	if err != nil {
		return nil, err
	}

	if err = startReporter(runtime); err != nil {
		return nil, err
	}

	return runtime, nil
}

func startReporter(runtime core_runtime.Runtime) error {
	return runtime.Add(component.ComponentFunc(func(stop <-chan struct{}) error {
		runtime_reports.Init(runtime, runtime.Config())
		<-stop
		return nil
	}))
}

func initializeBeforeBootstrap(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	for name, plugin := range core_plugins.Plugins().BootstrapPlugins() {
		if (cfg.Environment == config_core.KubernetesEnvironment && name == core_plugins.Universal) ||
			(cfg.Environment == config_core.UniversalEnvironment && name == core_plugins.Kubernetes) {
			continue
		}
		if err := plugin.BeforeBootstrap(builder, nil); err != nil {
			return err
		}
	}
	return nil
}

func initializeAfterBootstrap(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	for name, plugin := range core_plugins.Plugins().BootstrapPlugins() {
		if (cfg.Environment == config_core.KubernetesEnvironment && name == core_plugins.Universal) ||
			(cfg.Environment == config_core.UniversalEnvironment && name == core_plugins.Kubernetes) {
			continue
		}
		if err := plugin.AfterBootstrap(builder, nil); err != nil {
			return err
		}
	}
	return nil
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

func initializeResourceManager(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)

	meshValidator := mesh_managers.MeshValidator{
		CaManagers: builder.CaManagers(),
		Store:      builder.ResourceStore(),
	}
	meshManager := mesh_managers.NewMeshManager(builder.ResourceStore(), customizableManager, builder.CaManagers(), registry.Global(), meshValidator)
	customManagers[mesh.MeshType] = meshManager

	dpManager := dataplane.NewDataplaneManager(builder.ResourceStore(), builder.Config().Multizone.Remote.Zone)
	customManagers[mesh.DataplaneType] = dpManager

	dpInsightManager := dataplaneinsight.NewDataplaneInsightManager(builder.ResourceStore(), builder.Config().Metrics.Dataplane)
	customManagers[mesh.DataplaneInsightType] = dpInsightManager

	zoneValidator := zone.Validator{Store: builder.ResourceStore()}
	zoneManager := zone.NewZoneManager(builder.ResourceStore(), zoneValidator)
	customManagers[system.ZoneType] = zoneManager

	zoneInsightManager := zoneinsight.NewZoneInsightManager(builder.ResourceStore(), builder.Config().Metrics.Zone)
	customManagers[system.ZoneInsightType] = zoneInsightManager

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
	case config_core.Remote:
		secretValidator = secret_manager.ValidateDelete(func(ctx context.Context, secretName string, secretMesh string) error { return nil })
	default:
		secretValidator = secret_manager.NewSecretValidator(builder.CaManagers(), builder.ResourceStore())
	}
	secretManager := secret_manager.NewSecretManager(builder.SecretStore(), cipher, secretValidator)
	customManagers[system.SecretType] = secretManager
	customManagers[system.GlobalSecretType] = secret_manager.NewGlobalSecretManager(builder.SecretStore(), cipher)

	builder.WithResourceManager(customizableManager)

	if builder.Config().Store.Cache.Enabled {
		cachedManager, err := core_manager.NewCachedManager(customizableManager, builder.Config().Store.Cache.ExpirationTime, builder.Metrics())
		if err != nil {
			return err
		}
		builder.WithReadOnlyResourceManager(cachedManager)
	} else {
		builder.WithReadOnlyResourceManager(customizableManager)
	}
	return nil
}

func initializeDNSResolver(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	builder.WithDNSResolver(resolver.NewDNSResolver(cfg.DNSServer.Domain))
	return nil
}

func initializeConfigManager(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	builder.WithConfigManager(config_manager.NewConfigManager(builder.ConfigStore()))
	return nil
}

func customizeRuntime(rt core_runtime.Runtime) error {
	env := rt.Config().Environment
	for name, plugin := range core_plugins.Plugins().RuntimePlugins() {
		if (env == config_core.KubernetesEnvironment && name == core_plugins.Universal) ||
			(env == config_core.UniversalEnvironment && name == core_plugins.Kubernetes) {
			continue
		}
		if err := plugin.Customize(rt); err != nil {
			return err
		}
	}
	return nil
}
