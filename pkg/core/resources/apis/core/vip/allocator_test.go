package vip_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	meshextenralservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

var _ = Describe("VIP Allocator", func() {
	var stopCh chan struct{}
	var resManager manager.ResourceManager
	var metrics core_metrics.Metrics

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		metrics = m
		resManager = manager.NewResourceManager(memory.NewStore())

		allocator, err := vip.NewAllocator(
			logr.Discard(),
			50*time.Millisecond,
			map[string]model.ResourceTypeDescriptor{
				"241.0.0.0/8": meshservice_api.MeshServiceResourceTypeDescriptor,
				"242.0.0.0/8": meshextenralservice_api.MeshExternalServiceResourceTypeDescriptor,
			},
			metrics,
			resManager,
		)

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

	vipOfMeshService := func(name string) string {
		ms := meshservice_api.NewMeshServiceResource()
		err := resManager.Get(context.Background(), ms, store.GetByKey(name, model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		if len(ms.Status.VIPs) == 0 {
			return ""
		}
		return ms.Status.VIPs[0].IP
	}

	It("should allocate vip for MeshService without vip", func() {
		// when
		err := samples.MeshServiceBackendBuilder().WithoutVIP().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			g.Expect(vipOfMeshService("backend")).Should(Equal("241.0.0.0"))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should allocate vip for MeshExternalService without vip in different CIDR", func() {
		// when
		err := samples.MeshExternalServiceExampleBuilder().WithoutVIP().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			mes := meshextenralservice_api.NewMeshExternalServiceResource()
			err := resManager.Get(context.Background(), mes, store.GetByKey("example", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(mes.Status.VIP.IP).To(Equal("242.0.0.0"))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should not reuse IPs", func() {
		// given
		err := samples.MeshServiceBackendBuilder().WithoutVIP().Create(resManager)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			g.Expect(vipOfMeshService("backend")).Should(Equal("241.0.0.0"))
		}, "10s", "100ms").Should(Succeed())

		// when resource is reapplied
		err = resManager.Delete(context.Background(), meshservice_api.NewMeshServiceResource(), store.DeleteByKey("backend", model.DefaultMesh))
		Expect(err).ToNot(HaveOccurred())
		err = samples.MeshServiceBackendBuilder().WithoutVIP().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			g.Expect(vipOfMeshService("backend")).Should(Equal("241.0.0.1"))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_vip_allocator")).ToNot(BeNil())
		}, "10s", "100ms").Should(Succeed())
	})
})
