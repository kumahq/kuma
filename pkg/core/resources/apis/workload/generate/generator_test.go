package generate_test

import (
	"context"
	"net"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_manager "github.com/kumahq/kuma/v2/pkg/core/config/manager"
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/generate"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/dns/vips"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	test_metrics "github.com/kumahq/kuma/v2/pkg/test/metrics"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	cache_mesh "github.com/kumahq/kuma/v2/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	"github.com/kumahq/kuma/v2/pkg/xds/server"
)

var _ = Describe("Workload generator", func() {
	var stopCh chan struct{}
	var resManager manager.ResourceManager
	var metrics core_metrics.Metrics

	gracePeriodInterval := 500 * time.Millisecond

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		metrics = m
		store := memory.NewStore()
		resManager = manager.NewResourceManager(store)
		meshContextBuilder := xds_context.NewMeshContextBuilder(
			resManager,
			server.MeshResourceTypes(),
			net.LookupIP,
			"zone",
			vips.NewPersistence(resManager, config_manager.NewConfigManager(store), false),
			".mesh",
			80,
			xds_context.AnyToAnyReachableServicesGraphBuilder,
			nil,
		)
		meshCache, err := cache_mesh.NewCache(
			100*time.Millisecond,
			meshContextBuilder,
			metrics,
		)
		Expect(err).ToNot(HaveOccurred())
		generator, err := generate.New(
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
			Expect(generator.Start(stopCh)).To(Succeed())
		}()

		Expect(
			samples.MeshDefaultBuilder().Create(resManager),
		).To(Succeed())
	})

	AfterEach(func() {
		close(stopCh)
	})

	It("should generate Workload from a single Dataplane with workload label", func() {
		dp := builders.Dataplane().
			WithAddress("192.168.0.1").
			WithServices("backend").
			Build()
		err := resManager.Create(context.Background(), dp,
			store.CreateBy(model.MetaToResourceKey(dp.GetMeta())),
			store.CreateWithLabels(map[string]string{
				metadata.KumaWorkload: "test-workload",
			}))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			wl := workload_api.NewWorkloadResource()
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).To(Succeed())
			g.Expect(wl.GetMeta().GetLabels()[mesh_proto.ManagedByLabel]).To(Equal("workload-generator"))
			g.Expect(wl.GetMeta().GetLabels()[mesh_proto.EnvTag]).To(Equal(mesh_proto.UniversalEnvironment))
			g.Expect(wl.GetMeta().GetLabels()[mesh_proto.ZoneTag]).To(Equal("zone"))
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not generate Workload from Dataplane without workload label", func() {
		err := builders.Dataplane().
			WithAddress("192.168.0.1").
			WithServices("backend").
			Create(resManager)
		Expect(err).ToNot(HaveOccurred())

		Consistently(func(g Gomega) {
			wls := &workload_api.WorkloadResourceList{}
			g.Expect(resManager.List(context.Background(), wls)).To(Succeed())
			g.Expect(wls.GetItems()).To(BeEmpty())
		}, "1s", "100ms").Should(Succeed())
	})

	It("should generate single Workload for multiple Dataplanes with same workload label", func() {
		dp1 := builders.Dataplane().
			WithName("dp-1").
			WithAddress("192.168.0.1").
			WithServices("backend").
			Build()
		err := resManager.Create(context.Background(), dp1,
			store.CreateBy(model.MetaToResourceKey(dp1.GetMeta())),
			store.CreateWithLabels(map[string]string{
				metadata.KumaWorkload: "shared-workload",
			}))
		Expect(err).ToNot(HaveOccurred())

		dp2 := builders.Dataplane().
			WithName("dp-2").
			WithAddress("192.168.0.2").
			WithServices("backend").
			Build()
		err = resManager.Create(context.Background(), dp2,
			store.CreateBy(model.MetaToResourceKey(dp2.GetMeta())),
			store.CreateWithLabels(map[string]string{
				metadata.KumaWorkload: "shared-workload",
			}))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			wl := workload_api.NewWorkloadResource()
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("shared-workload", model.DefaultMesh))).To(Succeed())
			g.Expect(wl.GetMeta().GetLabels()[mesh_proto.ManagedByLabel]).To(Equal("workload-generator"))
		}, "2s", "100ms").Should(Succeed())

		// Verify only one Workload exists
		wls := &workload_api.WorkloadResourceList{}
		Expect(resManager.List(context.Background(), wls)).To(Succeed())
		Expect(wls.GetItems()).To(HaveLen(1))
	})

	It("should eventually delete Workload if all Dataplanes disappear", func() {
		dp := builders.Dataplane().
			WithAddress("192.168.0.1").
			WithServices("backend").
			Build()
		err := resManager.Create(context.Background(), dp,
			store.CreateBy(model.MetaToResourceKey(dp.GetMeta())),
			store.CreateWithLabels(map[string]string{
				metadata.KumaWorkload: "test-workload",
			}))
		Expect(err).ToNot(HaveOccurred())

		wl := workload_api.NewWorkloadResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).To(Succeed())
			g.Expect(wl.GetMeta().GetLabels()[mesh_proto.ManagedByLabel]).To(Equal("workload-generator"))
		}, "2s", "100ms").Should(Succeed())

		// Delete the dataplane
		Expect(resManager.Delete(context.Background(), dp, store.DeleteByKey("dp-1", model.DefaultMesh))).To(Succeed())

		// Workload should eventually be deleted after grace period
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).ToNot(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should not delete Workload immediately", func() {
		dp := builders.Dataplane().
			WithAddress("192.168.0.1").
			WithServices("backend").
			Build()
		err := resManager.Create(context.Background(), dp,
			store.CreateBy(model.MetaToResourceKey(dp.GetMeta())),
			store.CreateWithLabels(map[string]string{
				metadata.KumaWorkload: "test-workload",
			}))
		Expect(err).ToNot(HaveOccurred())

		wl := workload_api.NewWorkloadResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).To(Succeed())
			g.Expect(wl.GetMeta().GetLabels()[mesh_proto.ManagedByLabel]).To(Equal("workload-generator"))
		}, "2s", "100ms").Should(Succeed())

		// Delete the dataplane
		Expect(resManager.Delete(context.Background(), dp, store.DeleteByKey("dp-1", model.DefaultMesh))).To(Succeed())

		// Wait until the Workload has been marked with grace period start
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).To(Succeed())
			g.Expect(wl.GetMeta().GetLabels()).To(HaveKey(mesh_proto.DeletionGracePeriodStartedLabel))
		}, "2s", "100ms").Should(Succeed())

		// Workload should not be deleted immediately (grace period is 500ms)
		Consistently(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).To(Succeed())
		}, "200ms", "50ms").Should(Succeed())

		// Workload should be deleted after grace period expires
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).ToNot(Succeed())
		}, "1s", "100ms").Should(Succeed())
	})

	It("should not delete Workload if a Dataplane comes back", func() {
		dp := builders.Dataplane().
			WithAddress("192.168.0.1").
			WithServices("backend").
			Build()
		err := resManager.Create(context.Background(), dp,
			store.CreateBy(model.MetaToResourceKey(dp.GetMeta())),
			store.CreateWithLabels(map[string]string{
				metadata.KumaWorkload: "test-workload",
			}))
		Expect(err).ToNot(HaveOccurred())

		wl := workload_api.NewWorkloadResource()

		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).To(Succeed())
			g.Expect(wl.GetMeta().GetLabels()[mesh_proto.ManagedByLabel]).To(Equal("workload-generator"))
		}, "2s", "100ms").Should(Succeed())

		// Delete the dataplane
		Expect(resManager.Delete(context.Background(), dp, store.DeleteByKey("dp-1", model.DefaultMesh))).To(Succeed())

		// Wait until the Workload has been marked with grace period start
		Eventually(func(g Gomega) {
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).To(Succeed())
			g.Expect(wl.GetMeta().GetLabels()).To(HaveKey(mesh_proto.DeletionGracePeriodStartedLabel))
		}, "2s", "100ms").Should(Succeed())

		// Recreate the dataplane
		dp2 := builders.Dataplane().
			WithAddress("192.168.0.1").
			WithServices("backend").
			Build()
		Expect(resManager.Create(context.Background(), dp2,
			store.CreateBy(model.MetaToResourceKey(dp2.GetMeta())),
			store.CreateWithLabels(map[string]string{
				metadata.KumaWorkload: "test-workload",
			}))).To(Succeed())

		// The Workload isn't ever deleted
		Consistently(func(g Gomega) {
			wl := workload_api.NewWorkloadResource()
			g.Expect(resManager.Get(context.Background(), wl, store.GetByKey("test-workload", model.DefaultMesh))).To(Succeed())
		}, "2s", "100ms").Should(Succeed())
	})

	It("should emit metric", func() {
		Eventually(func(g Gomega) {
			g.Expect(test_metrics.FindMetric(metrics, "component_workload_generator")).ToNot(BeNil())
		}, "2s", "100ms").Should(Succeed())
	})
})
