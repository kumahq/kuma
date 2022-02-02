package runtime

import (
	"context"
	"fmt"
	"net"

	"github.com/kumahq/kuma/pkg/api-server/customization"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	mesh_managers "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	resources_access "github.com/kumahq/kuma/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
	"github.com/kumahq/kuma/pkg/dp-server/server"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/events"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/certs"
	"github.com/kumahq/kuma/pkg/plugins/ca/builtin"
	leader_memory "github.com/kumahq/kuma/pkg/plugins/leader/memory"
	resources_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	tokens_access "github.com/kumahq/kuma/pkg/tokens/builtin/access"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
	"github.com/kumahq/kuma/pkg/xds/secrets"
)

var _ core_runtime.RuntimeInfo = &TestRuntimeInfo{}

type TestRuntimeInfo struct {
	InstanceId string
	ClusterId  string
}

func (i *TestRuntimeInfo) GetInstanceId() string {
	return i.InstanceId
}

func (i *TestRuntimeInfo) SetClusterId(clusterId string) {
	i.ClusterId = clusterId
}

func (i *TestRuntimeInfo) GetClusterId() string {
	return i.ClusterId
}

func BuilderFor(appCtx context.Context, cfg kuma_cp.Config) (*core_runtime.Builder, error) {
	builder, err := core_runtime.BuilderFor(appCtx, cfg)
	if err != nil {
		return nil, err
	}

	builder.
		WithComponentManager(component.NewManager(leader_memory.NewAlwaysLeaderElector())).
		WithResourceStore(resources_memory.NewStore()).
		WithSecretStore(secret_store.NewSecretStore(builder.ResourceStore())).
		WithResourceValidators(core_runtime.ResourceValidators{
			Dataplane: dataplane.NewMembershipValidator(),
			Mesh:      mesh_managers.NewMeshValidator(builder.CaManagers(), builder.ResourceStore()),
		})

	rm := newResourceManager(builder)
	builder.WithResourceManager(rm).
		WithReadOnlyResourceManager(rm)

	metrics, _ := metrics.NewMetrics("Standalone")
	builder.WithMetrics(metrics)

	builder.WithDataSourceLoader(datasource.NewDataSourceLoader(builder.ResourceManager()))
	builder.WithCaManager("builtin", builtin.NewBuiltinCaManager(builder.ResourceManager()))
	builder.WithLeaderInfo(&component.LeaderInfoComponent{})
	builder.WithLookupIP(net.LookupIP)
	builder.WithEnvoyAdminClient(&DummyEnvoyAdminClient{})
	builder.WithEventReaderFactory(events.NewEventBus())
	builder.WithAPIManager(customization.NewAPIList())
	builder.WithXDSHooks(&xds_hooks.Hooks{})
	builder.WithDpServer(server.NewDpServer(*cfg.DpServer, metrics))
	builder.WithKDSContext(kds_context.DefaultContext(builder.ResourceManager(), cfg.Multizone.Zone.Name))
	builder.WithCAProvider(secrets.NewCaProvider(builder.CaManagers()))
	builder.WithAPIServerAuthenticator(certs.ClientCertAuthenticator)
	builder.WithAccess(core_runtime.Access{
		ResourceAccess:       resources_access.NewAdminResourceAccess(builder.Config().Access.Static.AdminResources),
		DataplaneTokenAccess: tokens_access.NewStaticGenerateDataplaneTokenAccess(builder.Config().Access.Static.GenerateDPToken),
	})

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

func newResourceManager(builder *core_runtime.Builder) core_manager.CustomizableResourceManager {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
	meshManager := mesh_managers.NewMeshManager(builder.ResourceStore(), customizableManager, builder.CaManagers(), registry.Global(), builder.ResourceValidators().Mesh)
	customManagers[core_mesh.MeshType] = meshManager

	secretManager := secret_manager.NewSecretManager(builder.SecretStore(), secret_cipher.None(), nil)
	customManagers[system.SecretType] = secretManager
	return customizableManager
}

type DummyEnvoyAdminClient struct {
	PostQuitCalled *int
}

func (d *DummyEnvoyAdminClient) GenerateAPIToken(dp *core_mesh.DataplaneResource) (string, error) {
	return "token", nil
}

func (d *DummyEnvoyAdminClient) PostQuit(dataplane *core_mesh.DataplaneResource) error {
	if d.PostQuitCalled != nil {
		*d.PostQuitCalled++
	}

	return nil
}

func (d *DummyEnvoyAdminClient) ConfigDump(proxy admin.ResourceWithAddress, defaultAdminPort uint32) ([]byte, error) {
	return []byte(fmt.Sprintf(`{"envoyAdminAddress": "%s"}`, proxy.AdminAddress(defaultAdminPort))), nil
}
