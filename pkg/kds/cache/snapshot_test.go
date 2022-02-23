package cache_test

import (
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/kds/cache"
)

var _ = Describe("Snapshot", func() {
	mustMarshalAny := func(pb proto.Message) *anypb.Any {
		a, err := anypb.New(pb)
		if err != nil {
			panic(err)
		}
		return a
	}

	Describe("Consistent()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *cache.Snapshot
			// then
			Expect(snapshot.Consistent()).To(MatchError("nil snapshot"))
		})

		It("non-`nil` snapshot should be always consistet", func() {
			// when
			snapshot := cache.NewSnapshotBuilder().Build("v1")
			// then
			Expect(snapshot.Consistent()).To(BeNil())

			// when
			snapshot = cache.NewSnapshotBuilder().
				With(string(core_mesh.MeshType), []envoy_types.Resource{
					&mesh_proto.KumaResource{
						Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
						Spec: mustMarshalAny(&mesh_proto.Mesh{}),
					},
				}).
				Build("v2")
			// then
			Expect(snapshot.Consistent()).To(BeNil())
		})
	})

	Describe("GetResources()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *cache.Snapshot
			// then
			Expect(snapshot.GetResources(string(core_mesh.MeshType))).To(BeNil())
		})

		It("should return Meshes", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			// when
			snapshot := cache.NewSnapshotBuilder().
				With(string(core_mesh.MeshType), []envoy_types.Resource{resources}).
				Build("v1")
			// then
			expected := map[string]envoy_types.Resource{
				"mesh1.mesh1": resources,
			}
			Expect(snapshot.GetResources(string(core_mesh.MeshType))).To(Equal(expected))
		})

		It("should return `nil` for unsupported resource types", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			// when
			snapshot := cache.NewSnapshotBuilder().
				With(string(core_mesh.MeshType), []envoy_types.Resource{resources}).
				Build("v1")
			// then
			Expect(snapshot.GetResources("UnsupportedType")).To(BeNil())
		})
	})

	Describe("GetVersion()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *cache.Snapshot
			// then
			Expect(snapshot.GetVersion(string(core_mesh.MeshType))).To(Equal(""))
		})

		It("should return proper version for a supported resource type", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			// when
			snapshot := cache.NewSnapshotBuilder().
				With(string(core_mesh.MeshType), []envoy_types.Resource{resources}).
				Build("v1")
			// then
			Expect(snapshot.GetVersion(string(core_mesh.MeshType))).To(Equal("v1"))
		})

		It("should return an empty string for unsupported resource type", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			// when
			snapshot := cache.NewSnapshotBuilder().
				With(string(core_mesh.MeshType), []envoy_types.Resource{resources}).
				Build("v1")
			// then
			Expect(snapshot.GetVersion("unsupported type")).To(Equal(""))
		})
	})

	Describe("WithVersion()", func() {
		It("should handle `nil`", func() {
			// given
			var snapshot *cache.Snapshot
			// when
			actual := snapshot.WithVersion(string(core_mesh.MeshType), "v1")
			// then
			Expect(actual).To(BeNil())
		})

		It("should return a new snapshot if version has changed", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			// when
			snapshot := cache.NewSnapshotBuilder().
				With(string(core_mesh.MeshType), []envoy_types.Resource{resources}).
				Build("v1")
			// when
			actual := snapshot.WithVersion(string(core_mesh.MeshType), "v2")
			// then
			Expect(actual.GetVersion(string(core_mesh.MeshType))).To(Equal("v2"))
			// and
			Expect(actual.GetVersion(string(core_mesh.CircuitBreakerType))).To(Equal("v1"))
		})

		It("should return the same snapshot if version has not changed", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			// when
			snapshot := cache.NewSnapshotBuilder().
				With(string(core_mesh.MeshType), []envoy_types.Resource{resources}).
				Build("v1")
			// when
			actual := snapshot.WithVersion(string(core_mesh.MeshType), "v1")
			// then
			Expect(actual.GetVersion(string(core_mesh.MeshType))).To(Equal("v1"))
			// and
			Expect(actual).To(BeIdenticalTo(snapshot))
		})

		It("should return the same snapshot if resource type is not supported", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			// when
			snapshot := cache.NewSnapshotBuilder().
				With(string(core_mesh.MeshType), []envoy_types.Resource{resources}).
				Build("v1")
			// when
			actual := snapshot.WithVersion("unsupported type", "v2")
			// then
			Expect(actual.GetVersion(string(core_mesh.MeshType))).To(Equal("v1"))
			// and
			Expect(actual).To(BeIdenticalTo(snapshot))
		})
	})
})
