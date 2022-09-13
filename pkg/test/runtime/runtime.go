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

func RuntimeFor(appCtx context.Context, cfg kuma_cp.Config) (core_runtime.Runtime, error) {
	r, err := core_runtime.NewRuntime(appCtx, cfg)
	if err != nil {
		return nil, err
	}

	m, _ := metrics.NewMetrics("Standalone")
	err = core_runtime.ApplyOpts(r,
		core_runtime.WithMetrics(m),
		core_runtime.WithComponentManager(component.NewManager(leader_memory.NewAlwaysLeaderElector())),
		core_runtime.WithResourceStore(resources_memory.NewStore()),
	)
	if err != nil {
		return nil, err
	}
	err = core_runtime.ApplyOpts(r,
		core_runtime.WithSecretStore(secret_store.NewSecretStore(r.ResourceStore())),
		core_runtime.WithResourceValidators(core_runtime.ResourceValidators{
			Dataplane: dataplane.NewMembershipValidator(),
			Mesh:      mesh_managers.NewMeshValidator(r.CaManagers(), r.ResourceStore()),
		}),
	)
	if err != nil {
		return nil, err
	}

	rm := newResourceManager(r)
	err = core_runtime.ApplyOpts(r, core_runtime.WithResourceManager(rm), core_runtime.WithReadOnlyResourceManager(rm))
	if err != nil {
		return nil, err
	}
	err = core_runtime.ApplyOpts(r,
		core_runtime.WithCaManager("builtin", builtin.NewBuiltinCaManager(r.ResourceManager())),
	)
	if err != nil {
		return nil, err
	}
	caProvider, err := secrets.NewCaProvider(r.CaManagers(), m)
	if err != nil {
		return nil, err
	}
	err = core_runtime.ApplyOpts(r, core_runtime.WithCAProvider(caProvider))
	if err != nil {
		return nil, err
	}

	err = core_runtime.ApplyOpts(r,
		core_runtime.WithDataSourceLoader(datasource.NewDataSourceLoader(r.ResourceManager())),
		core_runtime.WithLeaderInfo(&component.LeaderInfoComponent{}),
		core_runtime.WithLookupIP(func(s string) ([]net.IP, error) {
			return nil, errors.New("LookupIP not set, set one in your test to resolve things")
		}),
		core_runtime.WithEnvoyAdminClient(&DummyEnvoyAdminClient{}),
		core_runtime.WithEventReaderFactory(events.NewEventBus()),
		core_runtime.WithAPIManager(customization.NewAPIList()),
		core_runtime.WithDpServer(server.NewDpServer(*cfg.DpServer, m)),
		core_runtime.WithKDSContext(kds_context.DefaultContext(appCtx, r.ResourceManager(), cfg.Multizone.Zone.Name)),
		core_runtime.WithAPIServerAuthenticator(certs.ClientCertAuthenticator),
		core_runtime.WithAccess(core_runtime.Access{
			ResourceAccess:       resources_access.NewAdminResourceAccess(r.Config().Access.Static.AdminResources),
			DataplaneTokenAccess: tokens_access.NewStaticGenerateDataplaneTokenAccess(r.Config().Access.Static.GenerateDPToken),
		}),
		core_runtime.WithConfigManager(config_manager.NewConfigManager(r.ResourceStore())),
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func newResourceManager(r core_runtime.Runtime) core_manager.CustomizableResourceManager {
	defaultManager := core_manager.NewResourceManager(r.ResourceStore())
	customManagers := map[core_model.ResourceType]core_manager.ResourceManager{}
	customizableManager := core_manager.NewCustomizableResourceManager(defaultManager, customManagers)
	meshManager := mesh_managers.NewMeshManager(
		r.ResourceStore(),
		customizableManager,
		r.CaManagers(),
		registry.Global(),
		r.ResourceValidators().Mesh,
		r.Config().Store.UnsafeDelete,
	)
	customManagers[core_mesh.MeshType] = meshManager

	secretManager := secret_manager.NewSecretManager(r.SecretStore(), secret_cipher.None(), nil, r.Config().Store.UnsafeDelete)
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
