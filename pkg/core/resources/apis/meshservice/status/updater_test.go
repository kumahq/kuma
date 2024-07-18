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
	"github.com/kumahq/kuma/pkg/test/resources/builders"
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

	type testCase struct {
		meshBuilder            *builders.MeshBuilder
		dpInsightIssuedBackend string
		existingTLSStatus      meshservice_api.TLSStatus
		expectedTLSStatus      meshservice_api.TLSStatus
	}

	DescribeTable("mTLS updater",
		func(given testCase) {
			// given
			Expect(given.meshBuilder.WithName("test").Create(resManager)).To(Succeed())
			Expect(samples.DataplaneBackendBuilder().WithMesh("test").Create(resManager)).To(Succeed())
			Expect(samples.DataplaneInsightBackendBuilder().
				WithMesh("test").
				WithMTLSIssuedBackend(given.dpInsightIssuedBackend).
				Create(resManager)).To(Succeed())
			Expect(samples.MeshServiceBackendBuilder().
				WithMesh("test").
				WithTLSStatus(given.existingTLSStatus).
				Create(resManager)).To(Succeed())

			Eventually(func(g Gomega) {
				// when
				ms := meshservice_api.NewMeshServiceResource()
				err := resManager.Get(context.Background(), ms, store.GetByKey("backend", "test"))

				// then
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(ms.Status.TLS.Status).To(Equal(given.expectedTLSStatus))
			}, "10s", "50ms").MustPassRepeatedly(3).Should(Succeed())
			// MustPassRepeatedly is to make sure that when existingTLSStatus == expectedTLSStatus we actually preserve it.
		},
		Entry("should set TLS to NotReady when mTLS is disabled", testCase{
			meshBuilder:            samples.MeshDefaultBuilder(),
			dpInsightIssuedBackend: "builtin-1",
			existingTLSStatus:      meshservice_api.TLSReady,
			expectedTLSStatus:      meshservice_api.TLSNotReady,
		}),
		Entry("should set TLS to Ready when we issued certs to all DPPs", testCase{
			meshBuilder:            samples.MeshMTLSBuilder(),
			dpInsightIssuedBackend: "builtin-1",
			existingTLSStatus:      meshservice_api.TLSNotReady,
			expectedTLSStatus:      meshservice_api.TLSReady,
		}),
		Entry("should set TLS to NotReady when we did not issued certs to all DPPs", testCase{
			meshBuilder:            samples.MeshMTLSBuilder(),
			dpInsightIssuedBackend: "",
			existingTLSStatus:      "",
			expectedTLSStatus:      meshservice_api.TLSNotReady,
		}),
		Entry("should preserve TLS Ready even through we did not issue certs to all DPPs", testCase{
			meshBuilder:            samples.MeshMTLSBuilder(),
			dpInsightIssuedBackend: "",
			existingTLSStatus:      meshservice_api.TLSReady,
			expectedTLSStatus:      meshservice_api.TLSReady,
		}),
	)

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_ms_status_updater")).ToNot(BeNil())
		}, "10s", "100ms").Should(Succeed())
	})
})
