package store

import (
	"context"
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/store"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	sample_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	sample_model "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
)

func ExecuteStoreTests(
	createStore func() store.ResourceStore,
) {
	const mesh = "default-mesh"
	var s store.ClosableResourceStore

	BeforeEach(func() {
		s = store.NewStrictResourceStore(store.NewPaginationStore(createStore()))
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
			Spec: &sample_proto.TrafficRoute{
				Path: "demo",
			},
		}
		err := s.Create(context.Background(), &res, store.CreateByKey(name, mesh), store.CreatedAt(time.Now()))
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
			resource := sample_model.NewTrafficRouteResource()
			err := s.Get(context.Background(), resource, store.GetByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and it has same data
			Expect(resource.Meta.GetName()).To(Equal(name))
			Expect(resource.Meta.GetMesh()).To(Equal(mesh))
			Expect(resource.Meta.GetVersion()).ToNot(BeEmpty())
			Expect(resource.Meta.GetCreationTime().Unix()).ToNot(Equal(0))
			Expect(resource.Meta.GetCreationTime()).To(Equal(resource.Meta.GetModificationTime()))
			Expect(resource.Spec).To(MatchProto(created.Spec))
		})

		It("should not create a duplicate record", func() {
			// given
			name := "duplicated-record.demo"
			resource := createResource(name)

			// when try to create another one with same name
			resource.SetMeta(nil)
			err := s.Create(context.Background(), resource, store.CreateByKey(name, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceAlreadyExists(resource.Descriptor().Name, name, mesh)))
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
			Expect(err).To(MatchError(store.ErrorResourceConflict(resource.Descriptor().Name, name, mesh)))
		})

		It("should update an existing resource", func() {
			// given a resources in storage
			name := "to-be-updated.demo"
			resource := createResource(name)
			modificationTime := time.Now().Add(time.Second)
			versionBeforeUpdate := resource.Meta.GetVersion()

			// when
			resource.Spec.Path = "new-path"
			err := s.Update(context.Background(), resource, store.ModifiedAt(modificationTime))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and meta is updated (version and modification time)
			Expect(resource.Meta.GetVersion()).ToNot(Equal(versionBeforeUpdate))
			if reflect.TypeOf(createStore()) != reflect.TypeOf(&resources_k8s.KubernetesStore{}) {
				Expect(resource.Meta.GetModificationTime().Round(time.Millisecond).Nanosecond() / 1e6).To(Equal(modificationTime.Round(time.Millisecond).Nanosecond() / 1e6))
			}

			// when retrieve the resource
			res := sample_model.NewTrafficRouteResource()
			err = s.Get(context.Background(), res, store.GetByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(res.Spec.Path).To(Equal("new-path"))

			// and modification time is updated
			// on K8S modification time is always the creation time, because there is no data for modification time
			if reflect.TypeOf(createStore()) == reflect.TypeOf(&resources_k8s.KubernetesStore{}) {
				Expect(res.Meta.GetModificationTime()).To(Equal(res.Meta.GetCreationTime()))
			} else {
				Expect(res.Meta.GetModificationTime()).ToNot(Equal(res.Meta.GetCreationTime()))
				Expect(res.Meta.GetModificationTime().Round(time.Millisecond).Nanosecond() / 1e6).To(Equal(modificationTime.Round(time.Millisecond).Nanosecond() / 1e6))
			}
		})

		// todo(jakubdyszkiewicz) write tests for optimistic locking
	})

	Describe("Delete()", func() {
		It("should throw an error if resource is not found", func() {
			// given
			name := "non-existent-name.demo"
			resource := sample_model.NewTrafficRouteResource()

			// when
			err := s.Delete(context.TODO(), resource, store.DeleteByKey(name, mesh))

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.Descriptor().Name, name, mesh)))
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
			getResource := sample_model.NewTrafficRouteResource()
			err = s.Get(context.Background(), getResource, store.GetByKey(name, mesh))

			// then resource still exists
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// given a resources in storage
			name := "to-be-deleted.demo"
			createResource(name)

			// when
			resource := sample_model.NewTrafficRouteResource()
			err := s.Delete(context.TODO(), resource, store.DeleteByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource
			resource = sample_model.NewTrafficRouteResource()
			err = s.Get(context.Background(), resource, store.GetByKey(name, mesh))

			// then resource cannot be found
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.Descriptor().Name, name, mesh)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			name := "non-existing-resource.demo"
			resource := sample_model.NewTrafficRouteResource()

			// when
			err := s.Get(context.Background(), resource, store.GetByKey(name, mesh))

			// then
			Expect(err).To(MatchError(store.ErrorResourceNotFound(resource.Descriptor().Name, name, mesh)))
		})

		It("should return an error if resource is not found in given mesh", func() {
			// given a resources in mesh "mesh"
			name := "existing-resource.demo"
			mesh := "different-mesh"
			createResource(name)

			// when
			resource := sample_model.NewTrafficRouteResource()
			err := s.Get(context.Background(), resource, store.GetByKey(name, mesh))

			// then
			Expect(err).To(Equal(store.ErrorResourceNotFound(resource.Descriptor().Name, name, mesh)))
		})

		It("should return an existing resource", func() {
			// given a resources in storage
			name := "get-existing-resource.demo"
			createdResource := createResource(name)

			// when
			res := sample_model.NewTrafficRouteResource()
			err := s.Get(context.Background(), res, store.GetByKey(name, mesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(res.Meta.GetName()).To(Equal(name))
			Expect(res.Meta.GetVersion()).ToNot(BeEmpty())
			Expect(res.Spec).To(MatchProto(createdResource.Spec))
		})

		It("should get resource by version", func() {
			// given
			name := "existing-resource.demo"
			res := createResource(name)

			// when trying to retrieve resource with proper version
			err := s.Get(context.Background(), sample_model.NewTrafficRouteResource(), store.GetByKey(name, mesh), store.GetByVersion(res.GetMeta().GetVersion()))

			// then resource is found
			Expect(err).ToNot(HaveOccurred())

			// when trying to retrieve resource with different version
			err = s.Get(context.Background(), sample_model.NewTrafficRouteResource(), store.GetByKey(name, mesh), store.GetByVersion("9999999"))

			// then resource precondition failed error occurred
			Expect(store.IsResourcePreconditionFailed(err)).To(BeTrue())
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
			Expect(list.Pagination.Total).To(Equal(uint32(0)))
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
			Expect(list.Pagination.Total).To(Equal(uint32(2)))
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
			Expect(list.Pagination.Total).To(Equal(uint32(0)))
			// and
			Expect(list.Items).To(HaveLen(0))
		})

		Describe("Pagination", func() {
			It("should list all resources using pagination", func() {
				// given
				offset := ""
				pageSize := 2
				numOfResources := 5
				resourceNames := map[string]bool{}

				// setup create resources
				for i := 0; i < numOfResources; i++ {
					createResource(fmt.Sprintf("res-%d.demo", i))
				}

				// when list first two pages with 2 elements
				for i := 1; i <= 2; i++ {
					list := sample_model.TrafficRouteResourceList{}
					err := s.List(context.Background(), &list, store.ListByMesh(mesh), store.ListByPage(pageSize, offset))

					Expect(err).ToNot(HaveOccurred())
					Expect(list.Pagination.NextOffset).ToNot(BeEmpty())
					Expect(list.Items).To(HaveLen(2))

					resourceNames[list.Items[0].GetMeta().GetName()] = true
					resourceNames[list.Items[1].GetMeta().GetName()] = true
					offset = list.Pagination.NextOffset
				}

				// when list third page with 1 element (less than page size)
				list := sample_model.TrafficRouteResourceList{}
				err := s.List(context.Background(), &list, store.ListByMesh(mesh), store.ListByPage(pageSize, offset))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(list.Pagination.Total).To(Equal(uint32(numOfResources)))
				Expect(list.Pagination.NextOffset).To(BeEmpty())
				Expect(list.Items).To(HaveLen(1))
				resourceNames[list.Items[0].GetMeta().GetName()] = true

				// and all elements were retrieved
				Expect(resourceNames).To(HaveLen(numOfResources))
				for i := 0; i < numOfResources; i++ {
					Expect(resourceNames).To(HaveKey(fmt.Sprintf("res-%d.demo", i)))
				}
			})

			It("next offset should be null when queried collection with less elements than page has", func() {
				// setup
				createResource("res-1.demo")

				// when
				list := sample_model.TrafficRouteResourceList{}
				err := s.List(context.Background(), &list, store.ListByMesh(mesh), store.ListByPage(5, ""))

				// then
				Expect(list.Pagination.Total).To(Equal(uint32(1)))
				Expect(list.Items).To(HaveLen(1))
				Expect(err).ToNot(HaveOccurred())
				Expect(list.Pagination.NextOffset).To(BeEmpty())
			})

			It("next offset should be null when queried about size equals to elements available", func() {
				// setup
				createResource("res-1.demo")

				// when
				list := sample_model.TrafficRouteResourceList{}
				err := s.List(context.Background(), &list, store.ListByMesh(mesh), store.ListByPage(1, ""))

				// then
				Expect(list.Pagination.Total).To(Equal(uint32(1)))
				Expect(list.Items).To(HaveLen(1))
				Expect(err).ToNot(HaveOccurred())
				Expect(list.Pagination.NextOffset).To(BeEmpty())
			})

			It("next offset should be null when queried empty collection", func() {
				// when
				list := sample_model.TrafficRouteResourceList{}
				err := s.List(context.Background(), &list, store.ListByMesh("unknown-mesh"), store.ListByPage(2, ""))

				// then
				Expect(list.Pagination.Total).To(Equal(uint32(0)))
				Expect(list.Items).To(BeEmpty())
				Expect(err).ToNot(HaveOccurred())
				Expect(list.Pagination.NextOffset).To(BeEmpty())
			})

			It("next offset should return error when query with invalid offset", func() {
				// when
				list := sample_model.TrafficRouteResourceList{}
				err := s.List(context.Background(), &list, store.ListByMesh("unknown-mesh"), store.ListByPage(2, "123invalidOffset"))

				// then
				Expect(list.Pagination.Total).To(Equal(uint32(0)))
				Expect(err).To(Equal(store.ErrorInvalidOffset))
			})
		})
	})
}
