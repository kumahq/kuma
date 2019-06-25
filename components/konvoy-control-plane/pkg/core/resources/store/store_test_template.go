package store

import (
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// It takes a pointer to ResourceStore because with sample usage
//
// var _ = Describe("MemoryStore", func() {
//	var c store.ResourceStore
//
//	BeforeEach(func() {
//		c = store.NewStrictResourceStore(memory.NewStore())
//	})
//
//	store.ExecuteStoreTests(&c)
// })
//
// when calling ExecuteStorageTest, the `c` is not yet initialized by BeforeEach
func ExecuteStoreTests(
	createStore func() ResourceStore,
) {
	const namespace = "default"
	var s ClosableResourceStore

	BeforeEach(func() {
		s = NewStrictResourceStore(createStore())
	})

	AfterEach(func() {
		err := s.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	createResource := func(name string) *TrafficRouteResource {
		res := TrafficRouteResource{
			Spec: TrafficRoute{
				Path: "demo",
			},
		}
		err := s.Create(context.Background(), &res, CreateByName(namespace, name))
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
			resource := TrafficRouteResource{}
			err := s.Get(context.Background(), &resource, GetByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and it has same data
			Expect(resource.Meta.GetName()).To(Equal(name))
			Expect(resource.Meta.GetNamespace()).To(Equal(namespace))
			Expect(resource.Meta.GetVersion()).ToNot(BeEmpty())
			Expect(resource.Spec).To(Equal(created.Spec))
		})

		It("should not create a duplicate record", func() {
			// given
			name := "duplicated-record"
			resource := createResource(name)

			// when try to create another one with same name
			resource.SetMeta(nil)
			err := s.Create(context.Background(), resource, CreateByName(namespace, name))

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource already exists: type="TrafficRoute" namespace="%s" name="%s"`, namespace, name)))
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
				DeleteByName(resource.GetMeta().GetNamespace(), resource.Meta.GetName()),
			)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when trying to update nonexistent resource
			err = s.Update(context.Background(), resource)

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource not found: type="TrafficRoute" namespace="%s" name="%s"`, namespace, name)))
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
			res := TrafficRouteResource{}
			err = s.Get(context.Background(), &res, GetByName(namespace, name))

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
			resource := TrafficRouteResource{}

			// when
			err := s.Delete(context.TODO(), &resource, DeleteByName(namespace, "non-existent-name"))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// given a resources in storage
			name := "to-be-deleted"
			createResource(name)

			// when
			resource := TrafficRouteResource{}
			err := s.Delete(context.TODO(), &resource, DeleteByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource
			err = s.Get(context.Background(), &resource, GetByName(namespace, name))

			// then resource cannot be found
			Expect(err).To(Equal(ErrorResourceNotFound(resource.GetType(), namespace, name)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			resource := TrafficRouteResource{}

			// when
			err := s.Get(context.Background(), &resource, GetByName(namespace, "non-existing-resource"))

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource not found: type="TrafficRoute" namespace="%s" name="non-existing-resource"`, namespace)))
		})

		It("should return an existing resource", func() {
			// given a resources in storage
			name := "existing-resource"
			createdResource := createResource(name)

			// when
			res := TrafficRouteResource{}
			err := s.Get(context.Background(), &res, GetByName(namespace, name))

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
			list := TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list, ListByNamespace("non-existent-namespace"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(0))
		})

		It("should return a list of matching resource", func() {
			// given two resources
			createResource("res-1")
			createResource("res-2")

			list := TrafficRouteResourceList{}

			// when
			err := s.List(context.Background(), &list, ListByNamespace(namespace))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(2))
			// and
			names := []string { list.Items[0].Meta.GetName(), list.Items[1].Meta.GetName() }
			Expect(names).To(ConsistOf("res-1", "res-2"))
			Expect(list.Items[0].Meta.GetNamespace()).To(Equal(namespace))
			Expect(list.Items[0].Spec.Path).To(Equal("demo"))
			Expect(list.Items[1].Meta.GetNamespace()).To(Equal(namespace))
			Expect(list.Items[1].Spec.Path).To(Equal("demo"))
		})
	})
}

const (
	TrafficRouteType model.ResourceType = "TrafficRoute"
)

// TODO(jakubdyszkiewicz): delete after introducing protobuffs
type TrafficRoute struct {
	Path string `json:"path"`
}

var _ model.Resource = &TrafficRouteResource{}

type TrafficRouteResource struct {
	Meta model.ResourceMeta
	Spec TrafficRoute
}

func (t *TrafficRouteResource) GetType() model.ResourceType {
	return TrafficRouteType
}
func (t *TrafficRouteResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *TrafficRouteResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *TrafficRouteResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}

var _ model.ResourceList = &TrafficRouteResourceList{}

type TrafficRouteResourceList struct {
	Items []*TrafficRouteResource
}

func (l *TrafficRouteResourceList) GetItemType() model.ResourceType {
	return TrafficRouteType
}
func (l *TrafficRouteResourceList) NewItem() model.Resource {
	return &TrafficRouteResource{}
}
func (l *TrafficRouteResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*TrafficRouteResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*TrafficRouteResource)(nil), r)
	}
}