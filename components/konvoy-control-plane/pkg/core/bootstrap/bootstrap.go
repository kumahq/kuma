package bootstrap

import (
	"context"

	kuma_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/kuma-cp"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	builtin_ca "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/ca/builtin"
	mesh_managers "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/managers/apis/mesh"
	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/manager"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	secret_cipher "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/cipher"
	secret_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/manager"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	"github.com/pkg/errors"
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

	initializeBuiltinCaManager(builder)

	initializeResourceManager(builder)

	initializeXds(builder)

	return builder.Build()
}

func createDefaultMesh(runtime core_runtime.Runtime) error {
	resManager := runtime.ResourceManager()
	defaultMesh := mesh.MeshResource{}
	cfg := runtime.Config()

	namespace := core_model.DefaultNamespace
	if runtime.Config().Environment == kuma_cp.KubernetesEnvironment {
		namespace = runtime.Config().Store.Kubernetes.SystemNamespace
	}

	key := core_model.ResourceKey{Namespace: namespace, Mesh: core_model.DefaultMesh, Name: core_model.DefaultMesh}

	if err := resManager.Get(context.Background(), &defaultMesh, core_store.GetBy(key)); err != nil {
		if core_store.IsResourceNotFound(err) {
			core.Log.Info("Creating default mesh from the settings", "mesh", cfg.Defaults.Mesh)
			defaultMesh.Spec = cfg.Defaults.Mesh

			if err := resManager.Create(context.Background(), &defaultMesh, core_store.CreateBy(key)); err != nil {
				return errors.Wrapf(err, "Failed to create `default` Mesh resource in a given resource store")
			}
		} else {
			return err
		}
	}

	return nil
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
	return runtime.Add(core_runtime.ComponentFunc(func(stop <-chan struct{}) error {
		if err := createDefaultMesh(runtime); err != nil {
			return err
		}
		<-stop // it has to block, otherwise the k8s component manager stops all other components
		return nil
	}))
}

func initializeBootstrap(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	var pluginName core_plugins.PluginName
	switch cfg.Environment {
	case kuma_cp.KubernetesEnvironment:
		pluginName = core_plugins.Kubernetes
	case kuma_cp.UniversalEnvironment:
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
	case kuma_cp.KubernetesEnvironment:
		pluginName = core_plugins.Kubernetes
		pluginConfig = nil
	case kuma_cp.UniversalEnvironment:
		pluginName = core_plugins.Universal
		pluginConfig = cfg.Discovery.Universal
	default:
		return errors.Errorf("unknown environment type %s", cfg.Environment)
	}
	plugin, err := core_plugins.Plugins().Discovery(pluginName)
	if err != nil {
		return err
	}
	if source, err := plugin.NewDiscoverySource(builder, pluginConfig); err != nil {
		return err
	} else {
		builder.AddDiscoverySource(source)
	}
	return nil
}

func initializeXds(builder *core_runtime.Builder) {
	builder.WithXdsContext(core_xds.NewXdsContext())
}

func initializeBuiltinCaManager(builder *core_runtime.Builder) {
	builder.WithBuiltinCaManager(builtin_ca.NewBuiltinCaManager(builder.SecretManager()))
}

func initializeResourceManager(builder *core_runtime.Builder) {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	meshManager := mesh_managers.NewMeshManager(builder.ResourceStore(), builder.BuiltinCaManager())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{
		mesh.MeshType: meshManager,
	}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
	builder.WithResourceManager(customizableManager)
}
