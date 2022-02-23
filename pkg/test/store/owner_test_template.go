package store

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	sample_model "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
)

func ExecuteOwnerTests(
	createStore func() store.ResourceStore,
) {
	const mesh = "default-mesh"
	var s store.ClosableResourceStore

	BeforeEach(func() {
		s = store.NewStrictResourceStore(createStore())
	})

	AfterEach(func() {
		err := s.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should delete resource when its owner is deleted", func() {
		// setup
		meshRes := core_mesh.NewMeshResource()
		err := s.Create(context.Background(), meshRes, store.CreateByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		name := "resource-1"
		trRes := sample_model.TrafficRouteResource{
			Spec: &sample_proto.TrafficRoute{
				Path: "demo",
			},
		}
		err = s.Create(context.Background(), &trRes,
			store.CreateByKey(name, mesh),
			store.CreatedAt(time.Now()),
			store.CreateWithOwner(meshRes))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = s.Delete(context.Background(), meshRes, store.DeleteByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// then
		actual := sample_model.NewTrafficRouteResource()
		err = s.Get(context.Background(), actual, store.GetByKey(name, mesh))
		Expect(store.IsResourceNotFound(err)).To(BeTrue())
	})

	It("should delete several resources when their owner is deleted", func() {
		// setup
		meshRes := core_mesh.NewMeshResource()
		err := s.Create(context.Background(), meshRes, store.CreateByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 10; i++ {
			tr := sample_model.TrafficRouteResource{
				Spec: &sample_proto.TrafficRoute{
					Path: "demo",
				},
			}
			err = s.Create(context.Background(), &tr,
				store.CreateByKey(fmt.Sprintf("resource-%d", i), mesh),
				store.CreatedAt(time.Now()),
				store.CreateWithOwner(meshRes))
			Expect(err).ToNot(HaveOccurred())
		}
		actual := sample_model.TrafficRouteResourceList{}
		err = s.List(context.Background(), &actual, store.ListByMesh(mesh))
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(HaveLen(10))

		// when
		err = s.Delete(context.Background(), meshRes, store.DeleteByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// then
		actual = sample_model.TrafficRouteResourceList{}
		err = s.List(context.Background(), &actual, store.ListByMesh(mesh))
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(BeEmpty())
	})

	It("should delete owners chain", func() {
		// setup
		meshRes := core_mesh.NewMeshResource()
		err := s.Create(context.Background(), meshRes, store.CreateByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		var prev model.Resource = meshRes
		for i := 0; i < 10; i++ {
			curr := &sample_model.TrafficRouteResource{
				Spec: &sample_proto.TrafficRoute{
					Path: "demo",
				},
			}
			err := s.Create(context.Background(), curr,
				store.CreateByKey(fmt.Sprintf("resource-%d", i), mesh),
				store.CreatedAt(time.Now()),
				store.CreateWithOwner(prev))
			Expect(err).ToNot(HaveOccurred())
			prev = curr
		}

		actual := sample_model.TrafficRouteResourceList{}
		err = s.List(context.Background(), &actual, store.ListByMesh(mesh))
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(HaveLen(10))

		// when
		err = s.Delete(context.Background(), meshRes, store.DeleteByKey(mesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// then
		actual = sample_model.TrafficRouteResourceList{}
		err = s.List(context.Background(), &actual, store.ListByMesh(mesh))
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Items).To(BeEmpty())
	})
}
