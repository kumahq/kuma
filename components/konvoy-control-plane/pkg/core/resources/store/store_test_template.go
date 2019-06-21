package store

import (
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func ExecuteStoreTests(s *ResourceStore) {
	const namespace = "default"

	createResource := func(name string) *TrafficRouteResource {
		res := TrafficRouteResource{
			Spec: TrafficRoute{
				Path: "demo",
			},
		}
		err := (*s).Create(context.Background(), &res, CreateByName(namespace, name))
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
			err := (*s).Get(context.Background(), &resource, GetByName(namespace, name))

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
			created := createResource(name)

			// when try to create another one with same name
			created.SetMeta(nil)
			err := (*s).Create(context.Background(), created, CreateByName(namespace, name))

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource already exists: type="TrafficRoute" namespace="%s" name="%s"`, namespace, name)))
		})
	})

	Describe("Update()", func() {
		It("should return an error if resource is not found", func() {
			// given
			updateResource := &TrafficRouteResource{
				Meta: &resourceMetaObject{
					Namespace: namespace,
					Name:      "example",
				},
			}

			// when
			err := (*s).Update(context.Background(), updateResource)

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource not found: type="TrafficRoute" namespace="%s" name="example"`, namespace)))
		})

		It("should update an existing resource", func() {
			// given a resources in storage
			name := "to-be-updated"
			createResource := createResource(name)

			// when
			createResource.Spec.Path = "new-path"
			err := (*s).Update(context.Background(), createResource)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when retrieve the resource
			res := TrafficRouteResource{}
			err = (*s).Get(context.Background(), &res, GetByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(res.Spec.Path).To(Equal("new-path"))
		})
	})

	Describe("Delete()", func() {
		It("should succeed if resource is not found", func() {
			// given
			res := TrafficRouteResource{}

			// when
			err := (*s).Delete(context.TODO(), &res, DeleteByName(namespace, "non-existent-name"))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// given a resources in storage
			name := "to-be-deleted"
			createResource(name)

			// when
			res := TrafficRouteResource{}
			err := (*s).Delete(context.TODO(), &res, DeleteByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when query for deleted resource
			err = (*s).Get(context.Background(), &res, GetByName(namespace, name))

			// then resource cannot be found
			Expect(err).To(Equal(ErrorResourceNotFound(res.GetType(), namespace, name)))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			res := TrafficRouteResource{}

			// when
			err := (*s).Get(context.Background(), &res, GetByName(namespace, "non-existing-resource"))

			// then
			Expect(err).To(MatchError(fmt.Sprintf(`Resource not found: type="TrafficRoute" namespace="%s" name="non-existing-resource"`, namespace)))
		})

		It("should return an existing resource", func() {
			// given a resources in storage
			name := "existing-resource"
			createResource := createResource(name)

			// when
			res := TrafficRouteResource{}
			err := (*s).Get(context.Background(), &res, GetByName(namespace, name))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(res.Meta.GetName()).To(Equal(name))
			Expect(res.Meta.GetNamespace()).To(Equal(namespace))
			Expect(res.Meta.GetVersion()).ToNot(BeEmpty())
			Expect(res.Spec).To(Equal(createResource.Spec))
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			list := TrafficRouteResourceList{}

			// when
			err := (*s).List(context.Background(), &list, ListByNamespace("non-existent-namespace"))

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
			err := (*s).List(context.Background(), &list, ListByNamespace(namespace))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(list.Items).To(HaveLen(2))
			// and
			Expect(list.Items[0].Meta.GetName()).To(Equal("res-1"))
			Expect(list.Items[0].Meta.GetNamespace()).To(Equal(namespace))
			Expect(list.Items[0].Spec.Path).To(Equal("demo"))
			// and
			Expect(list.Items[1].Meta.GetName()).To(Equal("res-2"))
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

type resourceMetaObject struct {
	Name      string
	Namespace string
	Version   string
}

var _ model.ResourceMeta = &resourceMetaObject{}

func (r *resourceMetaObject) GetName() string {
	return r.Name
}

func (r *resourceMetaObject) GetNamespace() string {
	return r.Namespace
}

func (r *resourceMetaObject) GetVersion() string {
	return r.Version
}