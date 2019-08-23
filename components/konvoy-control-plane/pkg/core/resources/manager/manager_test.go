package manager_test

import (
	"context"
	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/manager"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Manager", func() {

	var resStore store.ResourceStore
	var resManager manager.ResourceManager

	BeforeEach(func() {
		resStore = memory.NewStore()
		resManager = manager.NewResourceManager(resStore)
	})

	createSampleMesh := func() error {
		meshRes := mesh.MeshResource{
			Spec: mesh_proto.Mesh{},
		}
		return resManager.Create(context.Background(), &meshRes, store.CreateByKey("default", "mesh-1", "mesh-1"))
	}

	createSampleResource := func() (*sample.TrafficRouteResource, error) {
		trRes := sample.TrafficRouteResource{
			Spec: v1alpha1.TrafficRoute{
				Path: "/some",
			},
		}
		err := resManager.Create(context.Background(), &trRes, store.CreateByKey("default", "tr-1", "mesh-1"))
		return &trRes, err
	}

	Describe("Create()", func() {
		It("should let create when mesh exists", func() {
			// given
			err := createSampleMesh()
			Expect(err).ToNot(HaveOccurred())

			// when
			_, err = createSampleResource()

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not let to create a resource when mesh not exists", func() {
			// given no mesh for resource

			// when
			_, err := createSampleResource()

			// then
			Expect(err.Error()).To(Equal("mesh of name mesh-1 is not found"))
		})
	})
})
