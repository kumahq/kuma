package manager_test

import (
	"context"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	test_resources "github.com/Kong/kuma/pkg/test/resources"
	"github.com/Kong/kuma/pkg/test/resources/apis/sample"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Manager", func() {

	var resStore store.ResourceStore
	var resManager manager.ResourceManager

	BeforeEach(func() {
		resStore = memory.NewStore()
		resManager = manager.NewResourceManager(resStore, test_resources.Global())
	})

	createSampleMesh := func(name string) error {
		meshRes := mesh.MeshResource{
			Spec: mesh_proto.Mesh{},
		}
		return resManager.Create(context.Background(), &meshRes, store.CreateByKey("default", name, name))
	}

	createSampleResource := func(mesh string) (*sample.TrafficRouteResource, error) {
		trRes := sample.TrafficRouteResource{
			Spec: v1alpha1.TrafficRoute{
				Path: "/some",
			},
		}
		err := resManager.Create(context.Background(), &trRes, store.CreateByKey("default", "tr-1", mesh))
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
		It("should delete all resources withing a mesh", func() {
			// setup
			Expect(createSampleMesh("mesh-1")).To(Succeed())
			Expect(createSampleMesh("mesh-2")).To(Succeed())
			_, err := createSampleResource("mesh-1")
			Expect(err).ToNot(HaveOccurred())
			_, err = createSampleResource("mesh-2")
			Expect(err).ToNot(HaveOccurred())

			// when
			err = resManager.DeleteAll(context.Background(), store.DeleteAllByMesh("mesh-1"))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and resource from mesh-1 is deleted
			res1 := sample.TrafficRouteResource{}
			err = resManager.Get(context.Background(), &res1, store.GetByKey("default", "tr-1", "mesh-1"))
			Expect(store.IsResourceNotFound(err)).To(BeTrue())

			// and resource from mesh-2 is retained
			res2 := sample.TrafficRouteResource{}
			err = resManager.Get(context.Background(), &res2, store.GetByKey("default", "tr-1", "mesh-2"))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
