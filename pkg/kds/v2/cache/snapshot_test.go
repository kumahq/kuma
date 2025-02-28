package cache_test

import (
	"fmt"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/kds/v2/cache"
)

var _ = Describe("Snapshot", func() {
	mustMarshalAny := func(pb proto.Message) *anypb.Any {
		a, err := anypb.New(pb)
		if err != nil {
			panic(err)
		}
		return a
	}

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
			snapshot := cache.NewSnapshotBuilder([]model.ResourceType{core_mesh.MeshType}).
				With(core_mesh.MeshType, []envoy_types.Resource{resources}).
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
			snapshot := cache.NewSnapshotBuilder([]model.ResourceType{core_mesh.MeshType}).
				With(core_mesh.MeshType, []envoy_types.Resource{resources}).
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
			snapshot := cache.NewSnapshotBuilder([]model.ResourceType{core_mesh.MeshType}).
				With(core_mesh.MeshType, []envoy_types.Resource{resources}).
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
			snapshot := cache.NewSnapshotBuilder([]model.ResourceType{core_mesh.MeshType}).
				With(core_mesh.MeshType, []envoy_types.Resource{resources}).
				Build("v1")
			// then
			Expect(snapshot.GetVersion("unsupported type")).To(Equal(""))
		})
	})

	Describe("ConstructVersionMap()", func() {
		It("should handle `nil`", func() {
			// when
			var snapshot *cache.Snapshot
			// then
			Expect(snapshot.ConstructVersionMap()).To(Equal(fmt.Errorf("missing snapshot")))
		})

		It("should construct version map for resource", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			snapshot := cache.NewSnapshotBuilder([]model.ResourceType{core_mesh.MeshType}).
				With(core_mesh.MeshType, []envoy_types.Resource{resources}).
				Build("v1")

			// when
			Expect(snapshot.ConstructVersionMap()).ToNot(HaveOccurred())

			// then
			Expect(snapshot.GetVersionMap(string(core_mesh.MeshType))["mesh1.mesh1"]).ToNot(BeEmpty())
		})

		It("should change version when resource has changed", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			snapshot := cache.NewSnapshotBuilder([]model.ResourceType{core_mesh.MeshType}).
				With(core_mesh.MeshType, []envoy_types.Resource{resources}).
				Build("v1")

			// when
			Expect(snapshot.ConstructVersionMap()).ToNot(HaveOccurred())

			// then
			Expect(snapshot.GetVersionMap(string(core_mesh.MeshType))["mesh1.mesh1"]).ToNot(BeEmpty())

			// when
			previousVersion := snapshot.GetVersionMap(string(core_mesh.MeshType))["mesh1.mesh1"]

			// given
			resources = &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "ca",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "ca",
								Type: "builtin",
							},
						},
					},
				}),
			}
			snapshot = cache.NewSnapshotBuilder([]model.ResourceType{core_mesh.MeshType}).
				With(core_mesh.MeshType, []envoy_types.Resource{resources}).
				Build("v1")

			// when
			Expect(snapshot.ConstructVersionMap()).ToNot(HaveOccurred())

			// then
			Expect(snapshot.GetVersionMap(string(core_mesh.MeshType))["mesh1.mesh1"]).ToNot(Equal(previousVersion))
		})

		It("should not change version when resource has not changed", func() {
			// given
			resources := &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			snapshot := cache.NewSnapshotBuilder([]model.ResourceType{core_mesh.MeshType}).
				With(core_mesh.MeshType, []envoy_types.Resource{resources}).
				Build("v1")

			// when
			Expect(snapshot.ConstructVersionMap()).ToNot(HaveOccurred())

			// then
			Expect(snapshot.GetVersionMap(string(core_mesh.MeshType))["mesh1.mesh1"]).ToNot(BeEmpty())

			// when
			previousVersion := snapshot.GetVersionMap(string(core_mesh.MeshType))["mesh1.mesh1"]

			// given
			resources = &mesh_proto.KumaResource{
				Meta: &mesh_proto.KumaResource_Meta{Name: "mesh1", Mesh: "mesh1"},
				Spec: mustMarshalAny(&mesh_proto.Mesh{}),
			}
			snapshot = cache.NewSnapshotBuilder([]model.ResourceType{core_mesh.MeshType}).
				With(core_mesh.MeshType, []envoy_types.Resource{resources}).
				Build("v1")

			// when
			Expect(snapshot.ConstructVersionMap()).ToNot(HaveOccurred())

			// then
			Expect(snapshot.GetVersionMap(string(core_mesh.MeshType))["mesh1.mesh1"]).To(Equal(previousVersion))
		})
	})
})
