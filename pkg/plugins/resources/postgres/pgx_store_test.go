package postgres_test

import (
	"context"

	"github.com/gruntwork-io/terratest/modules/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_postgres "github.com/kumahq/kuma/v2/pkg/config/plugins/resources/postgres"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/postgres"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/postgres/config"
	test_metrics "github.com/kumahq/kuma/v2/pkg/test/metrics"
)

var _ = Describe("PgxStore", func() {
	Describe("list query threshold exceeded metric", func() {
		var pStore store.ResourceStore
		var metrics core_metrics.Metrics

		BeforeEach(func() {
			var err error
			dbCfg, err := c.Config()
			Expect(err).ToNot(HaveOccurred())
			dbCfg.MaxListQueryElements = 3
			dbCfg.MaxOpenConnections = 2
			dbCfg.DriverName = config_postgres.DriverNamePgx

			metrics, err = core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())

			_, err = postgres.MigrateDb(dbCfg)
			if err != nil {
				logger.Default.Logf(GinkgoT(), "error migrating database: %v", err)
				c.PrintDebugInfo(dbCfg.DbName, dbCfg.Port)
			}
			Expect(err).ToNot(HaveOccurred())

			pStore, err = postgres.NewPgxStore(metrics, dbCfg, config.NoopPgxConfigCustomizationFn)
			if err != nil {
				logger.Default.Logf(GinkgoT(), "error connecting to database: %v", err)
				c.PrintDebugInfo(dbCfg.DbName, dbCfg.Port)
			}
			Expect(err).ToNot(HaveOccurred())
		})

		It("should increment metric when resource keys exceed threshold", func() {
			ctx := context.Background()

			// create mesh
			err := pStore.Create(ctx, core_mesh.NewMeshResource(), store.CreateByKey("default", core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// create resources
			for i := 0; i < 5; i++ {
				ts := &core_mesh.TrafficRouteResource{
					Spec: &mesh_proto.TrafficRoute{
						Sources:      []*mesh_proto.Selector{{Match: map[string]string{"kuma.io/service": "web"}}},
						Destinations: []*mesh_proto.Selector{{Match: map[string]string{"kuma.io/service": "backend"}}},
						Conf:         &mesh_proto.TrafficRoute_Conf{Destination: map[string]string{"kuma.io/service": "backend"}},
					},
				}
				err := pStore.Create(ctx, ts, store.CreateByKey("tr-"+string(rune('a'+i)), "default"))
				Expect(err).ToNot(HaveOccurred())
			}

			// list with resource keys count >= threshold (5 >= 3)
			resourceKeys := []core_model.ResourceKey{
				{Name: "tr-a", Mesh: "default"},
				{Name: "tr-b", Mesh: "default"},
				{Name: "tr-c", Mesh: "default"},
				{Name: "tr-d", Mesh: "default"},
				{Name: "tr-e", Mesh: "default"},
			}

			list := &core_mesh.TrafficRouteResourceList{}
			err = pStore.List(ctx, list, store.ListByResourceKeys(resourceKeys))
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(5))

			// verify metric incremented
			metric := test_metrics.FindMetric(metrics, "store_postgres_list_query_threshold_exceeded_total")
			Expect(metric).ToNot(BeNil())
			Expect(metric.GetCounter().GetValue()).To(Equal(float64(1)))

			// list again to verify counter increments
			list = &core_mesh.TrafficRouteResourceList{}
			err = pStore.List(ctx, list, store.ListByResourceKeys(resourceKeys))
			Expect(err).ToNot(HaveOccurred())

			metric = test_metrics.FindMetric(metrics, "store_postgres_list_query_threshold_exceeded_total")
			Expect(metric.GetCounter().GetValue()).To(Equal(float64(2)))
		})

		It("should not increment metric when resource keys below threshold", func() {
			ctx := context.Background()

			// create mesh
			err := pStore.Create(ctx, core_mesh.NewMeshResource(), store.CreateByKey("default", core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// create resources
			ts := &core_mesh.TrafficRouteResource{
				Spec: &mesh_proto.TrafficRoute{
					Sources:      []*mesh_proto.Selector{{Match: map[string]string{"kuma.io/service": "web"}}},
					Destinations: []*mesh_proto.Selector{{Match: map[string]string{"kuma.io/service": "backend"}}},
					Conf:         &mesh_proto.TrafficRoute_Conf{Destination: map[string]string{"kuma.io/service": "backend"}},
				},
			}
			err = pStore.Create(ctx, ts, store.CreateByKey("tr-a", "default"))
			Expect(err).ToNot(HaveOccurred())

			// list with resource keys count < threshold (1 < 3)
			resourceKeys := []core_model.ResourceKey{
				{Name: "tr-a", Mesh: "default"},
			}

			list := &core_mesh.TrafficRouteResourceList{}
			err = pStore.List(ctx, list, store.ListByResourceKeys(resourceKeys))
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(1))

			// metric should be 0
			metric := test_metrics.FindMetric(metrics, "store_postgres_list_query_threshold_exceeded_total")
			Expect(metric).ToNot(BeNil())
			Expect(metric.GetCounter().GetValue()).To(Equal(float64(0)))
		})

		It("should not increment metric when no resource keys provided", func() {
			ctx := context.Background()

			// create mesh
			err := pStore.Create(ctx, core_mesh.NewMeshResource(), store.CreateByKey("default", core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// list without resource keys
			list := &core_mesh.TrafficRouteResourceList{}
			err = pStore.List(ctx, list)
			Expect(err).ToNot(HaveOccurred())

			// metric should be 0
			metric := test_metrics.FindMetric(metrics, "store_postgres_list_query_threshold_exceeded_total")
			Expect(metric).ToNot(BeNil())
			Expect(metric.GetCounter().GetValue()).To(Equal(float64(0)))
		})
	})
})
