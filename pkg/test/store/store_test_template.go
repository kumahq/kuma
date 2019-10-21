package store

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/store"
	sample_proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	sample_model "github.com/Kong/kuma/pkg/test/resources/apis/sample"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

func ExecuteStoreTests(
	createStore func() store.ResourceStore,
) {
	var namespace string
	const mesh = "default-mesh"
	var s store.ClosableResourceStore

	BeforeEach(func() {
		namespace = string(uuid.New())
		s = store.NewStrictResourceStore(createStore())
	})

	AfterEach(func() {
		err := s.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	createResource := func(name string) *sample_model.TrafficRouteResource {
		res := sample_model.TrafficRouteResource{
			Spec: sample_proto.TrafficRoute{
				Path: "demo",
			},
		}
		err := s.Create(context.Background(), &res, store.CreateByKey(namespace, name, mesh))
		Expect(err).ToNot(HaveOccurred())
		return &res
	}

	Describe("Create()", func() {
		It("should create a new resource", func() {
			// given
			name := "resource1"

			// when
			created := createResource(name)

			// when retrieve created object
			resource := sample_model.TrafficRouteResource{}
			err := s.Get(context.Background(), &resource, store.GetByKey(namespace, name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and it has same data
			Expect(resource.Meta.GetName()).To(Equal(name))
			Expect(resource.Meta.GetNamespace()).To(Equal(namespace))
			Expect(resource.Meta.GetMesh()).To(Equal(mesh))
			Expect(resource.Meta.GetVersion()).ToNot(BeEmpty())
			Expect(resource.Spec).To(Equal(created.Spec))
		})

		It("should not create a duplicate record", func() {
			// given
			name := "duplicated-record"
			resource := createResource(name)

			// when try to create another one with same name
			resource.SetMeta(nil)
			err := s.Create(context.Background(), resource, store.CreateByKey(namespace, name, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceAlreadyExists(resource.GetType(), namespace, name, mesh)))
		})
	})

	Describe("Update()", func() {
		It("should return an error if resource is not found", func() {
			// given
			name := "to-be-updated"
			resource := createResource(name)

			// when delete resource
			err := s.Delete(
				context.Background(),
				resource,
				store.DeleteByKey(resource.GetMeta().GetNamespace(), resource.Meta.GetName(), mesh),
			)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when trying to update nonexistent resource
			err = s.Update(context.Background(), resource)

			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(resource.GetType(), namespace, name, mesh)))
		})

		It("should update an existing resource", func() {
			// given a resources in storage
			name := "to-be-updated"
			resource := createResource(name)

			// when
			resource.Spec.Path = "new-path"
			err := s.Update(context.Background(), resource)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when retrieve the resource
			res := sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &res, store.GetByKey(namespace, name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(res.Spec.Path).To(Equal("new-path"))
		})

		//todo(jakubdyszkiewicz) write tests for optimistic locking
	})

	Describe("Delete()", func() {
		It("should succeed if resource is not found", func() {
			// given
			resource := sample_model.TrafficRouteResource{}

			// when
			err := s.Delete(context.TODO(), &resource, store.DeleteByKey(namespace, "non-existent-name", mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not delete resource from another mesh", func() {
			// given
			name := "tr-1"
			resource := createResource(name)

			// when
			resource.SetMeta(nil) // otherwise the validation from strict client fires that mesh is different
			err := s.Delete(context.TODO(), resource, store.DeleteByKey(namespace, name, "different-mesh"))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and when getting the given resource
			getResource := sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &getResource, store.GetByKey(namespace, name, mesh))

			// then resource still exists
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// given a resources in storage
			name := "to-be-deleted"
			createResource(name)

			// when
			resource := sample_model.TrafficRouteResource{}
			err := s.Delete(context.TODO(), &resource, store.DeleteByKey(namespace, name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource
			resource = sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &resource, store.GetByKey(namespace, name, mesh))

			// then resource cannot be found
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.GetType(), namespace, name, mesh)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			name := "non-existing-resource"
			resource := sample_model.TrafficRouteResource{}

			// when
			err := s.Get(context.Background(), &resource, store.GetByKey(namespace, name, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(resource.GetType(), namespace, name, mesh)))
		})

		It("should return an error if resource is not found in given mesh", func() {
			// given a resources in mesh "mesh"
			name := "existing-resource"
			mesh := "different-mesh"
			createResource(name)

			// when
			resource := sample_model.TrafficRouteResource{}
			err := s.Get(context.Background(), &resource, store.GetByKey(namespace, name, mesh))

			// then
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.GetType(), namespace, name, mesh)))
		})

		It("should return an existing resource", func() {
			// given a resources in storage
			name := "existing-resource"
			createdResource := createResource(name)

			// when
			res := sample_model.TrafficRouteResource{}
			err := s.Get(context.Background(), &res, store.GetByKey(namespace, name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(res.Meta.GetName()).To(Equal(name))
			Expect(res.Meta.GetNamespace()).To(Equal(namespace))
			Expect(res.Meta.GetVersion()).ToNot(BeEmpty())
			Expect(res.Spec).To(Equal(createdResource.Spec))
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			list := sample_model.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list, store.ListByNamespace("non-existent-namespace"), store.ListByMesh(mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(0))
		})

		It("should return a list of resources", func() {
			// given two resources
			createResource("res-1")
			createResource("res-2")

			list := sample_model.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list, store.ListByNamespace(namespace))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(2))
			// and
			names := []string{list.Items[0].Meta.GetName(), list.Items[1].Meta.GetName()}
			Expect(names).To(ConsistOf("res-1", "res-2"))
			Expect(list.Items[0].Meta.GetNamespace()).To(Equal(namespace))
			Expect(list.Items[0].Meta.GetMesh()).To(Equal(mesh))
			Expect(list.Items[0].Spec.Path).To(Equal("demo"))
			Expect(list.Items[1].Meta.GetNamespace()).To(Equal(namespace))
			Expect(list.Items[1].Meta.GetMesh()).To(Equal(mesh))
			Expect(list.Items[1].Spec.Path).To(Equal("demo"))
		})

		It("should not return a list of resources in different namespace", func() {
			// given two resources
			createResource("res-1")
			createResource("res-2")

			list := sample_model.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list, store.ListByNamespace("different-namespace"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(0))
		})

		It("should not return a list of resources in different mesh", func() {
			// given two resources
			createResource("res-1")
			createResource("res-2")

			list := sample_model.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list, store.ListByMesh("different-mesh"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(0))
		})
	})
}
