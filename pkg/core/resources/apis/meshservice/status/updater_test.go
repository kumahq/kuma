package status

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

var _ = Describe("Updater", func() {
	var stopCh chan struct{}
	var resManager manager.ResourceManager
	var metrics core_metrics.Metrics

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		metrics = m
		resManager = manager.NewResourceManager(memory.NewStore())

		updater, err := NewStatusUpdater(logr.Discard(), resManager, resManager, 50*time.Millisecond, m, "east")
		Expect(err).ToNot(HaveOccurred())
		stopCh = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			Expect(updater.Start(stopCh)).To(Succeed())
		}()

		Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
	})

	AfterEach(func() {
		close(stopCh)
	})

	It("should add identity to status of service", func() {
		// when
		Expect(samples.MeshServiceBackendBuilder().Create(resManager)).To(Succeed())
		Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
		Expect(samples.DataplaneWebBuilder().Create(resManager)).To(Succeed()) // identity of web should not be added

		// then
		Eventually(func(g Gomega) {
			ms := meshservice_api.NewMeshServiceResource()
			err := resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(ms.Spec.Identities).To(Equal([]meshservice_api.MeshServiceIdentity{
				{
					Type:  meshservice_api.MeshServiceIdentityServiceTagType,
					Value: "backend",
				},
			}))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should not override identity to status of service from another zone", func() {
		// when
		Expect(samples.MeshServiceBackendBuilder().
			WithLabels(map[string]string{
				v1alpha1.ZoneTag: "west",
			}).
			AddServiceTagIdentity("backend").
			Create(resManager)).To(Succeed())
		// and there are no DPPs. If it was a local service it would have no identities

		// then
		Consistently(func(g Gomega) {
			ms := meshservice_api.NewMeshServiceResource()
			err := resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(ms.Spec.Identities).To(Equal([]meshservice_api.MeshServiceIdentity{
				{
					Type:  meshservice_api.MeshServiceIdentityServiceTagType,
					Value: "backend",
				},
			}))
		}, "1s", "100ms").Should(Succeed())
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_ms_status_updater")).ToNot(BeNil())
		}, "10s", "100ms").Should(Succeed())
	})
})
