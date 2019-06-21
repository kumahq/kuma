package postgres

import (
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("postgresResourceStore", func() {
	var s store.ResourceStore
	const namespace = "default"

	createResource := func(name string) *TrafficRouteResource {
		res := TrafficRouteResource{
			Spec: TrafficRoute{
				Path: "path-123",
			},
		}
		err := s.Create(context.Background(), &res, store.CreateByName(namespace, name))
		Expect(err).ToNot(HaveOccurred())
		return &res
	}

	BeforeEach(func() {
		config := Config{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "mysecretpassword",
			DbName:   "konvoy",
		}
		resStore, err := NewStore(config)
		if err != nil {
			panic(err)
		}
		s = resStore

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
			name := "resource1"

			// when
			created := createResource(name)

			// when retrieve created object
			resource := TrafficRouteResource{}
			err := s.Get(context.Background(), &resource, store.GetByName(namespace, name))
			Expect(err).ToNot(HaveOccurred())

			// then it has same data
			Expect(resource.Meta.GetName()).To(Equal(name))
			Expect(resource.Meta.GetNamespace()).To(Equal(namespace))
			//Expect(resource.Meta.GetVersion()).To(Equal("0"))
			Expect(resource.Spec).To(Equal(created.Spec))
		})

		It("should not create a duplicate record", func() {
			// given
			name := "duplicated-record"
			created := createResource(name)

			// when try to create another one with same name
			created.SetMeta(nil)
			err := s.Create(context.Background(), created, store.CreateByName(namespace, name))

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource already exists: type="TrafficRoute" namespace="%s" name="%s"`, namespace, name)))
		})
	})

	Describe("Update()", func() {
		It("should return an error if resource is not found", func() {
			// when
			updateResource := &TrafficRouteResource{
				Meta: &resourceMetaObject{
					Namespace: namespace,
					Name:      "example",
				},
			}
			err := s.Update(context.Background(), updateResource)

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource not found: type="TrafficRoute" namespace="%s" name="example"`, namespace)))
		})

		It("should update an existing resource", func() {
			// given a resources in storage
			name := "to-be-updated"
			createResource := createResource(name)

			// when
			createResource.Spec.Path = "new-path"
			err := s.Update(context.Background(), createResource)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when retrieved object again again
			res := TrafficRouteResource{}
			err = s.Get(context.Background(), &res, store.GetByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Spec.Path).To(Equal("new-path"))
		})
	})

	Describe("Delete()", func() {
		It("should succeed if resource is not found", func() {
			// when
			res := TrafficRouteResource{}
			err := s.Delete(context.TODO(), &res, store.DeleteByName(namespace, "non-existent-name"))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// given a resources in storage
			name := "to-be-deleted"
			createResource(name)

			// when
			res := TrafficRouteResource{}
			err := s.Delete(context.TODO(), &res, store.DeleteByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource
			err = s.Get(context.Background(), &res, store.GetByName(namespace, name))

			// then resource cannot be found
			Expect(err).To(Equal(store.ErrorResourceNotFound(res.GetType(), namespace, name)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			res := TrafficRouteResource{}

			// when
			err := s.Get(context.Background(), &res, store.GetByName(namespace, "non-existing-resource"))

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource not found: type="TrafficRoute" namespace="%s" name="non-existing-resource"`, namespace)))
		})

		It("should return an existing resource", func() {
			// given a resources in storage
			name := "existing-resource"
			createResource := createResource(name)

			// when
			res := TrafficRouteResource{}
			err := s.Get(context.Background(), &res, store.GetByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(res.Meta.GetName()).To(Equal(name))
			Expect(res.Meta.GetNamespace()).To(Equal(namespace))
			//Expect(resource.Meta.GetVersion()).To(Equal("0"))
			Expect(res.Spec).To(Equal(createResource.Spec))
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			list := TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list, store.ListByNamespace("non-existent-namespace"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(0))
		})

		It("should return a list of matching resource", func() {
			// given two resources
			createResource("res-1")
			createResource("res-2")

			list := TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list, store.ListByNamespace(namespace))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(2))

			Expect(list.Items[0].GetMeta().GetName()).To(Equal("res-1"))
			Expect(list.Items[1].GetMeta().GetName()).To(Equal("res-2"))
		})
	})
})
