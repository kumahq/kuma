package status_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/status"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/v2/pkg/test/metrics"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
)

var _ = Describe("MeshExternalService status updater", func() {
	var stopCh chan struct{}
	var resManager manager.ResourceManager
	var metrics core_metrics.Metrics

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		metrics = m
		resManager = manager.NewResourceManager(memory.NewStore())

		updater, err := status.NewStatusUpdater(logr.Discard(), resManager, resManager, 50*time.Millisecond, m)
		Expect(err).ToNot(HaveOccurred())
		stopCh = make(chan struct{})
		go func(stopCh chan struct{}) {
			defer GinkgoRecover()
			Expect(updater.Start(stopCh)).To(Succeed())
		}(stopCh)

		Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
	})

	AfterEach(func() {
		close(stopCh)
	})

	It("should set SNICompliant=True for a MES whose SNI fits the DNS limits", func() {
		Expect(builders.MeshExternalService().WithName("ext-backend").Create(resManager)).To(Succeed())

		Eventually(func(g Gomega) {
			mes := meshexternalservice_api.NewMeshExternalServiceResource()
			g.Expect(resManager.Get(context.Background(), mes, store.GetByKey("ext-backend", model.DefaultMesh))).To(Succeed())
			g.Expect(mes.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(common_api.SNICompliantCondition),
				"Status": Equal(kube_meta.ConditionTrue),
				"Reason": Equal(common_api.SNICompliantReason),
			})))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should set SNICompliant=False when the MES name contains a dot", func() {
		Expect(builders.MeshExternalService().WithName("ext.backend").Create(resManager)).To(Succeed())

		Eventually(func(g Gomega) {
			mes := meshexternalservice_api.NewMeshExternalServiceResource()
			g.Expect(resManager.Get(context.Background(), mes, store.GetByKey("ext.backend", model.DefaultMesh))).To(Succeed())
			g.Expect(mes.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(common_api.SNICompliantCondition),
				"Status": Equal(kube_meta.ConditionFalse),
				"Reason": Equal(common_api.SNINotCompliantReason),
			})))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_mes_status_updater")).ToNot(BeNil())
		}, "10s", "100ms").Should(Succeed())
	})
})
