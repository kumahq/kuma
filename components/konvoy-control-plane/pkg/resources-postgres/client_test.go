package resources_postgres

import (
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/client"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/client/example"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math/rand"
)

var _ = Describe("ResourceClient", func() {
	var c client.ResourceClient
	const namespace = "default"

	//createResource := func(name string) {
	//	createResource := example.TrafficRouteResource{
	//		Spec: example.TrafficRoute{
	//			Path: "path-123",
	//		},
	//	}
	//	err := c.Create(context.Background(), &createResource, client.CreateByName(namespace, name))
	//	Expect(err).ToNot(HaveOccurred())
	//}

	BeforeEach(func() {
		config := PostgresConfig{
			Host: "localhost",
			Port: 5432,
			User: "postgres",
			Password: "mysecretpassword",
			DbName: "konvoy",
		}
		resClient, err := NewResourceClient(config)
		if err != nil {
			panic(err)
		}
		c = resClient

		// wipe DB
		db, err := connectToDb(config)
		if err != nil {
			panic(err)
		}
		_, err = db.Exec("DELETE FROM resources")
		if err != nil {
			panic(err)
		}
	})

	Describe("Create()", func() {
		It("should create a new resource", func() {
			// given
			name := fmt.Sprintf("tr-%d", rand.Int())
			createResource := example.TrafficRouteResource{
				Spec: example.TrafficRoute{
					Path: "path-123",
				},
			}

			// when
			err := c.Create(context.Background(), &createResource, client.CreateByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when retrieve created object
			resource := example.TrafficRouteResource{}
			err = c.Get(context.Background(), &resource, client.GetByName(namespace, name))
			Expect(err).ToNot(HaveOccurred())

			// then it has same data
			Expect(resource.Meta.GetName()).To(Equal(name))
			Expect(resource.Meta.GetNamespace()).To(Equal(namespace))
			//Expect(resource.Meta.GetVersion()).To(Equal("0"))
			Expect(resource.Spec).To(Equal(createResource.Spec))
		})

		It("should not create a duplicate record", func() {
			// given
			name := fmt.Sprintf("tr-%d", rand.Int())

			// and resource in DB
			createResource := example.TrafficRouteResource{
				Spec: example.TrafficRoute{
					Path: "path-123",
				},
			}
			err := c.Create(context.Background(), &createResource, client.CreateByName(namespace, name))
			Expect(err).ToNot(HaveOccurred())

			// when try to create another one with same name
			createResource.SetMeta(nil)
			err = c.Create(context.Background(), &createResource, client.CreateByName(namespace, name))

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource already exists: type="TrafficRoute" namespace="%s" name="%s"`, namespace, name)))
		})
	})

	Describe("Update()", func() {
		It("should return an error if resource is not found", func() {
			// when
			updateResource := &example.TrafficRouteResource{
				Meta: &resourceMetaObject{
					Namespace: namespace,
					Name:      "example",
				},
			}
			err := c.Update(context.Background(), updateResource)

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource not found: type="TrafficRoute" namespace="%s" name="example"`, namespace)))
		})

		It("should update an existing resource", func() {
			// given a resources in storage
			name := "to-be-updated"
			createResource := &example.TrafficRouteResource{
				Spec: example.TrafficRoute{
					Path: "path-123",
				},
			}

			err := c.Create(context.Background(), createResource, client.CreateByName(namespace, name))
			Expect(err).ToNot(HaveOccurred())

			// when
			createResource.Spec.Path = "new-path"
			err = c.Update(context.Background(), createResource)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when retrieved object again again
			res := example.TrafficRouteResource{}
			err = c.Get(context.Background(), &res, client.GetByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Spec.Path).To(Equal("new-path"))
		})
	})

	Describe("Delete()", func() {
		It("should succeed if resource is not found", func() {
			// when
			res := example.TrafficRouteResource{}
			err := c.Delete(context.TODO(), &res, client.DeleteByName(namespace, "non-existent-name"))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// given a resources in storage
			name := "to-be-deleted"
			createResource := example.TrafficRouteResource{
				Spec: example.TrafficRoute{
					Path: "path-123",
				},
			}

			err := c.Create(context.Background(), &createResource, client.CreateByName(namespace, name))
			Expect(err).ToNot(HaveOccurred())

			// when
			res := example.TrafficRouteResource{}
			err = c.Delete(context.TODO(), &res, client.DeleteByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource
			err = c.Get(context.Background(), &res, client.GetByName(namespace, name))

			// then resource cannot be found
			Expect(err).To(Equal(client.ErrorResourceNotFound(res.GetType(), namespace, name)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			res := example.TrafficRouteResource{}

			// when
			err := c.Get(context.Background(), &res, client.GetByName(namespace, "non-existing-resource"))

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource not found: type="TrafficRoute" namespace="%s" name="non-existing-resource"`, namespace)))
		})

		It("should return an existing resource", func() {
			// given a resources in storage
			createResource := example.TrafficRouteResource{
				Spec: example.TrafficRoute{
					Path: "path-123",
				},
			}

			err := c.Create(context.Background(), &createResource, client.CreateByName(namespace, "res1"))
			Expect(err).ToNot(HaveOccurred())

			// when
			res := example.TrafficRouteResource{}
			err = c.Get(context.Background(), &res, client.GetByName(namespace, "res1"))

			// then
			Expect(res.Meta.GetName()).To(Equal("res1"))
			Expect(res.Meta.GetNamespace()).To(Equal(namespace))
			//Expect(resource.Meta.GetVersion()).To(Equal("0"))
			Expect(res.Spec).To(Equal(createResource.Spec))
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			list := example.TrafficRouteResourceList{}

			// when
			err := c.List(context.Background(), &list, client.ListByNamespace("non-existent-namespace"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(0))
		})

		It("should return a list of matching resource", func() {
			// given two resources
			createResource := example.TrafficRouteResource{
				Spec: example.TrafficRoute{
					Path: "path-123",
				},
			}

			err := c.Create(context.Background(), &createResource, client.CreateByName(namespace, "res1"))
			Expect(err).ToNot(HaveOccurred())

			createResource.SetMeta(nil)
			err = c.Create(context.Background(), &createResource, client.CreateByName(namespace, "res2"))
			Expect(err).ToNot(HaveOccurred())

			list := example.TrafficRouteResourceList{}

			// when
			err = c.List(context.Background(), &list, client.ListByNamespace(namespace))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(2))

			Expect(list.Items[0].GetMeta().GetName()).To(Equal("res1"))
			Expect(list.Items[1].GetMeta().GetName()).To(Equal("res2"))
		})
	})


})
