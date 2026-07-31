package postgres_test

import (
	"context"

	"github.com/gruntwork-io/terratest/modules/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_postgres "github.com/kumahq/kuma/v3/pkg/config/plugins/resources/postgres"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/postgres"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/postgres/config"
	test_metrics "github.com/kumahq/kuma/v3/pkg/test/metrics"
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
			for i := range 5 {
				ts := &meshexternalservice_api.MeshExternalServiceResource{
					Spec: &meshexternalservice_api.MeshExternalService{
						Match: meshexternalservice_api.Match{
							Type:     meshexternalservice_api.HostnameGeneratorType,
							Port:     8080,
							Protocol: core_meta.ProtocolHTTP,
						},
						Endpoints: &[]meshexternalservice_api.Endpoint{{
							Address: "192.168.0.1",
							Port:    8080,
						}},
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

			list := &meshexternalservice_api.MeshExternalServiceResourceList{}
			err = pStore.List(ctx, list, store.ListByResourceKeys(resourceKeys))
			Expect(err).ToNot(HaveOccurred())
			Expect(list.Items).To(HaveLen(5))

			// verify metric incremented
			metric := test_metrics.FindMetric(metrics, "store_postgres_list_query_threshold_exceeded_total")
			Expect(metric).ToNot(BeNil())
			Expect(metric.GetCounter().GetValue()).To(Equal(float64(1)))

			// list again to verify counter increments
			list = &meshexternalservice_api.MeshExternalServiceResourceList{}
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
			ts := &meshexternalservice_api.MeshExternalServiceResource{
				Spec: &meshexternalservice_api.MeshExternalService{
					Match: meshexternalservice_api.Match{
						Type:     meshexternalservice_api.HostnameGeneratorType,
						Port:     8080,
						Protocol: core_meta.ProtocolHTTP,
					},
					Endpoints: &[]meshexternalservice_api.Endpoint{{
						Address: "192.168.0.1",
						Port:    8080,
					}},
				},
			}
			err = pStore.Create(ctx, ts, store.CreateByKey("tr-a", "default"))
			Expect(err).ToNot(HaveOccurred())

			// list with resource keys count < threshold (1 < 3)
			resourceKeys := []core_model.ResourceKey{
				{Name: "tr-a", Mesh: "default"},
			}

			list := &meshexternalservice_api.MeshExternalServiceResourceList{}
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
			list := &meshexternalservice_api.MeshExternalServiceResourceList{}
			err = pStore.List(ctx, list)
			Expect(err).ToNot(HaveOccurred())

			// metric should be 0
			metric := test_metrics.FindMetric(metrics, "store_postgres_list_query_threshold_exceeded_total")
			Expect(metric).ToNot(BeNil())
			Expect(metric.GetCounter().GetValue()).To(Equal(float64(0)))
		})
	})
})
