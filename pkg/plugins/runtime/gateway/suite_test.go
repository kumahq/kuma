package gateway_test

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	cache_v3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/golang/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
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

type mockMetadataTracker struct{}

func (m mockMetadataTracker) Metadata(dpKey core_model.ResourceKey) *core_xds.DataplaneMetadata {
	return nil
}

func MakeDataplaneProxy(rt runtime.Runtime, key core_model.ResourceKey) *core_xds.Proxy {
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
	Expect(rt.ResourceManager().List(context.TODO(), &dataplanes, store.ListByMesh(key.Mesh))).To(Succeed())

	proxy, err := b.Build(key, &xds_context.Context{
		ControlPlane: &xds_context.ControlPlaneContext{
			AdminProxyKeyPair: nil,
			CLACache:          nil,
		},
		Mesh: xds_context.MeshContext{
			Resource:   mesh,
			Dataplanes: &dataplanes,
		},
		EnvoyAdminClient: nil,
	})
	Expect(err).To(Succeed())

	return proxy
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
	bytes, err := ioutil.ReadFile(path.Join("testdata", name))
	if err != nil {
		return err
	}

	return StoreInlineFixture(rt, bytes)
}

// StoreInlineFixture stores the given YAML object in the runtime resource manager.
func StoreInlineFixture(rt runtime.Runtime, object []byte) error {
	r, err := rest.UnmarshallToCore(object)
	if err != nil {
		return err
	}

	var opts []store.CreateOptionsFunc

	switch r.Descriptor().Scope {
	case core_model.ScopeGlobal:
		opts = append(opts, store.CreateByKey(r.GetMeta().GetName(), ""))
	case core_model.ScopeMesh:
		opts = append(opts, store.CreateByKey(r.GetMeta().GetName(), r.GetMeta().GetMesh()))
	}

	return rt.ResourceManager().Create(context.TODO(), r, opts...)
}

// BuildRuntime returns a fabricated test Runtime instance with which
// the gateway plugin is registered.
func BuildRuntime() (runtime.Runtime, error) {
	builder, err := test_runtime.BuilderFor(kuma_cp.DefaultConfig())
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

var _ = BeforeSuite(func() {
	// Ensure that the plugin is registered so that tests at least
	// have a chance of working.
	_, registered := plugins.Plugins().RuntimePlugins()["gateway"]
	Expect(registered).To(BeTrue(), "gateway plugin is registered")
})
