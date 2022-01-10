package gateway_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/Nordix/simple-ipam/pkg/ipam"
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cache_v3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/secrets"
	"github.com/kumahq/kuma/pkg/xds/sync"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

func TestGateway(t *testing.T) {
	test.RunSpecs(t, "Gateway Suite")
}

type ProtoMessage struct {
	Message proto.Message
}

func (p ProtoMessage) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(p.Message)
}

type ProtoResource struct {
	Resources map[string]ProtoMessage
}

type ProtoSnapshot struct {
	Clusters  ProtoResource
	Endpoints ProtoResource
	Listeners ProtoResource
	Routes    ProtoResource
	Runtimes  ProtoResource
	Secrets   ProtoResource
}

// MakeProtoResource wraps Go Control Plane resources in a map that
// implements the json.Marshaler so that the resulting JSON fully
// expands embedded Any protobufs (which are otherwise serialized
// as byte arrays).
func MakeProtoResource(resources cache_v3.Resources) ProtoResource {
	result := ProtoResource{
		Resources: map[string]ProtoMessage{},
	}

	for name, values := range resources.Items {
		result.Resources[name] = ProtoMessage{
			Message: values.Resource,
		}
	}

	return result
}

func MakeProtoSnapshot(snap cache_v3.Snapshot) ProtoSnapshot {
	return ProtoSnapshot{
		Clusters:  MakeProtoResource(snap.Resources[envoy_types.Cluster]),
		Endpoints: MakeProtoResource(snap.Resources[envoy_types.Endpoint]),
		Listeners: MakeProtoResource(snap.Resources[envoy_types.Listener]),
		Routes:    MakeProtoResource(snap.Resources[envoy_types.Route]),
		Runtimes:  MakeProtoResource(snap.Resources[envoy_types.Runtime]),
		Secrets:   MakeProtoResource(snap.Resources[envoy_types.Secret]),
	}
}

type mockMetadataTracker struct{}

func (m mockMetadataTracker) Metadata(dpKey core_model.ResourceKey) *core_xds.DataplaneMetadata {
	return nil
}

func MakeGeneratorContext(rt runtime.Runtime, key core_model.ResourceKey) (*xds_context.Context, *core_xds.Proxy) {
	b := sync.DataplaneProxyBuilder{
		CachingResManager:    rt.ReadOnlyResourceManager(),
		NonCachingResManager: rt.ResourceManager(),
		LookupIP:             rt.LookupIP(),
		DataSourceLoader:     rt.DataSourceLoader(),
		MetadataTracker:      mockMetadataTracker{},
		PermissionMatcher: permissions.TrafficPermissionsMatcher{
			ResourceManager: rt.ReadOnlyResourceManager(),
		},
		LogsMatcher: logs.TrafficLogsMatcher{
			ResourceManager: rt.ReadOnlyResourceManager(),
		},
		FaultInjectionMatcher: faultinjections.FaultInjectionMatcher{
			ResourceManager: rt.ReadOnlyResourceManager(),
		},
		RateLimitMatcher: ratelimits.RateLimitMatcher{
			ResourceManager: rt.ReadOnlyResourceManager(),
		},
		Zone:       rt.Config().Multizone.Zone.Name,
		APIVersion: envoy.APIV3,
	}

	mesh := core_mesh.NewMeshResource()
	Expect(rt.ReadOnlyResourceManager().Get(context.TODO(), mesh, store.GetByKey(key.Mesh, core_model.NoMesh))).
		To(Succeed())

	dataplanes := core_mesh.DataplaneResourceList{}
	Expect(rt.ResourceManager().List(context.TODO(), &dataplanes, store.ListByMesh(key.Mesh))).
		To(Succeed())

	cache, err := cla.NewCache(rt.Config().Store.Cache.ExpirationTime, rt.Metrics())
	Expect(err).To(Succeed())

	secrets, err := secrets.NewSecrets(
		secrets.NewCaProvider(rt.CaManagers()),
		secrets.NewIdentityProvider(rt.CaManagers()),
		rt.Metrics(),
	)
	Expect(err).To(Succeed())

	control, err := xds_context.BuildControlPlaneContext(rt.Config(), cache, secrets)
	Expect(err).To(Succeed())

	ctx := xds_context.Context{
		ControlPlane: control,
		Mesh: xds_context.MeshContext{
			Resource:    mesh,
			Dataplanes:  &dataplanes,
			EndpointMap: xds_topology.BuildEdsEndpointMap(mesh, rt.Config().Multizone.Zone.Name, dataplanes.Items, []*core_mesh.ZoneIngressResource{}),
		},
	}

	proxy, err := b.Build(key, &ctx)
	Expect(err).To(Succeed())

	return &ctx, proxy
}

