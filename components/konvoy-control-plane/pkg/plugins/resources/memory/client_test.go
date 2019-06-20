package memory_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
)

var _ = Describe("MemoryStore", func() {
	var e *memory.MemoryStore
	var c store.ResourceStore

	BeforeEach(func() {
		e = &memory.MemoryStore{}
		c = store.NewStrictResourceStore(e)
	})

	Describe("Create()", func() {
		It("should create a new resource", func() {
			// given
			trr := &TrafficRouteResource{
				Spec: TrafficRoute{
					Path: "/example",
				},
			}

			// when
			err := c.Create(context.Background(), trr, store.CreateByName("demo", "example"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(e.Records).To(ConsistOf(
				&memory.MemoryStoreRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
			))
		})

		It("should not create a duplicate record", func() {
			// setup
			e.Records = memory.MemoryStoreRecords{
				&memory.MemoryStoreRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
			}

			// given
			trr := &TrafficRouteResource{
				Spec: TrafficRoute{
					Path: "/example",
				},
			}

			// when
			err := c.Create(context.Background(), trr, store.CreateByName("demo", "example"))

			// then
			Expect(err).To(MatchError(`Resource already exists: type="TrafficRoute" namespace="demo" name="example"`))
		})
	})

	Describe("Update()", func() {
		It("should return an error if resource is not found", func() {
			// given
			trr := &TrafficRouteResource{
				Meta: &memory.MemoryMeta{
					Namespace: "demo",
					Name:      "example",
				},
			}

			// when
			err := c.Update(context.Background(), trr)

			// then
			Expect(err).To(MatchError(`Resource not found: type="TrafficRoute" namespace="demo" name="example"`))
		})

		It("should update an existing resource", func() {
			// setup
			e.Records = memory.MemoryStoreRecords{
				&memory.MemoryStoreRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
			}

			// given
			trr := &TrafficRouteResource{
				Meta: &memory.MemoryMeta{
					Namespace: "demo",
					Name:      "example",
				},
				Spec: TrafficRoute{
					Path: "/another",
				},
			}

			// when
			err := c.Update(context.Background(), trr)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(e.Records).To(ConsistOf(
				&memory.MemoryStoreRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/another"}`,
				},
			))
		})
	})

	Describe("Delete()", func() {
		It("should succeed if resource is not found", func() {
			// given
			trr := &TrafficRouteResource{}

			// when
			err := c.Delete(context.Background(), trr, store.DeleteByName("demo", "app"))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// setup
			e.Records = memory.MemoryStoreRecords{
				&memory.MemoryStoreRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
			}

			// given
			trr := &TrafficRouteResource{}

			// when
			err := c.Delete(context.Background(), trr, store.DeleteByName("demo", "example"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(e.Records).To(HaveLen(0))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			trr := &TrafficRouteResource{}

			// when
			err := c.Get(context.Background(), trr, store.GetByName("demo", "example"))

			// then
			Expect(err).To(MatchError(`Resource not found: type="TrafficRoute" namespace="demo" name="example"`))
		})

		It("should return an existing resource", func() {
			// setup
			e.Records = memory.MemoryStoreRecords{
				&memory.MemoryStoreRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
			}

			// given
			trr := &TrafficRouteResource{}

			// when
			err := c.Get(context.Background(), trr, store.GetByName("demo", "example"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trr.Meta.GetNamespace()).To(Equal("demo"))
			Expect(trr.Meta.GetName()).To(Equal("example"))
			Expect(trr.Spec.Path).To(Equal("/example"))
		})
	})

	Describe("List()", func() {
		It("should return an empty list if there are no matching resources", func() {
			// given
			trr := &TrafficRouteResourceList{}

			// when
			err := c.List(context.Background(), trr, store.ListByNamespace("demo"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trr.Items).To(HaveLen(0))
		})

		It("should return a list of matching resource", func() {
			// setup
			e.Records = memory.MemoryStoreRecords{
				&memory.MemoryStoreRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
				&memory.MemoryStoreRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "another",
					Spec:         `{"path":"/another"}`,
				},
			}

			// given
			trr := &TrafficRouteResourceList{}

			// when
			err := c.List(context.Background(), trr, store.ListByNamespace("demo"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trr.Items).To(HaveLen(2))
			// and
			Expect(trr.Items[0].Meta.GetNamespace()).To(Equal("demo"))
			Expect(trr.Items[0].Meta.GetName()).To(Equal("example"))
			Expect(trr.Items[0].Spec.Path).To(Equal("/example"))
			// and
			Expect(trr.Items[1].Meta.GetNamespace()).To(Equal("demo"))
			Expect(trr.Items[1].Meta.GetName()).To(Equal("another"))
			Expect(trr.Items[1].Spec.Path).To(Equal("/another"))
		})
	})
})
