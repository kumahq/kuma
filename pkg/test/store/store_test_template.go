package store

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/store"
	sample_proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	sample_model "github.com/Kong/kuma/pkg/test/resources/apis/sample"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func ExecuteStoreTests(
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

	BeforeEach(func() {
		list := sample_model.TrafficRouteResourceList{}
		err := s.List(context.Background(), &list)
		Expect(err).ToNot(HaveOccurred())
		for _, item := range list.Items {
			err := s.Delete(context.Background(), item, store.DeleteByKey(item.Meta.GetName(), item.Meta.GetMesh()))
			Expect(err).ToNot(HaveOccurred())
		}
	})

	createResource := func(name string) *sample_model.TrafficRouteResource {
		res := sample_model.TrafficRouteResource{
			Spec: sample_proto.TrafficRoute{
				Path: "demo",
			},
		}
		err := s.Create(context.Background(), &res, store.CreateByKey(name, mesh))
		Expect(err).ToNot(HaveOccurred())
		return &res
	}

	Describe("Create()", func() {
		It("should create a new resource", func() {
			// given
			name := "resource1.demo"

			// when
			created := createResource(name)

			// when retrieve created object
			resource := sample_model.TrafficRouteResource{}
			err := s.Get(context.Background(), &resource, store.GetByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and it has same data
			Expect(resource.Meta.GetName()).To(Equal(name))
			Expect(resource.Meta.GetMesh()).To(Equal(mesh))
			Expect(resource.Meta.GetVersion()).ToNot(BeEmpty())
			Expect(resource.Spec).To(Equal(created.Spec))
		})

		It("should not create a duplicate record", func() {
			// given
			name := "duplicated-record.demo"
			resource := createResource(name)

			// when try to create another one with same name
			resource.SetMeta(nil)
			err := s.Create(context.Background(), resource, store.CreateByKey(name, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceAlreadyExists(resource.GetType(), name, mesh)))
		})
	})

	Describe("Update()", func() {
		It("should return an error if resource is not found", func() {
			// given
			name := "to-be-updated.demo"
			resource := createResource(name)

			// when delete resource
			err := s.Delete(
				context.Background(),
				resource,
				store.DeleteByKey(resource.Meta.GetName(), mesh),
			)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when trying to update nonexistent resource
			err = s.Update(context.Background(), resource)

			// then
			Expect(err).To(MatchError(store.ErrorResourceConflict(resource.GetType(), name, mesh)))
		})

		It("should update an existing resource", func() {
			// given a resources in storage
			name := "to-be-updated.demo"
			resource := createResource(name)

			// when
			resource.Spec.Path = "new-path"
			err := s.Update(context.Background(), resource)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when retrieve the resource
			res := sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &res, store.GetByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(res.Spec.Path).To(Equal("new-path"))
		})

		//todo(jakubdyszkiewicz) write tests for optimistic locking
	})

	Describe("Delete()", func() {
		It("should throw an error if resource is not found", func() {
			// given
			name := "non-existent-name.demo"
			resource := sample_model.TrafficRouteResource{}

			// when
			err := s.Delete(context.TODO(), &resource, store.DeleteByKey(name, mesh))

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.GetType(), name, mesh)))
		})

		It("should not delete resource from another mesh", func() {
			// given
			name := "tr-1.demo"
			resource := createResource(name)

			// when
			resource.SetMeta(nil) // otherwise the validation from strict client fires that mesh is different
			err := s.Delete(context.TODO(), resource, store.DeleteByKey(name, "different-mesh"))

			// then
			Expect(err).To(HaveOccurred())
			Expect(store.IsResourceNotFound(err)).To(BeTrue())

			// and when getting the given resource
			getResource := sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &getResource, store.GetByKey(name, mesh))

			// then resource still exists
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// given a resources in storage
			name := "to-be-deleted.demo"
			createResource(name)

			// when
			resource := sample_model.TrafficRouteResource{}
			err := s.Delete(context.TODO(), &resource, store.DeleteByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource
			resource = sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &resource, store.GetByKey(name, mesh))

			// then resource cannot be found
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.GetType(), name, mesh)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			name := "non-existing-resource.demo"
			resource := sample_model.TrafficRouteResource{}

			// when
			err := s.Get(context.Background(), &resource, store.GetByKey(name, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(resource.GetType(), name, mesh)))
		})

		It("should return an error if resource is not found in given mesh", func() {
			// given a resources in mesh "mesh"
			name := "existing-resource.demo"
			mesh := "different-mesh"
			createResource(name)

			// when
			resource := sample_model.TrafficRouteResource{}
			err := s.Get(context.Background(), &resource, store.GetByKey(name, mesh))

			// then
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.GetType(), name, mesh)))
		})

		It("should return an existing resource", func() {
			// given a resources in storage
			name := "get-existing-resource.demo"
			createdResource := createResource(name)

			// when
			res := sample_model.TrafficRouteResource{}
			err := s.Get(context.Background(), &res, store.GetByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(res.Meta.GetName()).To(Equal(name))
			Expect(res.Meta.GetVersion()).ToNot(BeEmpty())
			Expect(res.Spec).To(Equal(createdResource.Spec))
		})

		It("should get resource by version", func() {
			// given
			name := "existing-resource.demo"
			res := createResource(name)

			// when trying to retrieve resource with proper version
			err := s.Get(context.Background(), &sample_model.TrafficRouteResource{}, store.GetByKey(name, mesh), store.GetByVersion(res.GetMeta().GetVersion()))

			// then resource is found
			Expect(err).ToNot(HaveOccurred())

			// when trying to retrieve resource with different version
			err = s.Get(context.Background(), &sample_model.TrafficRouteResource{}, store.GetByKey(name, mesh), store.GetByVersion("9999999"))

			// then resource is not found
			Expect(store.IsResourceNotFound(err)).To(BeTrue())
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			list := sample_model.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list, store.ListByMesh(mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(0))
		})

		It("should return a list of resources", func() {
			// given two resources
			createResource("res-1.demo")
			createResource("res-2.demo")

			list := sample_model.TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(2))
			// and
			names := []string{list.Items[0].Meta.GetName(), list.Items[1].Meta.GetName()}
			Expect(names).To(ConsistOf("res-1.demo", "res-2.demo"))
			Expect(list.Items[0].Meta.GetMesh()).To(Equal(mesh))
			Expect(list.Items[0].Spec.Path).To(Equal("demo"))
			Expect(list.Items[1].Meta.GetMesh()).To(Equal(mesh))
			Expect(list.Items[1].Spec.Path).To(Equal("demo"))
		})

		It("should not return a list of resources in different mesh", func() {
			// given two resources
			createResource("list-res-1.demo")
			createResource("list-res-2.demo")

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
