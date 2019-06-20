package example_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/client"
	. "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/client/example"
)

var _ = Describe("ExampleResourceClient", func() {
	var e *ExampleResourceClient
	var c client.ResourceClient

	BeforeEach(func() {
		e = &ExampleResourceClient{}
		c = client.NewStrictResourceClient(e)
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
			err := c.Create(context.Background(), trr, client.CreateByName("demo", "example"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(e.PersistedRecords).To(ConsistOf(
				&ExampleStorageRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
			))
		})

		It("should not create a duplicate record", func() {
			// setup
			e.PersistedRecords = ExampleStorageRecords{
				&ExampleStorageRecord{
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
			err := c.Create(context.Background(), trr, client.CreateByName("demo", "example"))

			// then
			Expect(err).To(MatchError(`Resource already exists: type="TrafficRoute" namespace="demo" name="example"`))
		})
	})

	Describe("Update()", func() {
		It("should return an error if resource is not found", func() {
			// given
			trr := &TrafficRouteResource{
				Meta: &ExampleMeta{
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
			e.PersistedRecords = ExampleStorageRecords{
				&ExampleStorageRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
			}

			// given
			trr := &TrafficRouteResource{
				Meta: &ExampleMeta{
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
			Expect(e.PersistedRecords).To(ConsistOf(
				&ExampleStorageRecord{
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
			err := c.Delete(context.Background(), trr, client.DeleteByName("demo", "app"))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete an existing resource", func() {
			// setup
			e.PersistedRecords = ExampleStorageRecords{
				&ExampleStorageRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
			}

			// given
			trr := &TrafficRouteResource{}

			// when
			err := c.Delete(context.Background(), trr, client.DeleteByName("demo", "example"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(e.PersistedRecords).To(HaveLen(0))
		})
	})

	Describe("Get()", func() {
		It("should return an error if resource is not found", func() {
			// given
			trr := &TrafficRouteResource{}

			// when
			err := c.Get(context.Background(), trr, client.GetByName("demo", "example"))

			// then
			Expect(err).To(MatchError(`Resource not found: type="TrafficRoute" namespace="demo" name="example"`))
		})

		It("should return an existing resource", func() {
			// setup
			e.PersistedRecords = ExampleStorageRecords{
				&ExampleStorageRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
			}

			// given
			trr := &TrafficRouteResource{}

			// when
			err := c.Get(context.Background(), trr, client.GetByName("demo", "example"))

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
			err := c.List(context.Background(), trr, client.ListByNamespace("demo"))

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(trr.Items).To(HaveLen(0))
		})

		It("should return a list of matching resource", func() {
			// setup
			e.PersistedRecords = ExampleStorageRecords{
				&ExampleStorageRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "example",
					Spec:         `{"path":"/example"}`,
				},
				&ExampleStorageRecord{
					ResourceType: "TrafficRoute",
					Namespace:    "demo",
					Name:         "another",
					Spec:         `{"path":"/another"}`,
				},
			}

			// given
			trr := &TrafficRouteResourceList{}

			// when
			err := c.List(context.Background(), trr, client.ListByNamespace("demo"))

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
