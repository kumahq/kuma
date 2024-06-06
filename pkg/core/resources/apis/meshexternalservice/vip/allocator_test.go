package vip

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

var _ = Describe("MeshExternalService VIP Allocator", func() {
	var stopCh chan struct{}
	var resManager manager.ResourceManager
	var metrics core_metrics.Metrics

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		metrics = m
		resManager = manager.NewResourceManager(memory.NewStore())
		externalServiceAllocator, err := NewMeshExternalServiceAllocator(logr.Discard(), "242.0.0.0/8", resManager, 50*time.Millisecond, m)
		Expect(err).ToNot(HaveOccurred())
		allocator, err := vip.NewAllocator(logr.Discard(), 50*time.Millisecond, []vip.VIPAllocator{externalServiceAllocator})
		Expect(err).ToNot(HaveOccurred())
		stopCh = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			Expect(allocator.Start(stopCh)).To(Succeed())
		}()

		Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
	})

	AfterEach(func() {
		close(stopCh)
	})

	vipOfMeshExternalService := func(name string) string {
		mes := meshexternalservice_api.NewMeshExternalServiceResource()
		err := resManager.Get(context.Background(), mes, store.GetByKey(name, model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		if mes.Status.VIP.IP == "" {
			return ""
		}
		return mes.Status.VIP.IP
	}

	It("should allocate vip for external service without vip", func() {
		// when
		err := samples.MeshExternalServiceExampleBuilder().WithoutVIP().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			g.Expect(vipOfMeshExternalService("example")).Should(Equal("242.0.0.0"))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should not reuse IPs", func() {
		// given
		err := samples.MeshExternalServiceExampleBuilder().WithoutVIP().Create(resManager)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			g.Expect(vipOfMeshExternalService("example")).Should(Equal("242.0.0.0"))
		}, "10s", "100ms").Should(Succeed())

		// when resource is reapplied
		err = resManager.Delete(context.Background(), meshexternalservice_api.NewMeshExternalServiceResource(), store.DeleteByKey("example", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		err = samples.MeshExternalServiceExampleBuilder().WithoutVIP().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			g.Expect(vipOfMeshExternalService("example")).Should(Equal("242.0.0.1"))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_mes_vip_allocator")).ToNot(BeNil())
		}, "10s", "100ms").Should(Succeed())
	})
})
