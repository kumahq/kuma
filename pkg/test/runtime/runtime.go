package runtime

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"

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
	"github.com/kumahq/kuma/pkg/dp-server/server"
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
	StartTime  time.Time
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

func (i *TestRuntimeInfo) GetStartTime() time.Time {
	return i.StartTime
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
	builder.WithLookupIP(func(s string) ([]net.IP, error) {
		return nil, errors.New("LookupIP not set, set one in your test to resolve things")
	})
	builder.WithEnvoyAdminClient(&DummyEnvoyAdminClient{})
	builder.WithEventReaderFactory(events.NewEventBus())
	builder.WithAPIManager(customization.NewAPIList())
	builder.WithXDSHooks(&xds_hooks.Hooks{})
	builder.WithDpServer(server.NewDpServer(*cfg.DpServer, metrics))
	builder.WithKDSContext(kds_context.DefaultContext(appCtx, builder.ResourceManager(), cfg.Multizone.Zone.Name))
	caProvider, err := secrets.NewCaProvider(builder.CaManagers(), metrics)
	if err != nil {
		return nil, err
	}
	builder.WithCAProvider(caProvider)
	builder.WithAPIServerAuthenticator(certs.ClientCertAuthenticator)
	builder.WithAccess(core_runtime.Access{
		ResourceAccess:       resources_access.NewAdminResourceAccess(builder.Config().Access.Static.AdminResources),
		DataplaneTokenAccess: tokens_access.NewStaticGenerateDataplaneTokenAccess(builder.Config().Access.Static.GenerateDPToken),
	})

	initializeConfigManager(builder)

	return builder, nil
}

func initializeConfigManager(builder *core_runtime.Builder) {
	configm := config_manager.NewConfigManager(builder.ResourceStore())
	builder.WithConfigManager(configm)
}

func newResourceManager(builder *core_runtime.Builder) core_manager.CustomizableResourceManager {
	defaultManager := core_manager.NewResourceManager(builder.ResourceStore())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
	meshManager := mesh_managers.NewMeshManager(
		builder.ResourceStore(),
		customizableManager,
		builder.CaManagers(),
		registry.Global(),
		builder.ResourceValidators().Mesh,
		builder.Config().Store.UnsafeDelete,
	)
	customManagers[core_mesh.MeshType] = meshManager

	secretManager := secret_manager.NewSecretManager(builder.SecretStore(), secret_cipher.None(), nil, builder.Config().Store.UnsafeDelete)
	customManagers[system.SecretType] = secretManager
	return customizableManager
}

type DummyEnvoyAdminClient struct {
	PostQuitCalled *int
}

func (d *DummyEnvoyAdminClient) Stats(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	return []byte("server.live: 1\n"), nil
}

func (d *DummyEnvoyAdminClient) Clusters(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	return []byte("kuma:envoy:admin\n"), nil
}

func (d *DummyEnvoyAdminClient) GenerateAPIToken(dp *core_mesh.DataplaneResource) (string, error) {
	return "token", nil
}

func (d *DummyEnvoyAdminClient) PostQuit(ctx context.Context, dataplane *core_mesh.DataplaneResource) error {
	if d.PostQuitCalled != nil {
		*d.PostQuitCalled++
	}

	return nil
}

func (d *DummyEnvoyAdminClient) ConfigDump(ctx context.Context, proxy core_model.ResourceWithAddress) ([]byte, error) {
	return []byte(fmt.Sprintf(`{"envoyAdminAddress": "%s"}`, proxy.AdminAddress(9901))), nil
}
