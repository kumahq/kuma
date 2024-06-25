package meshservice_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

var _ = Describe("MatchDataplanesWithMeshServices", func() {
	type testCase struct {
		dpps             []*core_mesh.DataplaneResource
		meshServices     []*meshservice_api.MeshServiceResource
		expectedSummary  map[model.ResourceKey][]model.ResourceKey
		matchOnlyHealthy bool
	}

	DescribeTable("matching dpps for mesh service",
		func(given testCase) {
			result := meshservice.MatchDataplanesWithMeshServices(given.dpps, given.meshServices, given.matchOnlyHealthy)
			summary := map[model.ResourceKey][]model.ResourceKey{}
			for ms, dpps := range result {
				msKey := model.MetaToResourceKey(ms.GetMeta())
				summary[msKey] = []model.ResourceKey{}
				for _, dpp := range dpps {
					summary[msKey] = append(summary[msKey], model.MetaToResourceKey(dpp.GetMeta()))
				}
			}
			Expect(summary).To(Equal(given.expectedSummary))
		},
		Entry("should match by tags", testCase{
			dpps: []*core_mesh.DataplaneResource{
				builders.Dataplane().
					WithName("redis-01").
					WithInboundOfTags("kuma.io/service", "redis", "app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
				builders.Dataplane().
					WithName("redis-02").
					WithInboundOfTags("kuma.io/service", "redis", "app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
				builders.Dataplane().
					WithName("demo-app-01").
					WithInboundOfTags("kuma.io/service", "demo-app", "app", "demo-app", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			meshServices: []*meshservice_api.MeshServiceResource{
				builders.MeshService().
					WithName("redis").
					WithDataplaneTagsSelectorKV("app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
				builders.MeshService().
					WithName("demo-app").
					WithDataplaneTagsSelectorKV("app", "demo-app", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			expectedSummary: map[model.ResourceKey][]model.ResourceKey{
				{Mesh: "default", Name: "redis"}: {
					{Mesh: "default", Name: "redis-01"},
					{Mesh: "default", Name: "redis-02"},
				},
				{Mesh: "default", Name: "demo-app"}: {
					{Mesh: "default", Name: "demo-app-01"},
				},
			},
		}),
		Entry("should not duplicate multiple inbounds in the result", testCase{
			dpps: []*core_mesh.DataplaneResource{
				builders.Dataplane().
					WithName("demo-app-01").
					WithoutInbounds().
					AddInboundOfTags("kuma.io/service", "demo-app", "app", "demo-app", "k8s.kuma.io/namespace", "kuma-demo").
					AddInboundOfTags("kuma.io/service", "demo-app-admin", "app", "demo-app", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			meshServices: []*meshservice_api.MeshServiceResource{
				builders.MeshService().
					WithName("demo-app").
					WithDataplaneTagsSelectorKV("app", "demo-app", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			expectedSummary: map[model.ResourceKey][]model.ResourceKey{
				{Mesh: "default", Name: "demo-app"}: {
					{Mesh: "default", Name: "demo-app-01"},
				},
			},
		}),
		Entry("should not mix meshes", testCase{
			dpps: []*core_mesh.DataplaneResource{
				builders.Dataplane().
					WithName("redis-01").
					WithInboundOfTags("kuma.io/service", "redis", "app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
				builders.Dataplane().
					WithName("redis-02").
					WithMesh("mesh-other").
					WithInboundOfTags("kuma.io/service", "redis", "app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			meshServices: []*meshservice_api.MeshServiceResource{
				builders.MeshService().
					WithName("redis").
					WithDataplaneTagsSelectorKV("app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
				builders.MeshService().
					WithName("redis").
					WithMesh("mesh-other").
					WithDataplaneTagsSelectorKV("app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			expectedSummary: map[model.ResourceKey][]model.ResourceKey{
				{Mesh: "default", Name: "redis"}: {
					{Mesh: "default", Name: "redis-01"},
				},
				{Mesh: "mesh-other", Name: "redis"}: {
					{Mesh: "mesh-other", Name: "redis-02"},
				},
			},
		}),
		Entry("should not match mesh service that selects nothing", testCase{
			dpps: []*core_mesh.DataplaneResource{
				builders.Dataplane().
					WithName("redis-01").
					WithInboundOfTags("kuma.io/service", "redis", "app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			meshServices: []*meshservice_api.MeshServiceResource{
				builders.MeshService().
					WithName("demo-app").
					WithDataplaneTagsSelectorKV("app", "demo-app", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			expectedSummary: map[model.ResourceKey][]model.ResourceKey{
				{Mesh: "default", Name: "demo-app"}: {},
			},
		}),
		Entry("should not match unhealthy inbound matched by tags", testCase{
			matchOnlyHealthy: true,
			dpps: []*core_mesh.DataplaneResource{
				builders.Dataplane().
					WithName("redis-01").
					WithInboundOfTags("kuma.io/service", "redis", "app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					With(func(resource *core_mesh.DataplaneResource) {
						resource.Spec.Networking.Inbound[0].State = mesh_proto.Dataplane_Networking_Inbound_NotReady
					}).
					Build(),
			},
			meshServices: []*meshservice_api.MeshServiceResource{
				builders.MeshService().
					WithName("redis").
					WithDataplaneTagsSelectorKV("app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			expectedSummary: map[model.ResourceKey][]model.ResourceKey{
				{Mesh: "default", Name: "redis"}: {},
			},
		}),
		Entry("should match by ref name", testCase{
			dpps: []*core_mesh.DataplaneResource{
				builders.Dataplane().
					WithName("redis-01").
					WithInboundOfTags("kuma.io/service", "redis", "app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
				builders.Dataplane().
					WithName("demo-app-01").
					WithInboundOfTags("kuma.io/service", "demo-app", "app", "demo-app", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			meshServices: []*meshservice_api.MeshServiceResource{
				builders.MeshService().
					WithName("redis").
					WithDataplaneRefNameSelector("redis-01").
					Build(),
				builders.MeshService().
					WithName("demo-app").
					WithDataplaneRefNameSelector("demo-app-01").
					Build(),
			},
			expectedSummary: map[model.ResourceKey][]model.ResourceKey{
				{Mesh: "default", Name: "redis"}: {
					{Mesh: "default", Name: "redis-01"},
				},
				{Mesh: "default", Name: "demo-app"}: {
					{Mesh: "default", Name: "demo-app-01"},
				},
			},
		}),
		Entry("should not match names that did not match", testCase{
			dpps: []*core_mesh.DataplaneResource{
				builders.Dataplane().
					WithName("redis-01").
					WithInboundOfTags("kuma.io/service", "redis", "app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					Build(),
			},
			meshServices: []*meshservice_api.MeshServiceResource{
				builders.MeshService().
					WithName("redis").
					WithDataplaneRefNameSelector("redis-not-found").
					Build(),
			},
			expectedSummary: map[model.ResourceKey][]model.ResourceKey{
				{Mesh: "default", Name: "redis"}: {},
			},
		}),
		Entry("should not match unhealthy inbound matched by name", testCase{
			matchOnlyHealthy: true,
			dpps: []*core_mesh.DataplaneResource{
				builders.Dataplane().
					WithName("redis-01").
					WithInboundOfTags("kuma.io/service", "redis", "app", "redis", "k8s.kuma.io/namespace", "kuma-demo").
					With(func(resource *core_mesh.DataplaneResource) {
						resource.Spec.Networking.Inbound[0].State = mesh_proto.Dataplane_Networking_Inbound_NotReady
					}).
					Build(),
			},
			meshServices: []*meshservice_api.MeshServiceResource{
				builders.MeshService().
					WithName("redis").
					WithDataplaneRefNameSelector("redis-01").
					Build(),
			},
			expectedSummary: map[model.ResourceKey][]model.ResourceKey{
				{Mesh: "default", Name: "redis"}: {},
			},
		}),
	)
})
