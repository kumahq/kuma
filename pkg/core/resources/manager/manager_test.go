package manager_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	"github.com/Kong/kuma/pkg/test/resources/apis/sample"
)

var _ = Describe("Resource Manager", func() {

	var resStore store.ResourceStore
	var resManager manager.ResourceManager

	BeforeEach(func() {
		resStore = memory.NewStore()
		resManager = manager.NewResourceManager(resStore)
	})

	createSampleMesh := func(name string) error {
		meshRes := mesh.MeshResource{
			Spec: mesh_proto.Mesh{},
		}
		return resManager.Create(context.Background(), &meshRes, store.CreateByKey(name, name))
	}

	createSampleResource := func(mesh string) (*sample.TrafficRouteResource, error) {
		trRes := sample.TrafficRouteResource{
			Spec: v1alpha1.TrafficRoute{
				Path: "/some",
			},
		}
		err := resManager.Create(context.Background(), &trRes, store.CreateByKey("tr-1", mesh))
		return &trRes, err
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

			tlKey := model.ResourceKey{
				Mesh: "mesh-1",
				Name: "tl-1",
			}
			trafficLog := &mesh.TrafficLogResource{
				Spec: mesh_proto.TrafficLog{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"service": "*",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"service": "*",
							},
						},
					},
				},
			}
			err = resManager.Create(context.Background(), trafficLog, store.CreateBy(tlKey))
			Expect(err).ToNot(HaveOccurred())

			// when
			err = resManager.DeleteAll(context.Background(), &sample.TrafficRouteResourceList{}, store.DeleteAllByMesh("mesh-1"))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and resource from mesh-1 is deleted
			res1 := sample.TrafficRouteResource{}
			err = resManager.Get(context.Background(), &res1, store.GetByKey("tr-1", "mesh-1"))
			Expect(store.IsResourceNotFound(err)).To(BeTrue())

			// and only TrafficRoutes are deleted
			Expect(resManager.Get(context.Background(), &mesh.TrafficLogResource{}, store.GetBy(tlKey))).To(Succeed())

			// and resource from mesh-2 is retained
			res2 := sample.TrafficRouteResource{}
			err = resManager.Get(context.Background(), &res2, store.GetByKey("tr-1", "mesh-2"))
			Expect(err).ToNot(HaveOccurred())

		})
	})
})
