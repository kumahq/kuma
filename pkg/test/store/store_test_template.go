package store

import (
	"context"
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	. "github.com/kumahq/kuma/pkg/test/matchers"
)

func ExecuteStoreTests(
	createStore func() store.ResourceStore,
	storeName string,
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
		list := core_mesh.TrafficRouteResourceList{}
		err := s.List(context.Background(), &list)
		Expect(err).ToNot(HaveOccurred())
		for _, item := range list.Items {
			err := s.Delete(context.Background(), item, store.DeleteByKey(item.Meta.GetName(), item.Meta.GetMesh()))
			Expect(err).ToNot(HaveOccurred())
		}
	})

	createResource := func(name string, keyAndValues ...string) *core_mesh.TrafficRouteResource {
		res := core_mesh.TrafficRouteResource{
			Spec: &v1alpha1.TrafficRoute{
				Conf: &v1alpha1.TrafficRoute_Conf{
					Destination: map[string]string{
						"path": "demo",
					},
				},
			},
		}
		labels := map[string]string{}
		for i := 0; i < len(keyAndValues); i += 2 {
			labels[keyAndValues[i]] = keyAndValues[i+1]
		}

		err := s.Create(context.Background(), &res, store.CreateByKey(name, mesh),
			store.CreatedAt(time.Now()),
			store.CreateWithLabels(labels))
		Expect(err).ToNot(HaveOccurred())
		return &res
	}

	Context("Store: "+storeName, func() {
		Describe("Create()", func() {
			It("should create a new resource", func() {
				// given
				name := "resource1.demo"

				// when
				created := createResource(name, "foo", "bar")

				// when retrieve created object
				resource := core_mesh.NewTrafficRouteResource()
				err := s.Get(context.Background(), resource, store.GetByKey(name, mesh))

				// then
				Expect(err).ToNot(HaveOccurred())

				// and it has same data
				Expect(resource.Meta.GetName()).To(Equal(name))
				Expect(resource.Meta.GetMesh()).To(Equal(mesh))
				Expect(resource.Meta.GetVersion()).ToNot(BeEmpty())
				Expect(resource.Meta.GetCreationTime()).ToNot(BeZero())
				Expect(resource.Meta.GetCreationTime()).To(Equal(resource.Meta.GetModificationTime()))
				Expect(resource.Meta.GetLabels()).To(HaveKeyWithValue("foo", "bar"))
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
				resource := createResource(name, "foo", "bar")
				modificationTime := time.Now().Add(time.Second)
				versionBeforeUpdate := resource.Meta.GetVersion()

				// when
				resource.Spec.Conf.Destination["path"] = "new-path"
				newLabels := map[string]string{
					"foo":      "barbar",
					"newlabel": "newvalue",
				}
				err := s.Update(context.Background(), resource, store.ModifiedAt(modificationTime), store.UpdateWithLabels(newLabels))

				// then
				Expect(err).ToNot(HaveOccurred())

				// and meta is updated (version and modification time)
				Expect(resource.Meta.GetVersion()).ToNot(Equal(versionBeforeUpdate))
				Expect(resource.Meta.GetLabels()).To(And(HaveKeyWithValue("foo", "barbar"), HaveKeyWithValue("newlabel", "newvalue")))
				if reflect.TypeOf(createStore()) != reflect.TypeOf(&resources_k8s.KubernetesStore{}) {
					Expect(resource.Meta.GetModificationTime().Round(time.Millisecond).Nanosecond() / 1e6).To(Equal(modificationTime.Round(time.Millisecond).Nanosecond() / 1e6))
				}

				// when retrieve the resource
				res := core_mesh.NewTrafficRouteResource()
				err = s.Get(context.Background(), res, store.GetByKey(name, mesh))

				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				Expect(res.Spec.Conf.Destination["path"]).To(Equal("new-path"))
				Expect(resource.Meta.GetLabels()).To(And(HaveKeyWithValue("foo", "barbar"), HaveKeyWithValue("newlabel", "newvalue")))

				// and modification time is updated
				// on K8S modification time is always the creation time, because there is no data for modification time
				if reflect.TypeOf(createStore()) == reflect.TypeOf(&resources_k8s.KubernetesStore{}) {
					Expect(res.Meta.GetModificationTime()).To(Equal(res.Meta.GetCreationTime()))
				} else {
					Expect(res.Meta.GetModificationTime()).ToNot(Equal(res.Meta.GetCreationTime()))
					Expect(res.Meta.GetModificationTime().Round(time.Millisecond).Nanosecond() / 1e6).To(Equal(modificationTime.Round(time.Millisecond).Nanosecond() / 1e6))
				}
			})

<<<<<<< HEAD
=======
			It("should preserve labels", func() {
				// given
				name := "to-be-updated.demo"
				resource := createResource(name, "foo", "bar")

				// when
				resource.Spec.Conf.Destination["path"] = "new-path"
				err := s.Update(context.Background(), resource)

				// then
				Expect(err).ToNot(HaveOccurred())

				res := core_mesh.NewTrafficRouteResource()
				err = s.Get(context.Background(), res, store.GetByKey(name, mesh))
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Meta.GetLabels()).To(HaveKeyWithValue("foo", "bar"))
			})

			It("should delete labels", func() {
				// given a resources in storage
				name := "to-be-updated.demo"
				resource := createResource(name, "foo", "bar")

				// when
				resource.Spec.Conf.Destination["path"] = "new-path"
				err := s.Update(context.Background(), resource, store.UpdateWithLabels(map[string]string{}))

				// then
				Expect(err).ToNot(HaveOccurred())

				res := core_mesh.NewTrafficRouteResource()
				err = s.Get(context.Background(), res, store.GetByKey(name, mesh))
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Meta.GetLabels()).ToNot(HaveKeyWithValue("foo", "bar"))
			})

			It("should update resource with status", func() {
				// given
				updated := meshservice_api.MeshServiceResource{
					Spec: &meshservice_api.MeshService{
						Selector: meshservice_api.Selector{
							DataplaneTags: map[string]string{
								"a": "b",
							},
						},
						Ports: []meshservice_api.Port{
							{
								Port:       80,
								TargetPort: intstr.FromInt(80),
								Protocol:   "http",
							},
						},
					},
					Status: &meshservice_api.MeshServiceStatus{
						VIPs: []meshservice_api.VIP{
							{
								IP: "10.0.0.1",
							},
						},
					},
				}
				err := s.Create(context.Background(), &updated, store.CreateByKey("ms-2.demo", mesh))
				Expect(err).ToNot(HaveOccurred())

				// when
				updated.Status.VIPs[0].IP = "10.0.0.2"
				updated.Spec.Ports[0].Port = 81
				err = s.Update(context.Background(), &updated)

				// then
				Expect(err).ToNot(HaveOccurred())

				// and
				ms := meshservice_api.NewMeshServiceResource()
				err = s.Get(context.Background(), ms, store.GetByKey("ms-2.demo", mesh))
				Expect(err).ToNot(HaveOccurred())
				Expect(ms.Status).To(Equal(updated.Status))
				Expect(ms.Spec).To(Equal(updated.Spec))
			})

>>>>>>> b0abc25a4 (feat(store): update does not wipe out labels (#10335))
			// todo(jakubdyszkiewicz) write tests for optimistic locking
		})

		Describe("Delete()", func() {
			It("should throw an error if resource is not found", func() {
				// given
				name := "non-existent-name.demo"
				resource := core_mesh.NewTrafficRouteResource()

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
				getResource := core_mesh.NewTrafficRouteResource()
				err = s.Get(context.Background(), getResource, store.GetByKey(name, mesh))

				// then resource still exists
				Expect(err).ToNot(HaveOccurred())
			})

			It("should delete an existing resource", func() {
				// given a resources in storage
				name := "to-be-deleted.demo"
				createResource(name)

				// when
				resource := core_mesh.NewTrafficRouteResource()
				err := s.Delete(context.TODO(), resource, store.DeleteByKey(name, mesh))

				// then
				Expect(err).ToNot(HaveOccurred())

				// when query for deleted resource
				resource = core_mesh.NewTrafficRouteResource()
				err = s.Get(context.Background(), resource, store.GetByKey(name, mesh))

				// then resource cannot be found
				Expect(err).To(Equal(store.ErrorResourceNotFound(resource.Descriptor().Name, name, mesh)))
			})
		})

		Describe("Get()", func() {
			It("should return an error if resource is not found", func() {
				// given
				name := "non-existing-resource.demo"
				resource := core_mesh.NewTrafficRouteResource()

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
				resource := core_mesh.NewTrafficRouteResource()
				err := s.Get(context.Background(), resource, store.GetByKey(name, mesh))

				// then
				Expect(err).To(Equal(store.ErrorResourceNotFound(resource.Descriptor().Name, name, mesh)))
			})

			It("should return an existing resource", func() {
				// given a resources in storage
				name := "get-existing-resource.demo"
				createdResource := createResource(name)

				// when
				res := core_mesh.NewTrafficRouteResource()
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
				err := s.Get(context.Background(), core_mesh.NewTrafficRouteResource(), store.GetByKey(name, mesh), store.GetByVersion(res.GetMeta().GetVersion()))

				// then resource is found
				Expect(err).ToNot(HaveOccurred())

				// when trying to retrieve resource with different version
				err = s.Get(context.Background(), core_mesh.NewTrafficRouteResource(), store.GetByKey(name, mesh), store.GetByVersion("9999999"))

				// then resource precondition failed error occurred
				Expect(err).Should(MatchError(&store.ResourceConflictError{}))
			})
		})

		Describe("List()", func() {
			It("should return an empty list if there are no matching resources", func() {
				// given
				list := core_mesh.TrafficRouteResourceList{}

				// when
				err := s.List(context.Background(), &list, store.ListByMesh(mesh))

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(list.Pagination.Total).To(Equal(uint32(0)))
				// and
				Expect(list.Items).To(BeEmpty())
			})

			It("should return a list of resources", func() {
				// given two resources
				createResource("res-1.demo")
				createResource("res-2.demo")

				list := core_mesh.TrafficRouteResourceList{}

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
				Expect(list.Items[0].Spec.Conf.Destination["path"]).To(Equal("demo"))
				Expect(list.Items[1].Meta.GetMesh()).To(Equal(mesh))
				Expect(list.Items[1].Spec.Conf.Destination["path"]).To(Equal("demo"))
			})

			It("should not return a list of resources in different mesh", func() {
				// given two resources
				createResource("list-res-1.demo")
				createResource("list-res-2.demo")

				list := core_mesh.TrafficRouteResourceList{}

				// when
				err := s.List(context.Background(), &list, store.ListByMesh("different-mesh"))

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(list.Pagination.Total).To(Equal(uint32(0)))
				// and
				Expect(list.Items).To(BeEmpty())
			})

			It("should return a list of resources with prefix from all meshes", func() {
				// given two resources
				createResource("list-res-1.demo")
				createResource("list-res-2.demo")
				createResource("list-mes-1.demo")

				list := core_mesh.TrafficRouteResourceList{}

				// when
				err := s.List(context.Background(), &list, store.ListByNameContains("list-res"))

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(list.Pagination.Total).To(Equal(uint32(2)))
				// and
				Expect(list.Items).To(WithTransform(func(itms []*core_mesh.TrafficRouteResource) []string {
					var res []string
					for _, v := range itms {
						res = append(res, v.GetMeta().GetName())
					}
					return res
				}, Equal([]string{"list-res-1.demo", "list-res-2.demo"})))
			})

			It("should return a list of resources with prefix from the specific mesh", func() {
				// given two resources
				createResource("list-res-1.demo")
				createResource("list-res-2.demo")
				createResource("list-mes-1.demo")

				list := core_mesh.TrafficRouteResourceList{}

				// when
				err := s.List(context.Background(), &list, store.ListByNameContains("list-res"), store.ListByMesh(mesh))

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(list.Pagination.Total).To(Equal(uint32(2)))
				// and
				Expect(list.Items).To(WithTransform(func(itms []*core_mesh.TrafficRouteResource) []string {
					var res []string
					for _, v := range itms {
						res = append(res, v.GetMeta().GetName())
					}
					return res
				}, Equal([]string{"list-res-1.demo", "list-res-2.demo"})))
			})

			It("should return a list of 2 resources by resource key", func() {
				// given two resources
				createResource("list-res-1.demo")
				createResource("list-res-2.demo")
				rs3 := createResource("list-mes-1.demo")
				rs4 := createResource("list-mes-1.default")

				list := core_mesh.TrafficRouteResourceList{}
				rk := []core_model.ResourceKey{core_model.MetaToResourceKey(rs3.GetMeta()), core_model.MetaToResourceKey(rs4.GetMeta())}

				// when
				err := s.List(context.Background(), &list, store.ListByResourceKeys(rk))

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(list.Pagination.Total).To(Equal(uint32(2)))
				// and
				Expect(list.Items).To(WithTransform(func(itms []*core_mesh.TrafficRouteResource) []string {
					var res []string
					for _, v := range itms {
						res = append(res, v.GetMeta().GetName())
					}
					return res
				}, Equal([]string{"list-mes-1.default", "list-mes-1.demo"})))
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
						list := core_mesh.TrafficRouteResourceList{}
						err := s.List(context.Background(), &list, store.ListByMesh(mesh), store.ListByPage(pageSize, offset))

						Expect(err).ToNot(HaveOccurred())
						Expect(list.Pagination.NextOffset).ToNot(BeEmpty())
						Expect(list.Items).To(HaveLen(2))

						resourceNames[list.Items[0].GetMeta().GetName()] = true
						resourceNames[list.Items[1].GetMeta().GetName()] = true
						offset = list.Pagination.NextOffset
					}

					// when list third page with 1 element (less than page size)
					list := core_mesh.TrafficRouteResourceList{}
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
					list := core_mesh.TrafficRouteResourceList{}
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
					list := core_mesh.TrafficRouteResourceList{}
					err := s.List(context.Background(), &list, store.ListByMesh(mesh), store.ListByPage(1, ""))

					// then
					Expect(list.Pagination.Total).To(Equal(uint32(1)))
					Expect(list.Items).To(HaveLen(1))
					Expect(err).ToNot(HaveOccurred())
					Expect(list.Pagination.NextOffset).To(BeEmpty())
				})

				It("next offset should be null when queried empty collection", func() {
					// when
					list := core_mesh.TrafficRouteResourceList{}
					err := s.List(context.Background(), &list, store.ListByMesh("unknown-mesh"), store.ListByPage(2, ""))

					// then
					Expect(list.Pagination.Total).To(Equal(uint32(0)))
					Expect(list.Items).To(BeEmpty())
					Expect(err).ToNot(HaveOccurred())
					Expect(list.Pagination.NextOffset).To(BeEmpty())
				})

				It("next offset should return error when query with invalid offset", func() {
					// when
					list := core_mesh.TrafficRouteResourceList{}
					err := s.List(context.Background(), &list, store.ListByMesh("unknown-mesh"), store.ListByPage(2, "123invalidOffset"))

					// then
					Expect(list.Pagination.Total).To(Equal(uint32(0)))
					Expect(err).To(Equal(store.ErrorInvalidOffset))
				})
			})
		})
	})
}
