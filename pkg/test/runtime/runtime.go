package runtime

import (
	builtin_ca "github.com/Kong/kuma/pkg/core/ca/builtin"
	provided_ca "github.com/Kong/kuma/pkg/core/ca/provided"
	mesh_managers "github.com/Kong/kuma/pkg/core/managers/apis/mesh"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	secret_cipher "github.com/Kong/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	secret_store "github.com/Kong/kuma/pkg/core/secrets/store"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	bootstrap_universal "github.com/Kong/kuma/pkg/plugins/bootstrap/universal"
	resources_memory "github.com/Kong/kuma/pkg/plugins/resources/memory"
)

var _ core_runtime.RuntimeInfo = TestRuntimeInfo{}

type TestRuntimeInfo struct {
	InstanceId string
}

func (i TestRuntimeInfo) GetInstanceId() string {
	return i.InstanceId
}

func BuilderFor(cfg kuma_cp.Config) *core_runtime.Builder {
	builder := core_runtime.BuilderFor(cfg).
		WithComponentManager(bootstrap_universal.NewComponentManager()).
		WithResourceStore(resources_memory.NewStore()).
		WithXdsContext(core_xds.NewXdsContext())

	builder.
		WithSecretManager(newSecretManager(builder)).
		WithBuiltinCaManager(newBuiltinCaManager(builder)).
		WithProvidedCaManager(newProvidedCaManager(builder)).
		WithResourceManager(newResourceManager(builder))

	return builder
}

func newSecretManager(builder *core_runtime.Builder) secret_manager.SecretManager {
	secretStore := secret_store.NewSecretStore(builder.ResourceStore())
	secretManager := secret_manager.NewSecretManager(secretStore, secret_cipher.None())
	return secretManager
}

func newProvidedCaManager(builder *core_runtime.Builder) provided_ca.ProvidedCaManager {
	return provided_ca.NewProvidedCaManager(builder.SecretManager())
}

func newBuiltinCaManager(builder *core_runtime.Builder) builtin_ca.BuiltinCaManager {
	return builtin_ca.NewBuiltinCaManager(builder.SecretManager())
}

func newResourceManager(builder *core_runtime.Builder) core_manager.ResourceManager {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
	meshManager := mesh_managers.NewMeshManager(builder.ResourceStore(), builder.BuiltinCaManager(), builder.ProvidedCaManager(), customizableManager, builder.SecretManager(), registry.Global())
	customManagers[core_mesh.MeshType] = meshManager
	return customizableManager
}
