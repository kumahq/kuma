package runtime

import (
	"net"

	"github.com/kumahq/kuma/pkg/api-server/customization"
	"github.com/kumahq/kuma/pkg/dp-server/server"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	mesh_managers "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	secret_cipher "github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/dns/resolver"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/ca/builtin"
	leader_memory "github.com/kumahq/kuma/pkg/plugins/leader/memory"
	resources_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ core_runtime.RuntimeInfo = TestRuntimeInfo{}

type TestRuntimeInfo struct {
	InstanceId string
	ClusterId  string
}

func (i TestRuntimeInfo) GetInstanceId() string {
	return i.InstanceId
}

func (i TestRuntimeInfo) SetClusterId(clusterId string) {
	i.ClusterId = clusterId
}

func (i TestRuntimeInfo) GetClusterId() string {
	return i.ClusterId
}

func BuilderFor(cfg kuma_cp.Config) (*core_runtime.Builder, error) {
	stopCh := make(chan struct{})
	builder, err := core_runtime.BuilderFor(cfg, stopCh)
	if err != nil {
		return nil, err
	}
	builder.
		WithComponentManager(component.NewManager(leader_memory.NewAlwaysLeaderElector())).
		WithResourceStore(resources_memory.NewStore())

	metrics, _ := metrics.NewMetrics("Standalone")
	builder.WithMetrics(metrics)

	builder.WithSecretStore(secret_store.NewSecretStore(builder.ResourceStore()))
	builder.WithDataSourceLoader(datasource.NewDataSourceLoader(builder.ResourceManager()))

	rm := newResourceManager(builder)
	builder.WithResourceManager(rm).
		WithReadOnlyResourceManager(rm)

	builder.WithCaManager("builtin", builtin.NewBuiltinCaManager(builder.ResourceManager()))
	builder.WithLeaderInfo(&component.LeaderInfoComponent{})
	builder.WithLookupIP(net.LookupIP)
	builder.WithEnvoyAdminClient(&DummyEnvoyAdminClient{})
	builder.WithEventReaderFactory(events.NewEventBus())
	builder.WithAPIManager(customization.NewAPIList())
	builder.WithXDSHooks(&xds_hooks.Hooks{})
	builder.WithDpServer(server.NewDpServer(*cfg.DpServer, metrics))
	builder.WithKDSContext(kds_context.DefaultContext(builder.ResourceManager(), cfg.Multizone.Remote.Zone))

	_ = initializeConfigManager(cfg, builder)
	_ = initializeDNSResolver(cfg, builder)

	return builder, nil
}

func initializeConfigManager(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	configm := config_manager.NewConfigManager(builder.ResourceStore())
	builder.WithConfigManager(configm)
	return nil
}

func initializeDNSResolver(cfg kuma_cp.Config, builder *core_runtime.Builder) error {
	builder.WithDNSResolver(resolver.NewDNSResolver("mesh"))
	return nil
}

func newResourceManager(builder *core_runtime.Builder) core_manager.ResourceManager {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
	validator := mesh_managers.MeshValidator{
		CaManagers: builder.CaManagers(),
	}
	meshManager := mesh_managers.NewMeshManager(builder.ResourceStore(), customizableManager, builder.CaManagers(), registry.Global(), validator)
	customManagers[core_mesh.MeshType] = meshManager

	secretManager := secret_manager.NewSecretManager(builder.SecretStore(), secret_cipher.None(), nil)
	customManagers[system.SecretType] = secretManager
	return customizableManager
}

type DummyEnvoyAdminClient struct {
	PostQuitCalled *int
}

func (d *DummyEnvoyAdminClient) GenerateAPIToken(dp *mesh_core.DataplaneResource) (string, error) {
	return "token", nil
}

func (d *DummyEnvoyAdminClient) PostQuit(dataplane *mesh_core.DataplaneResource) error {
	if d.PostQuitCalled != nil {
		*d.PostQuitCalled++
	}

	return nil
}
