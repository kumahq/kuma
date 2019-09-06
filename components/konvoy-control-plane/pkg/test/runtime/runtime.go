package runtime

import (
	builtin_ca "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/ca/builtin"
	mesh_managers "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/managers/apis/mesh"
	core_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/manager"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	secret_cipher "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/cipher"
	secret_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/manager"
	secret_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/secrets/store"

	konvoy_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-cp"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	bootstrap_universal "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/bootstrap/universal"
	resources_memory "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
)

var _ core_runtime.RuntimeInfo = TestRuntimeInfo{}

type TestRuntimeInfo struct {
	InstanceId string
}

func (i TestRuntimeInfo) GetInstanceId() string {
	return i.InstanceId
}

func BuilderFor(cfg konvoy_cp.Config) *core_runtime.Builder {
	builder := core_runtime.BuilderFor(cfg).
		WithComponentManager(bootstrap_universal.NewComponentManager()).
		WithResourceStore(resources_memory.NewStore()).
		WithXdsContext(core_xds.NewXdsContext())

	builder.
		WithSecretManager(newSecretManager(builder)).
		WithBuiltinCaManager(newBuiltinCaManager(builder)).
		WithResourceManager(newResourceManager(builder))

	return builder
}

func newSecretManager(builder *core_runtime.Builder) secret_manager.SecretManager {
	secretStore := secret_store.NewSecretStore(builder.ResourceStore())
	secretManager := secret_manager.NewSecretManager(secretStore, secret_cipher.None())
	return secretManager
}

func newBuiltinCaManager(builder *core_runtime.Builder) builtin_ca.BuiltinCaManager {
	return builtin_ca.NewBuiltinCaManager(builder.SecretManager())
}

func newResourceManager(builder *core_runtime.Builder) core_manager.ResourceManager {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	meshManager := mesh_managers.NewMeshManager(builder.ResourceStore(), builder.BuiltinCaManager())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{
		core_mesh.MeshType: meshManager,
	}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
	return customizableManager
}
