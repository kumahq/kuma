package gc_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/gc"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Collector", func() {
	Describe("Dataplane", func() {
		var rm manager.ResourceManager
		createDpAndDpInsight := func(name, mesh string, disconnectTime time.Time) {
			dp := samples.DataplaneBackendBuilder().
				WithName(name).
				WithMesh(mesh).
				Build()
			dpInsight := &core_mesh.DataplaneInsightResource{
				Meta: &model.ResourceMeta{Name: name, Mesh: mesh},
				Spec: &mesh_proto.DataplaneInsight{
					Subscriptions: []*mesh_proto.DiscoverySubscription{
						{
							DisconnectTime: proto.MustTimestampProto(disconnectTime),
						},
					},
				},
			}
			err := rm.Create(context.Background(), dp, store.CreateByKey(name, mesh))
			Expect(err).ToNot(HaveOccurred())
			err = rm.Create(context.Background(), dpInsight, store.CreateByKey(name, mesh))
			Expect(err).ToNot(HaveOccurred())
		}

		BeforeEach(func() {
			rm = manager.NewResourceManager(memory.NewStore())
			err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should cleanup old dataplanes", func() {
			now := time.Now()
			ticks := make(chan time.Time)
			defer close(ticks)
			// given 5 dataplanes now
			for i := 0; i < 5; i++ {
				createDpAndDpInsight(fmt.Sprintf("dp-%d", i), "default", now)
			}
			// given 5 dataplanes after an hour
			for i := 5; i < 10; i++ {
				createDpAndDpInsight(fmt.Sprintf("dp-%d", i), "default", now.Add(time.Hour))
			}

			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			collector, err := gc.NewCollector(rm, func() *time.Ticker {
				return &time.Ticker{C: ticks}
			}, 1*time.Hour, metrics, "dp", map[gc.InsightType]gc.ResourceType{
				gc.InsightType(core_mesh.DataplaneInsightType): gc.ResourceType(core_mesh.DataplaneType),
			})
			Expect(err).ToNot(HaveOccurred())

			stop := make(chan struct{})
			defer close(stop)
			go func() {
				_ = collector.Start(stop)
			}()

			// Run a first call to gc after 30 mins nothing happens (just disconnected)
			ticks <- now.Add(30 * time.Minute)
			Consistently(func(g Gomega) {
				dataplanes := &core_mesh.DataplaneResourceList{}
				g.Expect(rm.List(context.Background(), dataplanes)).To(Succeed())
				g.Expect(dataplanes.Items).To(HaveLen(10))
			}).Should(Succeed())

			// after 61 then first 5 dataplanes that are offline for more than 1 hour are deleted
			ticks <- now.Add(61 * time.Minute)
			Eventually(func(g Gomega) {
				dataplanes := &core_mesh.DataplaneResourceList{}
				g.Expect(rm.List(context.Background(), dataplanes)).To(Succeed())
				g.Expect(dataplanes.Items).To(HaveLen(5))
				g.Expect(dataplanes).To(WithTransform(func(actual *core_mesh.DataplaneResourceList) []string {
					var names []string
					for _, dp := range actual.Items {
						names = append(names, dp.Meta.GetName())
					}
					return names
				}, Equal([]string{"dp-5", "dp-6", "dp-7", "dp-8", "dp-9"})))
			}).Should(Succeed())
		})
	})

	Describe("ZoneResource", func() {
		var rm manager.ResourceManager
		createZoneIngressAndInsight := func(name string, disconnectTime time.Time) {
			zoneIngress := builders.ZoneIngress().
				WithAddress("1.1.1.1").
				WithPort(10001).
				WithName(name).
				Build()
			zoneIngressInsight := &core_mesh.ZoneIngressInsightResource{
				Meta: &model.ResourceMeta{Name: name, Mesh: core_model.NoMesh},
				Spec: &mesh_proto.ZoneIngressInsight{
					Subscriptions: []*mesh_proto.DiscoverySubscription{
						{
							DisconnectTime: proto.MustTimestampProto(disconnectTime),
						},
					},
				},
			}
			err := rm.Create(context.Background(), zoneIngress, store.CreateByKey(name, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			err = rm.Create(context.Background(), zoneIngressInsight, store.CreateByKey(name, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
		}

		createZoneEgressAndInsight := func(name string, disconnectTime time.Time) {
			zoneEgress := builders.ZoneEgress().
				WithAddress("1.1.1.1").
				WithPort(10002).
				WithName(name).
				Build()
			zoneEgressInsight := &core_mesh.ZoneEgressInsightResource{
				Meta: &model.ResourceMeta{Name: name, Mesh: core_model.NoMesh},
				Spec: &mesh_proto.ZoneEgressInsight{
					Subscriptions: []*mesh_proto.DiscoverySubscription{
						{
							DisconnectTime: proto.MustTimestampProto(disconnectTime),
						},
					},
				},
			}
			err := rm.Create(context.Background(), zoneEgress, store.CreateByKey(name, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			err = rm.Create(context.Background(), zoneEgressInsight, store.CreateByKey(name, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
		}

		BeforeEach(func() {
			rm = manager.NewResourceManager(memory.NewStore())
			err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should cleanup old zone ingresses and zone egresses", func() {
			now := time.Now()
			ticks := make(chan time.Time)
			defer close(ticks)
			// given 8 zone egresses now
			for i := 0; i < 8; i++ {
				createZoneEgressAndInsight(fmt.Sprintf("ze-%d", i), now)
			}
			// given 2 zone egresses after an hour
			for i := 8; i < 10; i++ {
				createZoneEgressAndInsight(fmt.Sprintf("ze-%d", i), now.Add(time.Hour))
			}
			// given 5 zone ingresses now
			for i := 0; i < 5; i++ {
				createZoneIngressAndInsight(fmt.Sprintf("zi-%d", i), now)
			}
			// given 5 zone ingresses after an hour
			for i := 5; i < 10; i++ {
				createZoneIngressAndInsight(fmt.Sprintf("zi-%d", i), now.Add(time.Hour))
			}

			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			collector, err := gc.NewCollector(rm, func() *time.Ticker {
				return &time.Ticker{C: ticks}
			}, 1*time.Hour, metrics, "zone", map[gc.InsightType]gc.ResourceType{
				gc.InsightType(core_mesh.ZoneIngressInsightType): gc.ResourceType(core_mesh.ZoneIngressType),
				gc.InsightType(core_mesh.ZoneEgressInsightType):  gc.ResourceType(core_mesh.ZoneEgressType),
			})
			Expect(err).ToNot(HaveOccurred())

			stop := make(chan struct{})
			defer close(stop)
			go func() {
				_ = collector.Start(stop)
			}()

			// Run a first call to gc after 30 mins nothing happens (just disconnected)
			ticks <- now.Add(30 * time.Minute)
			Consistently(func(g Gomega) {
				ze := &core_mesh.ZoneEgressResourceList{}
				g.Expect(rm.List(context.Background(), ze)).To(Succeed())
				g.Expect(ze.Items).To(HaveLen(10))
			}).Should(Succeed())

			Consistently(func(g Gomega) {
				zi := &core_mesh.ZoneIngressResourceList{}
				g.Expect(rm.List(context.Background(), zi)).To(Succeed())
				g.Expect(zi.Items).To(HaveLen(10))
			}).Should(Succeed())

			// after 61 then first 2 zone egresses that are offline for more than 1 hour are deleted
			ticks <- now.Add(61 * time.Minute)
			Eventually(func(g Gomega) {
				ze := &core_mesh.ZoneEgressResourceList{}
				g.Expect(rm.List(context.Background(), ze)).To(Succeed())
				g.Expect(ze.Items).To(HaveLen(2))
				g.Expect(ze).To(WithTransform(func(actual *core_mesh.ZoneEgressResourceList) []string {
					var names []string
					for _, ze := range actual.Items {
						names = append(names, ze.Meta.GetName())
					}
					return names
				}, Equal([]string{"ze-8", "ze-9"})))
			}).Should(Succeed())
			// after 61 then first 5 zone ingresses that are offline for more than 1 hour are deleted
			Eventually(func(g Gomega) {
				zi := &core_mesh.ZoneIngressResourceList{}
				g.Expect(rm.List(context.Background(), zi)).To(Succeed())
				g.Expect(zi.Items).To(HaveLen(5))
				g.Expect(zi).To(WithTransform(func(actual *core_mesh.ZoneIngressResourceList) []string {
					var names []string
					for _, zi := range actual.Items {
						names = append(names, zi.Meta.GetName())
					}
					return names
				}, Equal([]string{"zi-5", "zi-6", "zi-7", "zi-8", "zi-9"})))
			}).Should(Succeed())
		})
	})
})
