package status

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	workload_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v2/pkg/util/proto"
)

var _ = Describe("Updater", func() {
	var stopCh chan struct{}
	var resManager manager.ResourceManager

	BeforeEach(func() {
		m, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		resManager = manager.NewResourceManager(memory.NewStore())

		updater, err := NewStatusUpdater(logr.Discard(), resManager, resManager, 50*time.Millisecond, m)
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

	It("should set empty stats when no dataplanes", func() {
		// given
		workload := &workload_api.WorkloadResource{
			Spec: &workload_api.Workload{},
		}
		Expect(resManager.Create(context.Background(), workload, store.CreateByKey("test-workload", model.DefaultMesh))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			w := workload_api.NewWorkloadResource()
			err := resManager.Get(context.Background(), w, store.GetByKey("test-workload", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(w.Status.DataplaneProxies).To(Equal(workload_api.DataplaneProxies{
				Connected: 0,
				Healthy:   0,
				Total:     0,
			}))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should count connected dataplanes", func() {
		// given
		workload := &workload_api.WorkloadResource{
			Spec: &workload_api.Workload{},
		}
		Expect(resManager.Create(context.Background(), workload, store.CreateByKey("backend", model.DefaultMesh))).To(Succeed())

		// when create dataplanes with workload label
		dpConnected := samples.DataplaneBackendBuilder().WithName("dp-connected").Build()
		Expect(resManager.Create(context.Background(), dpConnected, store.CreateByKey("dp-connected", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			metadata.KumaWorkload: "backend",
		}))).To(Succeed())

		dpDisconnected := samples.DataplaneBackendBuilder().WithName("dp-disconnected").Build()
		Expect(resManager.Create(context.Background(), dpDisconnected, store.CreateByKey("dp-disconnected", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			metadata.KumaWorkload: "backend",
		}))).To(Succeed())

		dpNeverConnected := samples.DataplaneBackendBuilder().WithName("dp-never-connected").Build()
		Expect(resManager.Create(context.Background(), dpNeverConnected, store.CreateByKey("dp-never-connected", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			metadata.KumaWorkload: "backend",
		}))).To(Succeed())

		// and insights
		insightConnected := samples.DataplaneInsightBackendBuilder().
			WithName("dp-connected").
			AddSubscription(&mesh_proto.DiscoverySubscription{
				ConnectTime: proto.MustTimestampProto(time.Now()),
			}).Build()
		Expect(resManager.Create(context.Background(), insightConnected, store.CreateByKey("dp-connected", model.DefaultMesh))).To(Succeed())

		insightDisconnected := samples.DataplaneInsightBackendBuilder().
			WithName("dp-disconnected").
			AddSubscription(&mesh_proto.DiscoverySubscription{
				ConnectTime:    proto.MustTimestampProto(time.Now()),
				DisconnectTime: proto.MustTimestampProto(time.Now()),
			}).Build()
		Expect(resManager.Create(context.Background(), insightDisconnected, store.CreateByKey("dp-disconnected", model.DefaultMesh))).To(Succeed())

		insightNeverConnected := samples.DataplaneInsightBackendBuilder().WithName("dp-never-connected").Build()
		Expect(resManager.Create(context.Background(), insightNeverConnected, store.CreateByKey("dp-never-connected", model.DefaultMesh))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			w := workload_api.NewWorkloadResource()
			err := resManager.Get(context.Background(), w, store.GetByKey("backend", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(w.Status.DataplaneProxies).To(Equal(workload_api.DataplaneProxies{
				Connected: 1,
				Healthy:   3, // all have healthy inbounds by default
				Total:     3,
			}))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should count healthy dataplanes", func() {
		// given
		workload := &workload_api.WorkloadResource{
			Spec: &workload_api.Workload{},
		}
		Expect(resManager.Create(context.Background(), workload, store.CreateByKey("backend", model.DefaultMesh))).To(Succeed())

		// when create dataplanes with different health states
		dpHealthy := samples.DataplaneBackendBuilder().WithName("dp-healthy").Build()
		Expect(resManager.Create(context.Background(), dpHealthy, store.CreateByKey("dp-healthy", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			metadata.KumaWorkload: "backend",
		}))).To(Succeed())

		dpUnhealthy := builders.Dataplane().
			WithName("dp-unhealthy").
			WithAddress("192.168.0.2").
			AddInboundOfTagsMap(map[string]string{"kuma.io/service": "backend"}).
			With(func(resource *core_mesh.DataplaneResource) {
				resource.Spec.Networking.Inbound[0].State = mesh_proto.Dataplane_Networking_Inbound_NotReady
			}).Build()
		Expect(resManager.Create(context.Background(), dpUnhealthy, store.CreateByKey("dp-unhealthy", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			metadata.KumaWorkload: "backend",
		}))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			w := workload_api.NewWorkloadResource()
			err := resManager.Get(context.Background(), w, store.GetByKey("backend", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(w.Status.DataplaneProxies).To(Equal(workload_api.DataplaneProxies{
				Connected: 0, // no insights created
				Healthy:   1, // only dp-healthy has ready inbound
				Total:     2,
			}))
		}, "10s", "100ms").Should(Succeed())
	})

	It("should not update status of workload from another zone", func() {
		// given workload from another zone
		workload := &workload_api.WorkloadResource{
			Spec: &workload_api.Workload{},
		}
		Expect(resManager.Create(context.Background(), workload, store.CreateByKey("backend", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			mesh_proto.ZoneTag:             "west",
			mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
		}))).To(Succeed())

		// when dataplane exists locally
		dp := samples.DataplaneBackendBuilder().Build()
		Expect(resManager.Create(context.Background(), dp, store.CreateByKey("dp-1", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			metadata.KumaWorkload: "backend",
		}))).To(Succeed())

		// then status should not be updated (stays empty)
		Consistently(func(g Gomega) {
			w := workload_api.NewWorkloadResource()
			err := resManager.Get(context.Background(), w, store.GetByKey("backend", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(w.Status.DataplaneProxies).To(Equal(workload_api.DataplaneProxies{
				Connected: 0,
				Healthy:   0,
				Total:     0,
			}))
		}, "1s", "100ms").Should(Succeed())
	})

	It("should only count dataplanes matching workload name", func() {
		// given two workloads
		workloadA := &workload_api.WorkloadResource{
			Spec: &workload_api.Workload{},
		}
		Expect(resManager.Create(context.Background(), workloadA, store.CreateByKey("workload-a", model.DefaultMesh))).To(Succeed())

		workloadB := &workload_api.WorkloadResource{
			Spec: &workload_api.Workload{},
		}
		Expect(resManager.Create(context.Background(), workloadB, store.CreateByKey("workload-b", model.DefaultMesh))).To(Succeed())

		// when create dataplanes with different workload labels
		dpA1 := samples.DataplaneBackendBuilder().WithName("dp-a-1").Build()
		Expect(resManager.Create(context.Background(), dpA1, store.CreateByKey("dp-a-1", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			metadata.KumaWorkload: "workload-a",
		}))).To(Succeed())

		dpA2 := samples.DataplaneBackendBuilder().WithName("dp-a-2").Build()
		Expect(resManager.Create(context.Background(), dpA2, store.CreateByKey("dp-a-2", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			metadata.KumaWorkload: "workload-a",
		}))).To(Succeed())

		dpB1 := samples.DataplaneBackendBuilder().WithName("dp-b-1").Build()
		Expect(resManager.Create(context.Background(), dpB1, store.CreateByKey("dp-b-1", model.DefaultMesh), store.CreateWithLabels(map[string]string{
			metadata.KumaWorkload: "workload-b",
		}))).To(Succeed())

		// then each workload should only count its own dataplanes
		Eventually(func(g Gomega) {
			wA := workload_api.NewWorkloadResource()
			err := resManager.Get(context.Background(), wA, store.GetByKey("workload-a", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(wA.Status.DataplaneProxies.Total).To(Equal(int32(2)))

			wB := workload_api.NewWorkloadResource()
			err = resManager.Get(context.Background(), wB, store.GetByKey("workload-b", model.DefaultMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(wB.Status.DataplaneProxies.Total).To(Equal(int32(1)))
		}, "10s", "100ms").Should(Succeed())
	})
})
