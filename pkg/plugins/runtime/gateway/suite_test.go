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
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/test"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/secrets"
	"github.com/kumahq/kuma/pkg/xds/server"
	"github.com/kumahq/kuma/pkg/xds/sync"
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
func MakeProtoResource(resources map[string]envoy_types.ResourceWithTTL) ProtoResource {
	result := ProtoResource{
		Resources: map[string]ProtoMessage{},
	}

	for name, values := range resources {
		result.Resources[name] = ProtoMessage{
			Message: values.Resource,
		}
	}

	return result
}

func MakeProtoSnapshot(snap cache_v3.ResourceSnapshot) ProtoSnapshot {
	return ProtoSnapshot{
		Clusters:  MakeProtoResource(snap.GetResourcesAndTTL(resource.ClusterType)),
		Endpoints: MakeProtoResource(snap.GetResourcesAndTTL(resource.EndpointType)),
		Listeners: MakeProtoResource(snap.GetResourcesAndTTL(resource.ListenerType)),
		Routes:    MakeProtoResource(snap.GetResourcesAndTTL(resource.RouteType)),
		Runtimes:  MakeProtoResource(snap.GetResourcesAndTTL(resource.RuntimeType)),
		Secrets:   MakeProtoResource(snap.GetResourcesAndTTL(resource.SecretType)),
	}
}

func MakeGeneratorContext(rt runtime.Runtime, key core_model.ResourceKey) (*xds_context.Context, *core_xds.Proxy) {
	b := sync.DataplaneProxyBuilder{
		Zone:       rt.Config().Multizone.Zone.Name,
		APIVersion: envoy.APIV3,
	}

	cache, err := cla.NewCache(rt.Config().Store.Cache.ExpirationTime.Duration, rt.Metrics())
	Expect(err).To(Succeed())

	idProvider, err := secrets.NewIdentityProvider(rt.CaManagers(), rt.Metrics())
	Expect(err).To(Succeed())

	secrets, err := secrets.NewSecrets(
		rt.CAProvider(),
		idProvider,
		rt.Metrics(),
	)
	Expect(err).To(Succeed())

	cpCtx := &xds_context.ControlPlaneContext{
		CLACache: cache,
		Secrets:  secrets,
		Zone:     rt.Config().Multizone.Zone.Name,
	}

	meshCtxBuilder := xds_context.NewMeshContextBuilder(
		rt.ReadOnlyResourceManager(),
		server.MeshResourceTypes(),
		rt.LookupIP(),
		rt.Config().Multizone.Zone.Name,
		vips.NewPersistence(rt.ReadOnlyResourceManager(), rt.ConfigManager(), false),
		rt.Config().DNSServer.Domain,
		rt.Config().DNSServer.ServiceVipPort,
		xds_context.AnyToAnyReachableServicesGraphBuilder,
	)

	meshCtx, err := meshCtxBuilder.Build(context.TODO(), key.Mesh)
	Expect(err).To(Succeed())

	ctx := xds_context.Context{
		ControlPlane: cpCtx,
		Mesh:         meshCtx,
	}

	proxy, err := b.Build(context.TODO(), key, meshCtx)
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
	r, err := rest.YAML.UnmarshalCore(object)
	if err != nil {
		return err
	}

	return StoreFixture(rt.ResourceManager(), r)
}

// StoreFixture stores or updates the given resource in the runtime
// resource manager.
func StoreFixture(mgr manager.ResourceManager, r core_model.Resource) error {
	ctx := context.Background()

	key := core_model.MetaToResourceKey(r.GetMeta())
	current, err := registry.Global().NewObject(r.Descriptor().Name)
	if err != nil {
		return err
	}

	return manager.Upsert(ctx, mgr, key, current,
		func(resource core_model.Resource) error {
			return resource.SetSpec(r.GetSpec())
		},
	)
}

// BuildRuntime returns a fabricated test Runtime instance with which
// the gateway plugin is registered.
func BuildRuntime() (runtime.Runtime, error) {
	config := kuma_cp.DefaultConfig()
	builder, err := test_runtime.BuilderFor(context.Background(), config)
	if err != nil {
		return nil, err
	}

	var plugin plugins.BootstrapPlugin
	for _, p := range plugins.Plugins().BootstrapPlugins() {
		if p.Name() == metadata.PluginName {
			plugin = p
			break
		}
	}
	if err != nil {
		return nil, err
	}
	if err := plugin.BeforeBootstrap(builder, config); err != nil {
		return nil, err
	}
	if err := plugin.AfterBootstrap(builder, config); err != nil {
		return nil, err
	}

	rt, err := builder.Build()
	if err != nil {
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
	dp.SetMeta(&rest_v1alpha1.ResourceMeta{
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
			},
		},
	}

	for i := 0; i < len(tags); i += 2 {
		dp.Spec.GetNetworking().GetInbound()[0].Tags[tags[i]] = tags[i+1]
	}

	Expect(StoreFixture(d.Manager, dp)).To(Succeed())
}

var _ = BeforeSuite(func() {
	// Ensure that the plugin is registered so that tests at least
	// have a chance of working.
	Expect(plugins.Plugins().BootstrapPlugins()).To(
		WithTransform(func(in []plugins.BootstrapPlugin) []string {
			var out []string
			for _, p := range in {
				out = append(out, string(p.Name()))
			}
			return out
		}, ContainElement(metadata.PluginName)))
	Expect(plugins.Plugins().ProxyPlugins()).To(
		WithTransform(func(in map[plugins.PluginName]plugins.ProxyPlugin) []string {
			var out []string
			for k := range in {
				out = append(out, string(k))
			}
			return out
		}, ContainElement(metadata.PluginName)))
})
