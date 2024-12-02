package generate_test

import (
	"context"
	"net"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/generate"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	cache_mesh "github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/server"
)

var _ = Describe("MeshService generator", func() {
	var stopCh chan struct{}
	var resManager manager.ResourceManager
	var meshContextBuilder xds_context.MeshContextBuilder
	var metrics core_metrics.Metrics

	gracePeriodInterval := 500 * time.Millisecond

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		metrics = m
		store := memory.NewStore()
		resManager = manager.NewResourceManager(store)
		meshContextBuilder = xds_context.NewMeshContextBuilder(
			resManager,
			server.MeshResourceTypes(),
			net.LookupIP,
			"zone",
			vips.NewPersistence(resManager, config_manager.NewConfigManager(store), false),
			".mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
		)
		meshCache, err := cache_mesh.NewCache(
			100*time.Millisecond,
			meshContextBuilder,
			metrics,
		)
		Expect(err).ToNot(HaveOccurred())
		allocator, err := generate.New(
			logr.Discard(),
			50*time.Millisecond,
			gracePeriodInterval,
			metrics,
			resManager,
			meshCache,
			"zone",
		)

		Expect(err).ToNot(HaveOccurred())
		stopCh = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			Expect(allocator.Start(stopCh)).To(Succeed())
		}()

		Expect(
			samples.MeshDefaultBuilder().WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere).Create(resManager),
		).To(Succeed())
	})

	AfterEach(func() {
		close(stopCh)
	})

	It("should generate MeshService from a single Dataplane", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			ms := meshservice_api.NewMeshServiceResource()
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        "80",
					Port:        80,
					TargetPort:  intstr.FromInt(80),
					AppProtocol: core_mesh.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should generate MeshService from a single Dataplane with inbound name", func() {
		err := samples.DataplaneBackendBuilder().WithoutInbounds().
			AddInbound(
				builders.Inbound().
					WithPort(builders.FirstInboundPort).
					WithServicePort(builders.FirstInboundServicePort).
					WithName("main").
					WithTags(map[string]string{mesh_proto.ServiceTag: "backend", mesh_proto.ProtocolTag: "tcp"}),
			).Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			ms := meshservice_api.NewMeshServiceResource()
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        "main",
					Port:        80,
					TargetPort:  intstr.FromInt(80),
					AppProtocol: core_mesh.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not change MeshService if a conflicting Dataplanes appears", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        "80",
					Port:        80,
					TargetPort:  intstr.FromInt(80),
					AppProtocol: core_mesh.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		err = samples.DataplaneBackendBuilder().WithName("dp-2").
			AddInbound(
				builders.Inbound().
					WithPort(81).
					WithServicePort(81).
					WithTags(map[string]string{
						mesh_proto.ServiceTag: "backend",
					}),
			).
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		Consistently(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        "80",
					Port:        80,
					TargetPort:  intstr.FromInt(80),
					AppProtocol: core_mesh.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should allow MeshService to be changed if all Dataplanes change", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        "80",
					Port:        80,
					TargetPort:  intstr.FromInt(80),
					AppProtocol: core_mesh.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		dp := core_mesh.NewDataplaneResource()
		Expect(resManager.Get(context.Background(), dp, store.GetByKey("dp-1", model.DefaultMesh))).To(Succeed())
		dp.Spec.Networking.Inbound[0].Port += 1
		dp.Spec.Networking.Inbound[0].ServicePort += 1
		Expect(resManager.Update(context.Background(), dp)).To(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        "81",
					Port:        81,
					TargetPort:  intstr.FromInt(81),
					AppProtocol: core_mesh.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should eventually delete MeshService if all Dataplanes disappear", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        "80",
					Port:        80,
					TargetPort:  intstr.FromInt(80),
					AppProtocol: core_mesh.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		dp := core_mesh.NewDataplaneResource()
		Expect(resManager.Delete(context.Background(), dp, store.DeleteByKey("dp-1", model.DefaultMesh))).To(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).ToNot(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not delete MeshService not managed by the generator", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		Expect(samples.MeshServiceWebBuilder().Create(resManager)).To(Succeed())

		ms := meshservice_api.NewMeshServiceResource()
		Consistently(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("web", model.DefaultMesh))).To(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not delete MeshService immediately", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        "80",
					Port:        80,
					TargetPort:  intstr.FromInt(80),
					AppProtocol: core_mesh.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		dp := core_mesh.NewDataplaneResource()
		Expect(resManager.Delete(context.Background(), dp, store.DeleteByKey("dp-1", model.DefaultMesh))).To(Succeed())

		labelGracePeriodStartedAt := ""
		// Wait until the MeshService has been marked with grace period start
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.GetMeta().GetLabels()).To(HaveKey(mesh_proto.DeletionGracePeriodStartedLabel))
			labelGracePeriodStartedAt = ms.GetMeta().GetLabels()[mesh_proto.DeletionGracePeriodStartedLabel]
		}, "2s", "100ms").Should(Succeed())

		gracePeriodStartedAt := time.Time{}
		Expect(gracePeriodStartedAt.UnmarshalText([]byte(labelGracePeriodStartedAt))).To(Succeed())

		gracePeriodEndsAt := gracePeriodStartedAt.Add(gracePeriodInterval)
		// Before the grace period it still exists and afterwards it eventually
		// disappears
		Consistently(func(g Gomega) {
			err := resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))
			if time.Now().Before(gracePeriodEndsAt) {
				g.Expect(err).To(Succeed())
			}
		}, time.Until(gracePeriodEndsAt.Add(-50*time.Millisecond)).String(), "50ms").Should(Succeed())
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).ToNot(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not delete MeshService if a Dataplane comes back", func() {
		err := samples.DataplaneBackendBuilder().Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		ms := meshservice_api.NewMeshServiceResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.Spec.Ports).To(Equal([]meshservice_api.Port{
				{
					Name:        "80",
					Port:        80,
					TargetPort:  intstr.FromInt(80),
					AppProtocol: core_mesh.ProtocolTCP,
				},
			}))
		}, "2s", "100ms").Should(Succeed())

		dp := core_mesh.NewDataplaneResource()
		Expect(resManager.Delete(context.Background(), dp, store.DeleteByKey("dp-1", model.DefaultMesh))).To(Succeed())

		// Wait until the MeshService has been marked with grace period start
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
			g.Expect(ms.GetMeta().GetLabels()).To(HaveKey("kuma.io/deletion-grace-period-started-at"))
		}, "2s", "100ms").Should(Succeed())

		Expect(
			samples.DataplaneBackendBuilder().Create(resManager),
		).To(Succeed())

		// The MeshService isn't ever deleted
		Consistently(func(g Gomega) {
			ms := meshservice_api.NewMeshServiceResource()
			g.Expect(resManager.Get(context.Background(), ms, store.GetByKey("backend", model.DefaultMesh))).To(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_meshservice_generator")).ToNot(BeNil())
		}, "2s", "100ms").Should(Succeed())
	})
})
