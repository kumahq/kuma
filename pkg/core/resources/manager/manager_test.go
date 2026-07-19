package manager_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

var _ = Describe("Resource Manager", func() {
	var resStore store.ResourceStore
	var resManager manager.ResourceManager

	BeforeEach(func() {
		resStore = memory.NewStore()
		resManager = manager.NewResourceManager(resStore)
	})

	createSampleMesh := func(name string) error {
		meshRes := core_mesh.MeshResource{
			Spec: &mesh_proto.Mesh{},
		}
		return resManager.Create(context.Background(), &meshRes, store.CreateByKey(name, model.NoMesh))
	}

	createSampleResource := func(mesh string) (*core_mesh.ExternalServiceResource, error) {
		esRes := core_mesh.ExternalServiceResource{
			Spec: &mesh_proto.ExternalService{
				Networking: &mesh_proto.ExternalService_Networking{
					Address: "192.168.0.1:8080",
				},
				Tags: map[string]string{
					mesh_proto.ServiceTag: "es-1",
				},
			},
		}
		err := resManager.Create(context.Background(), &esRes, store.CreateByKey("tr-1", mesh))
		return &esRes, err
	}

	Describe("Create()", func() {
		It("should let create when mesh exists", func() {
			// given
			err := createSampleMesh("mesh-1")
			Expect(err).ToNot(HaveOccurred())

			// when
			_, err = createSampleResource("mesh-1")

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not let to create a resource when mesh not exists", func() {
			// given no mesh for resource

			// when
			_, err := createSampleResource("mesh-1")

			// then
			Expect(err.Error()).To(Equal("mesh of name mesh-1 is not found"))
		})
	})

	Describe("DeleteAll()", func() {
		It("should delete all resources within a mesh", func() {
			// setup
			Expect(createSampleMesh("mesh-1")).To(Succeed())
			Expect(createSampleMesh("mesh-2")).To(Succeed())
			_, err := createSampleResource("mesh-1")
			Expect(err).ToNot(HaveOccurred())
			_, err = createSampleResource("mesh-2")
			Expect(err).ToNot(HaveOccurred())

			dpKey := model.ResourceKey{
				Mesh: "mesh-1",
				Name: "dp-1",
			}
			dataplane := &core_mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port: 1234,
								Tags: map[string]string{
									"kuma.io/service": "backend",
								},
							},
						},
					},
				},
			}
			err = resManager.Create(context.Background(), dataplane, store.CreateBy(dpKey))
			Expect(err).ToNot(HaveOccurred())

			// when
			err = resManager.DeleteAll(context.Background(), &core_mesh.ExternalServiceResourceList{}, store.DeleteAllByMesh("mesh-1"))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and resource from mesh-1 is deleted
			res1 := core_mesh.NewExternalServiceResource()
			err = resManager.Get(context.Background(), res1, store.GetByKey("tr-1", "mesh-1"))
			Expect(store.IsNotFound(err)).To(BeTrue())

			// and only ExternalServices are deleted
			Expect(resManager.Get(context.Background(), core_mesh.NewDataplaneResource(), store.GetBy(dpKey))).To(Succeed())

			// and resource from mesh-2 is retained
			res2 := core_mesh.NewExternalServiceResource()
			err = resManager.Get(context.Background(), res2, store.GetByKey("tr-1", "mesh-2"))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
