package metrics_test

import (
	"context"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	store_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metered Store", func() {

	var metrics core_metrics.Metrics
	var store core_store.ResourceStore

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics()
		metrics = m
		Expect(err).ToNot(HaveOccurred())
		memoryStore := store_memory.NewStore()
		store, err = core_metrics.NewMeteredStore(memoryStore, metrics)
		Expect(err).ToNot(HaveOccurred())

		// setup test data
		err = memoryStore.Create(context.Background(), &core_mesh.MeshResource{}, core_store.CreateByKey("default", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should public metrics of GET", func() {
		// when
		err := store.Get(context.Background(), &core_mesh.MeshResource{}, core_store.GetByKey("default", "default"))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(test_metrics.FindMetric(metrics, "store", "operation", "get", "resource_type", "Mesh").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
	})

	It("should public metrics of LIST", func() {
		// when
		err := store.List(context.Background(), &core_mesh.MeshResourceList{}, core_store.ListByMesh("default"))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(test_metrics.FindMetric(metrics, "store", "operation", "list", "resource_type", "Mesh").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
	})

	It("should public metrics of DELETE", func() {
		// when
		err := store.Delete(context.Background(), &core_mesh.MeshResource{}, core_store.DeleteByKey("default", "default"))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(test_metrics.FindMetric(metrics, "store", "operation", "delete", "resource_type", "Mesh").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
	})

	It("should public metrics of UPDATE", func() {
		// when
		mesh := &core_mesh.MeshResource{}
		err := store.Get(context.Background(), mesh, core_store.GetByKey("default", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = store.Update(context.Background(), mesh)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(test_metrics.FindMetric(metrics, "store", "operation", "get", "resource_type", "Mesh").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
		Expect(test_metrics.FindMetric(metrics, "store", "operation", "update", "resource_type", "Mesh").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
	})
})
