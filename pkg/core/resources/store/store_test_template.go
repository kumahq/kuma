package store

import (
	"context"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	sample_proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
	sample_model "github.com/Kong/kuma/pkg/test/resources/apis/sample"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

func ExecuteStoreTests(
	createStore func() ResourceStore,
) {
	var namespace string
	const mesh = "default-mesh"
	var s ClosableResourceStore

	BeforeEach(func() {
		namespace = string(uuid.New())
		s = NewStrictResourceStore(createStore())
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
		err := s.Create(context.Background(), &res, CreateByKey(namespace, name, mesh))
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
			err := s.Get(context.Background(), &resource, GetByKey(namespace, name, mesh))

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
			err := s.Create(context.Background(), resource, CreateByKey(namespace, name, mesh))

			// then
			Expect(err).To(MatchError(ErrorResourceAlreadyExists(resource.GetType(), namespace, name, mesh)))
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
				DeleteByKey(resource.GetMeta().GetNamespace(), resource.Meta.GetName(), mesh),
			)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when trying to update nonexistent resource
			err = s.Update(context.Background(), resource)

			// then
			Expect(err).To(MatchError(ErrorResourceConflict(resource.GetType(), namespace, name, mesh)))
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
			err = s.Get(context.Background(), &res, GetByKey(namespace, name, mesh))

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
			err := s.Delete(context.TODO(), &resource, DeleteByKey(namespace, "non-existent-name", mesh))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not delete resource from another mesh", func() {
			// given
			name := "tr-1"
			resource := createResource(name)

			// when
			resource.SetMeta(nil) // otherwise the validation from strict client fires that mesh is different
			err := s.Delete(context.TODO(), resource, DeleteByKey(namespace, name, "different-mesh"))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and when getting the given resource
			getResource := sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &getResource, GetByKey(namespace, name, mesh))

			// then resource still exists
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// given a resources in storage
			name := "to-be-deleted"
			createResource(name)

			// when
			resource := sample_model.TrafficRouteResource{}
			err := s.Delete(context.TODO(), &resource, DeleteByKey(namespace, name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource
			resource = sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &resource, GetByKey(namespace, name, mesh))

			// then resource cannot be found
			Expect(err).To(Equal(ErrorResourceNotFound(resource.GetType(), namespace, name, mesh)))
		})
	})

	Describe("DeleteMany()", func() {
		BeforeEach(func() {
			trRes := sample_model.TrafficRouteResource{
				Spec: sample_proto.TrafficRoute{
					Path: "demo",
				},
			}
			err := s.Create(context.Background(), &trRes, CreateByKey(namespace, "tr-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())

			dpRes := core_mesh.DataplaneResource{
				Spec: mesh_proto.Dataplane{},
			}
			err = s.Create(context.Background(), &dpRes, CreateByKey(namespace, "dp-1", "mesh-2"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete all resources", func() {
			// when
			err := s.DeleteMany(context.Background())

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource
			resource := sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &resource, GetByKey(namespace, "tr-1", "mesh-1"))

			// then resource cannot be found
			Expect(err).To(Equal(ErrorResourceNotFound(resource.GetType(), namespace, "tr-1", "mesh-1")))

			// when query for deleted resource
			dpResource := core_mesh.DataplaneResource{}
			err = s.Get(context.Background(), &dpResource, GetByKey(namespace, "dp-1", "mesh-2"))

			// then resource cannot be found
			Expect(err).To(Equal(ErrorResourceNotFound(dpResource.GetType(), namespace, "dp-1", "mesh-2")))
		})

		It("should delete resources by mesh", func() {
			// when
			err := s.DeleteMany(context.Background(), DeleteManyByMesh("mesh-1"))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource in given mesh
			resource := sample_model.TrafficRouteResource{}
			err = s.Get(context.Background(), &resource, GetByKey(namespace, "tr-1", "mesh-1"))

			// then resource cannot be found
			Expect(err).To(Equal(ErrorResourceNotFound(resource.GetType(), namespace, "tr-1", "mesh-1")))

			// when query for resource in another mesh
			dpResource := core_mesh.DataplaneResource{}
			err = s.Get(context.Background(), &dpResource, GetByKey(namespace, "dp-1", "mesh-2"))

			// then resource is not deleted
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			name := "non-existing-resource"
			resource := sample_model.TrafficRouteResource{}

			// when
			err := s.Get(context.Background(), &resource, GetByKey(namespace, name, mesh))

			// then
			Expect(err).To(MatchError(ErrorResourceNotFound(resource.GetType(), namespace, name, mesh)))
		})

		It("should return an error if resource is not found in given mesh", func() {
			// given a resources in mesh "mesh"
			name := "existing-resource"
			mesh := "different-mesh"
			createResource(name)

			// when
			resource := sample_model.TrafficRouteResource{}
			err := s.Get(context.Background(), &resource, GetByKey(namespace, name, mesh))

			// then
			Expect(err).To(Equal(ErrorResourceNotFound(resource.GetType(), namespace, name, mesh)))
		})

		It("should return an existing resource", func() {
			// given a resources in storage
			name := "existing-resource"
			createdResource := createResource(name)

			// when
			res := sample_model.TrafficRouteResource{}
			err := s.Get(context.Background(), &res, GetByKey(namespace, name, mesh))

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
			err := s.List(context.Background(), &list, ListByNamespace("non-existent-namespace"), ListByMesh(mesh))

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
			err := s.List(context.Background(), &list, ListByNamespace(namespace))

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
			err := s.List(context.Background(), &list, ListByNamespace("different-namespace"))

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
			err := s.List(context.Background(), &list, ListByMesh("different-mesh"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(0))
		})
	})
}
