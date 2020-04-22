package bootstrap

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core/datasource"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/Kong/kuma/pkg/config/core"
	"github.com/Kong/kuma/pkg/config/core/resources/store"
	mesh_managers "github.com/Kong/kuma/pkg/core/managers/apis/mesh"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
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
	if err := initializeSecretManager(cfg, builder); err != nil {
		return nil, err
	}
	if err := initializeDiscovery(cfg, builder); err != nil {
		return nil, err
	}

	initializeResourceManager(builder)

	builder.WithDataSourceLoader(datasource.NewDataSourceLoader(builder.SecretManager()))

	if err := initializeCaManagers(builder); err != nil {
		return nil, err
	}

	initializeXds(builder)

	rt, err := builder.Build()
	if err != nil {
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
		return builtin_issuer.CreateDefaultSigningKey(runtime.SecretManager())
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

func initializeSecretManager(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	var pluginConfig core_plugins.PluginConfig
	var cipher secret_cipher.Cipher
	switch cfg.Store.Type {
	case store.KubernetesStore:
		pluginName = core_plugins.Kubernetes
		cipher = secret_cipher.None() // deliberately turn encryption off on Kubernetes
	case store.MemoryStore, store.PostgresStore:
		pluginName = core_plugins.Universal
		cipher = secret_cipher.TODO() // get back to encryption in universal case
	default:
		return errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	plugin, err := core_plugins.Plugins().SecretStore(pluginName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve secret store %s plugin", pluginName)
	}
	if secretStore, err := plugin.NewSecretStore(builder, pluginConfig); err != nil {
		return err
	} else {
		builder.WithSecretManager(secret_manager.NewSecretManager(secretStore, cipher))
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

func initializeResourceManager(builder *core_runtime.Builder) {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)

	validator := mesh_managers.MeshValidator{
		CaManagers: builder.CaManagers(),
	}
	meshManager := mesh_managers.NewMeshManager(builder.ResourceStore(), customizableManager, builder.SecretManager(), builder.CaManagers(), registry.Global(), validator)
	customManagers[mesh.MeshType] = meshManager
	builder.WithResourceManager(customizableManager)

	if builder.Config().Store.Cache.Enabled {
		builder.WithReadOnlyResourceManager(core_manager.NewCachedManager(customizableManager, builder.Config().Store.Cache.ExpirationTime))
	} else {
		builder.WithReadOnlyResourceManager(customizableManager)
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
