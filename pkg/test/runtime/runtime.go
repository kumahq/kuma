package runtime

import (
	config_manager "github.com/Kong/kuma/pkg/core/config/manager"
	config_store "github.com/Kong/kuma/pkg/core/config/store"
	"github.com/Kong/kuma/pkg/dns"

	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/core/datasource"
	mesh_managers "github.com/Kong/kuma/pkg/core/managers/apis/mesh"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	secret_cipher "github.com/Kong/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	secret_store "github.com/Kong/kuma/pkg/core/secrets/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/plugins/ca/builtin"
	leader_memory "github.com/Kong/kuma/pkg/plugins/leader/memory"
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
		WithComponentManager(component.NewManager(leader_memory.NewAlwaysLeaderElector())).
		WithResourceStore(resources_memory.NewStore()).
		WithXdsContext(core_xds.NewXdsContext())

	builder.WithDataSourceLoader(datasource.NewDataSourceLoader(builder.SecretManager()))
	builder.WithSecretManager(newSecretManager(builder))

	rm := newResourceManager(builder)
	builder.WithResourceManager(rm).
		WithReadOnlyResourceManager(rm)

	builder.WithCaManager("builtin", builtin.NewBuiltinCaManager(builder.SecretManager()))
	builder.WithLeaderInfo(&component.LeaderInfoComponent{})

	_ = initializeConfigManager(cfg, builder)
	_ = initializeDNSResolver(cfg, builder)

	return builder
}

func initializeConfigManager(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	store := config_store.NewConfigStore(builder.ResourceStore())
	configm := config_manager.NewConfigManager(store)
	builder.WithConfigManager(configm)
	return nil
}

func initializeDNSResolver(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	builder.WithDNSResolver(dns.NewDNSResolver("mesh"))
	return nil
}

func newSecretManager(builder *core_runtime.Builder) secret_manager.SecretManager {
	secretStore := secret_store.NewSecretStore(builder.ResourceStore())
	secretManager := secret_manager.NewSecretManager(secretStore, secret_cipher.None(), nil)
	return secretManager
}

func newResourceManager(builder *core_runtime.Builder) core_manager.ResourceManager {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
	validator := mesh_managers.MeshValidator{
		CaManagers: builder.CaManagers(),
	}
	meshManager := mesh_managers.NewMeshManager(builder.ResourceStore(), customizableManager, builder.SecretManager(), builder.CaManagers(), registry.Global(), validator)
	customManagers[core_mesh.MeshType] = meshManager
	return customizableManager
}
