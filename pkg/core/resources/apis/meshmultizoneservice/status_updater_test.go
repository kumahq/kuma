package meshmultizoneservice_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
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

		updater, err := meshmultizoneservice.NewStatusUpdater(logr.Discard(), resManager, resManager, 50*time.Millisecond, m)
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

	It("should add mesh services and port to the status of multizone service", func() {
		// when
		ms1Builder := samples.MeshServiceBackendBuilder().
			WithName("backend").
			WithDataplaneTagsSelector(map[string]string{
				"app": "backend",
			}).
			WithLabels(map[string]string{
				mesh_proto.DisplayName: "backend",
				mesh_proto.ZoneTag:     "east",
			})
		Expect(ms1Builder.Create(resManager)).To(Succeed())
		Expect(samples.MeshServiceWebBuilder().Create(resManager)).To(Succeed()) // to check if we ignore it
		Expect(samples.MeshMultiZoneServiceBackendBuilder().Create(resManager)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			mzsvc := meshmzservice_api.NewMeshMultiZoneServiceResource()

			err := resManager.Get(context.Background(), mzsvc, store.GetByKey("backend", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(mzsvc.Status.MeshServices).To(Equal([]meshmzservice_api.MatchedMeshService{{Name: "backend"}}))
			g.Expect(mzsvc.Status.Ports).To(Equal(ms1Builder.Build().Spec.Ports))
		}, "10s", "100ms").Should(Succeed())

		// when new service is added
		ms2Builder := samples.MeshServiceBackendBuilder().
			WithName("backend-syncedhash").
			WithDataplaneTagsSelector(map[string]string{
				"app": "backend",
			}).
			AddIntPort(71, 8081, core_mesh.ProtocolHTTP).
			WithLabels(map[string]string{
				mesh_proto.DisplayName: "backend",
				mesh_proto.ZoneTag:     "west",
			})
		Expect(ms2Builder.Create(resManager)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			mzsvc := meshmzservice_api.NewMeshMultiZoneServiceResource()

			err := resManager.Get(context.Background(), mzsvc, store.GetByKey("backend", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(mzsvc.Status.MeshServices).To(Equal([]meshmzservice_api.MatchedMeshService{
				{Name: "backend"},
				{Name: "backend-syncedhash"},
			}))
			// ports are sorted
			g.Expect(mzsvc.Status.Ports).To(Equal([]meshservice_api.Port{
				ms2Builder.Build().Spec.Ports[1],
				ms1Builder.Build().Spec.Ports[0],
			}))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_mzms_status_updater")).ToNot(BeNil())
		}, "10s", "100ms").Should(Succeed())
	})
})