// FetchNamedFixture retrieves the named resource from the runtime
// resource manager.
func FetchNamedFixture(
	rt runtime.Runtime,
	resourceType core_model.ResourceType,
	key core_model.ResourceKey,
) (core_model.Resource, error) {
	r, err := registry.Global().NewObject(resourceType)
	if err != nil {
		return nil, err
	}

	if err := rt.ReadOnlyResourceManager().Get(context.TODO(), r, store.GetBy(key)); err != nil {
		return nil, err
	}

	return r, nil
}

// StoreNamedFixture reads the given YAML file name from the testdata
// directory, then stores it in the runtime resource manager.
func StoreNamedFixture(rt runtime.Runtime, name string) error {
	bytes, err := os.ReadFile(path.Join("testdata", name))
	if err != nil {
		return err
	}

	return StoreInlineFixture(rt, bytes)
}

// StoreInlineFixture stores or updates the given YAML object in the
// runtime resource manager.
func StoreInlineFixture(rt runtime.Runtime, object []byte) error {
	r, err := rest.UnmarshallToCore(object)
	if err != nil {
		return err
	}

	return StoreFixture(rt.ResourceManager(), r)
}

// StoreFixture stores or updates the given resource in the runtime
// resource manager.
func StoreFixture(mgr manager.ResourceManager, r core_model.Resource) error {
	key := core_model.MetaToResourceKey(r.GetMeta())
	current, err := registry.Global().NewObject(r.Descriptor().Name)
	if err != nil {
		return err
	}

	return manager.Upsert(mgr, key, current,
		func(resource core_model.Resource) error {
			return resource.SetSpec(r.GetSpec())
		},
	)
}

// BuildRuntime returns a fabricated test Runtime instance with which
// the gateway plugin is registered.
func BuildRuntime() (runtime.Runtime, error) {
	builder, err := test_runtime.BuilderFor(context.Background(), kuma_cp.DefaultConfig())
	if err != nil {
		return nil, err
	}

	rt, err := builder.Build()
	if err != nil {
		return nil, err
	}

	if err := plugins.Plugins().RuntimePlugins()["gateway"].Customize(rt); err != nil {
		return nil, err
	}

	return rt, nil
}

// DataplaneGenerator generates Dataplane resources and stores them in the resource manager.
type DataplaneGenerator struct {
	Mesh      string
	Addresser *ipam.IPAM
	Manager   manager.ResourceManager
	NextPort  uint32
}

// Generate creates a single Dataplane resource.
func (d *DataplaneGenerator) Generate(
	serviceName string,
	tags ...string,
) {
	d.generate(serviceName+"-0", serviceName, tags...)
}

// GenerateN creates a multiple Dataplane resources.
func (d *DataplaneGenerator) GenerateN(
	count int,
	serviceName string,
	tags ...string,
) {
	for i := 0; i < count; i++ {
		d.generate(fmt.Sprintf("%s-%d", serviceName, i), serviceName, tags...)
	}
}

func (d *DataplaneGenerator) init() {
	if d.Mesh == "" {
		d.Mesh = "default"
	}

	if d.Addresser == nil {
		i, err := ipam.New("192.168.1.0/24")
		Expect(err).To(Succeed(), "IPAM initializer failed")

		d.Addresser = i
		d.Addresser.ReserveFirstAndLast()
	}

	if d.NextPort == 0 {
		d.NextPort = 20000
	}
}

func (d *DataplaneGenerator) generate(
	resourceName string,
	serviceName string,
	tags ...string,
) {
	d.init()

	// Tags have to come in pairs.
	Expect(len(tags) % 2).To(BeZero())

	addr, err := d.Addresser.Allocate()
	Expect(err).To(Succeed())

	d.NextPort++

	dp := core_mesh.NewDataplaneResource()
	dp.SetMeta(&rest.ResourceMeta{
		Type:             string(dp.Descriptor().Name),
		Mesh:             "default",
		Name:             resourceName,
		CreationTime:     core.Now(),
		ModificationTime: core.Now(),
	})
	dp.Spec.Networking = &mesh_proto.Dataplane_Networking{
		Address: addr.String(),
		Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
			{
				Port:        d.NextPort,
				ServicePort: d.NextPort,
				Tags: map[string]string{
					mesh_proto.ServiceTag:  serviceName,
					mesh_proto.ProtocolTag: "http",
				},
			}},
	}

	for i := 0; i < len(tags); i += 2 {
		dp.Spec.GetNetworking().GetInbound()[0].Tags[tags[i]] = tags[i+1]
	}

	Expect(StoreFixture(d.Manager, dp)).To(Succeed())
}

var _ = BeforeSuite(func() {
	// Ensure that the plugin is registered so that tests at least
	// have a chance of working.
	_, registered := plugins.Plugins().RuntimePlugins()["gateway"]
	Expect(registered).To(BeTrue(), "gateway plugin is registered")
})
