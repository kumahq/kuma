package bootstrap

import (
	"context"

	"github.com/Kong/kuma/pkg/config/mode"

	config_manager "github.com/Kong/kuma/pkg/core/config/manager"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"

	"github.com/Kong/kuma/pkg/core/managers/apis/dataplane"

	"github.com/Kong/kuma/pkg/clusters/poller"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/dns"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/Kong/kuma/pkg/config/core"
	"github.com/Kong/kuma/pkg/config/core/resources/store"
	"github.com/Kong/kuma/pkg/core/datasource"
	"github.com/Kong/kuma/pkg/core/managers/apis/dataplaneinsight"
	mesh_managers "github.com/Kong/kuma/pkg/core/managers/apis/mesh"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	runtime_reports "github.com/Kong/kuma/pkg/core/runtime/reports"
	secret_cipher "github.com/Kong/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	builtin_issuer "github.com/Kong/kuma/pkg/tokens/builtin/issuer"
)

func buildRuntime(cfg kuma_cp.Config) (core_runtime.Runtime, error) {
	if err := autoconfigure(&cfg); err != nil {
		return nil, err
	}
	builder := core_runtime.BuilderFor(cfg)
	if err := initializeBootstrap(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeResourceStore(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeSecretStore(cfg, builder); err != nil {
		return nil, err
	}
	// we add Secret store to unified ResourceStore so global<->remote synchronizer can use unified interface
	builder.WithResourceStore(core_store.NewCustomizableResourceStore(builder.ResourceStore(), map[core_model.ResourceType]core_store.ResourceStore{
		system.SecretType: builder.SecretStore(),
	}))
	if err := initializeDiscovery(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeClusters(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeConfigManager(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeDNSResolver(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeResourceManager(cfg, builder); err != nil {
		return nil, err
	}

	builder.WithDataSourceLoader(datasource.NewDataSourceLoader(builder.ResourceManager()))

	if err := initializeCaManagers(builder); err != nil {
		return nil, err
	}

	initializeXds(builder)

	leaderInfoComponent := &component.LeaderInfoComponent{}
	builder.WithLeaderInfo(leaderInfoComponent)

	rt, err := builder.Build()
	if err != nil {
		return nil, err
	}

	if err := rt.Add(&component.LeaderInfoComponent{}); err != nil {
		return nil, err
	}

	if err := customizeRuntime(rt); err != nil {
		return nil, err
	}

	return rt, nil
}

func Bootstrap(cfg kuma_cp.Config) (core_runtime.Runtime, error) {
	runtime, err := buildRuntime(cfg)
	if err != nil {
		return nil, err
	}

	if err = onStartup(runtime); err != nil {
		return nil, err
	}

	return runtime, nil
}

func onStartup(runtime core_runtime.Runtime) error {
	if err := createDefaultMesh(runtime); err != nil {
		return err
	}
	if err := createDefaultSigningKey(runtime); err != nil {
		return err
	}
	return startReporter(runtime)
}

func createDefaultSigningKey(runtime core_runtime.Runtime) error {
	switch env := runtime.Config().Environment; env {
	case config_core.KubernetesEnvironment:
		// we use service account token on K8S, so there is no need for dataplane token server
		return nil
	case config_core.UniversalEnvironment:
		return builtin_issuer.CreateDefaultSigningKey(runtime.ResourceManager())
	default:
		return errors.Errorf("unknown environment type %s", env)
	}
}

func createDefaultMesh(runtime core_runtime.Runtime) error {
	switch env := runtime.Config().Environment; env {
	case config_core.KubernetesEnvironment:
		// default Mesh on Kubernetes is managed by a Controller
		return nil
	case config_core.UniversalEnvironment:
		return mesh_managers.CreateDefaultMesh(runtime.ResourceManager(), runtime.Config().Defaults.MeshProto())
	default:
		return errors.Errorf("unknown environment type %s", env)
	}
}

func startReporter(runtime core_runtime.Runtime) error {
	return runtime.Add(component.ComponentFunc(func(stop <-chan struct{}) error {
		runtime_reports.Init(runtime, runtime.Config())
		<-stop
		return nil
	}))
}

func initializeBootstrap(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	switch cfg.Environment {
	case config_core.KubernetesEnvironment:
		pluginName = core_plugins.Kubernetes
	case config_core.UniversalEnvironment:
		pluginName = core_plugins.Universal
	default:
		return errors.Errorf("unknown environment type %s", cfg.Environment)
	}
	plugin, err := core_plugins.Plugins().Bootstrap(pluginName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve bootstrap %s plugin", pluginName)
	}
	if err := plugin.Bootstrap(builder, nil); err != nil {
		return err
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
	if rs, err := plugin.NewResourceStore(builder, pluginConfig); err != nil {
		return err
	} else {
		builder.WithResourceStore(rs)
		return nil
	}
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

func initializeDiscovery(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	var pluginConfig core_plugins.PluginConfig
	switch cfg.Environment {
	case config_core.KubernetesEnvironment:
		pluginName = core_plugins.Kubernetes
		pluginConfig = nil
	case config_core.UniversalEnvironment:
		// there is no discovery mechanism for Universal. Dataplanes are applied via API
		return nil
	default:
		return errors.Errorf("unknown environment type %s", cfg.Environment)
	}
	plugin, err := core_plugins.Plugins().Discovery(pluginName)
	if err != nil {
		return err
	}
	if err := plugin.StartDiscovering(builder, pluginConfig); err != nil {
		return err
	}
	return nil
}

func initializeXds(builder *core_runtime.Builder) {
	builder.WithXdsContext(core_xds.NewXdsContext())
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
	}
	meshManager := mesh_managers.NewMeshManager(builder.ResourceStore(), customizableManager, builder.CaManagers(), registry.Global(), meshValidator)
	customManagers[mesh.MeshType] = meshManager

	dpManager := dataplane.NewDataplaneManager(builder.ResourceStore(), builder.Config().Mode.Remote.Zone)
	customManagers[mesh.DataplaneType] = dpManager

	dpInsightManager := dataplaneinsight.NewDataplaneInsightManager(builder.ResourceStore(), builder.Config().Metrics.Dataplane)
	customManagers[mesh.DataplaneInsightType] = dpInsightManager

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
	switch cfg.Mode.Mode {
	case mode.Remote:
		secretValidator = secret_manager.ValidateDelete(func(ctx context.Context, secretName string, secretMesh string) error { return nil })
	default:
		secretValidator = secret_manager.NewSecretValidator(builder.CaManagers(), builder.ResourceStore())
	}
	secretManager := secret_manager.NewSecretManager(builder.SecretStore(), cipher, secretValidator)
	customManagers[system.SecretType] = secretManager

	builder.WithResourceManager(customizableManager)

	if builder.Config().Store.Cache.Enabled {
		builder.WithReadOnlyResourceManager(core_manager.NewCachedManager(customizableManager, builder.Config().Store.Cache.ExpirationTime))
	} else {
		builder.WithReadOnlyResourceManager(customizableManager)
	}
	return nil
}

func initializeDNSResolver(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	builder.WithDNSResolver(dns.NewDNSResolver(cfg.DNSServer.Domain))
	return nil
}

func initializeClusters(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	poller, err := poller.NewClustersStatusPoller(cfg.Mode.Global)
	if err != nil {
		return err
	}

	builder.WithClusters(poller)
	return nil
}

func initializeConfigManager(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
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
		return errors.Wrapf(err, "could not retrieve config store %s plugin", pluginName)
	}
	if configStore, err := plugin.NewConfigStore(builder, pluginConfig); err != nil {
		return err
	} else {
		builder.WithConfigManager(config_manager.NewConfigManager(configStore))
		return nil
	}
}

func customizeRuntime(rt core_runtime.Runtime) error {
	var pluginName core_plugins.PluginName
	switch env := rt.Config().Environment; env {
	case config_core.KubernetesEnvironment:
		pluginName = core_plugins.Kubernetes
	case config_core.UniversalEnvironment:
		return nil
	default:
		return errors.Errorf("unknown environment type %q", env)
	}
	plugin, err := core_plugins.Plugins().Runtime(pluginName)
	if err != nil {
		return err
	}
	return plugin.Customize(rt)
}
