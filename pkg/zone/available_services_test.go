package zone_test

import (
	"context"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	cache_mesh "github.com/kumahq/kuma/v3/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/server"
	"github.com/kumahq/kuma/v3/pkg/zone"
)

var _ = Describe("AvailableServices Tracker", func() {
	// meshServices.mode no longer affects the tracker at all: kuma.io/service based
	// dataplane services and legacy ExternalServices are always represented by
	// MeshService/MeshExternalService now, so ZoneIngress.AvailableServices is always
	// left empty regardless of what mode a Mesh declares.
	Context("meshServices.mode explicitly Disabled", func() {
		var resManager manager.ResourceManager
		var meshContextBuilder xds_context.MeshContextBuilder
		var metrics core_metrics.Metrics
		var meshCache *cache_mesh.Cache

		var stop chan struct{}
		var done chan struct{}
		BeforeEach(func() {
			resourceStore := memory.NewStore()
			resManager = manager.NewResourceManager(resourceStore)

			Expect(samples.MeshMTLSBuilder().WithEgressRoutingEnabled().Create(resManager)).To(Succeed())

			meshContextBuilder = xds_context.NewMeshContextBuilder(
				resManager,
				server.MeshResourceTypes(),
				net.LookupIP,
				"zone",
				nil,
			)
			var err error
			metrics, err = core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())

			meshCache, err = cache_mesh.NewCache(
				1*time.Second,
				meshContextBuilder,
				metrics,
			)
			Expect(err).ToNot(HaveOccurred())

			tracker, err := zone.NewZoneAvailableServicesTracker(
				core.Log.WithName("test"),
				metrics,
				resManager,
				meshCache,
				20*time.Millisecond,
				nil,
				"zone",
			)
			Expect(err).ToNot(HaveOccurred())

			stop = make(chan struct{})
			done = make(chan struct{})
			go func() {
				defer GinkgoRecover()
				Expect(tracker.Start(stop)).To(Succeed())
				close(done)
			}()
		})
		AfterEach(func() {
			close(stop)
			Eventually(done).Should(BeClosed())
		})
		It("should not populate AvailableServices from kuma.io/service tags", func() {
			Expect(builders.ZoneIngress().Create(resManager)).To(Succeed())
			Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
			Expect(samples.DataplaneWebBuilder().Create(resManager)).To(Succeed())

			Eventually(func(g Gomega) {
				zi := builders.ZoneIngress().Build()
				g.Expect(resManager.Get(context.Background(), zi, store.GetByKey("zoneingress-1", ""))).To(Succeed())
				g.Expect(zi.Spec.AvailableServices).To(BeEmpty())
			}).Should(Succeed())
		})
	})

	Context("Disabled Available Services", func() {
		var resManager manager.ResourceManager
		var stop chan struct{}

		BeforeEach(func() {
			resourceStore := memory.NewStore()
			resManager = manager.NewResourceManager(resourceStore)
			var err error
			metrics, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())

			Expect(samples.MeshMTLSBuilder().WithEgressRoutingEnabled().Create(resManager)).To(Succeed())

			meshContextBuilder := xds_context.NewMeshContextBuilder(
				resManager,
				server.MeshResourceTypes(),
				net.LookupIP,
				"zone",
				nil,
			)
			meshCache, err := cache_mesh.NewCache(
				1*time.Second,
				meshContextBuilder,
				metrics,
			)
			Expect(err).ToNot(HaveOccurred())

			tracker, err := zone.NewZoneAvailableServicesTracker(
				core.Log.WithName("test"),
				metrics,
				resManager,
				meshCache, // not used when available services are disabled
				20*time.Millisecond,
				nil,
				"zone",
			)
			Expect(err).ToNot(HaveOccurred())

			stop = make(chan struct{})
			go func() {
				defer GinkgoRecover()
				Expect(tracker.Start(stop)).To(Succeed())
			}()
		})

		AfterEach(func() {
			close(stop)
		})

		It("should clear available services list when available services are disabled", func() {
			Expect(builders.ZoneIngress().Create(resManager)).To(Succeed())
			Expect(samples.DataplaneBackendBuilder().Create(resManager)).To(Succeed())
			Expect(samples.DataplaneWebBuilder().Create(resManager)).To(Succeed())

			Eventually(func(g Gomega) {
				zi := builders.ZoneIngress().Build()
				g.Expect(resManager.Get(context.Background(), zi, store.GetByKey("zoneingress-1", ""))).To(Succeed())
				g.Expect(zi.Spec.AvailableServices).To(BeEmpty())
			}).Should(Succeed())
		})
	})
})
