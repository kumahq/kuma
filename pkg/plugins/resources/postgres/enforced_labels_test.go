package postgres_test

import (
	"context"
	"maps"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	config_postgres "github.com/kumahq/kuma/v3/pkg/config/plugins/resources/postgres"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/postgres"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/postgres/config"
)

var _ = Describe("PgxStore enforced read labels", func() {
	newStore := func(cp resource_labels.ControlPlane) store.ResourceStore {
		dbCfg, err := c.Config()
		Expect(err).ToNot(HaveOccurred())
		dbCfg.MaxOpenConnections = 2
		dbCfg.DriverName = config_postgres.DriverNamePgx

		metrics, err := core_metrics.NewMetrics("Zone")
		Expect(err).ToNot(HaveOccurred())

		_, err = postgres.MigrateDb(dbCfg)
		Expect(err).ToNot(HaveOccurred())

		pStore, err := postgres.NewPgxStore(metrics, dbCfg, config.NoopPgxConfigCustomizationFn, cp)
		Expect(err).ToNot(HaveOccurred())
		return pStore
	}

	newResource := func() *meshexternalservice_api.MeshExternalServiceResource {
		return &meshexternalservice_api.MeshExternalServiceResource{
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
	}

	labelsOf := func(s store.ResourceStore, name string) (map[string]string, map[string]string) {
		ctx := context.Background()

		got := meshexternalservice_api.NewMeshExternalServiceResource()
		Expect(s.Get(ctx, got, store.GetByKey(name, "default"))).To(Succeed())

		list := &meshexternalservice_api.MeshExternalServiceResourceList{}
		Expect(s.List(ctx, list, store.ListByMesh("default"))).To(Succeed())
		var listed map[string]string
		for _, item := range list.Items {
			if item.GetMeta().GetName() == name {
				listed = item.GetMeta().GetLabels()
			}
		}
		Expect(listed).NotTo(BeNil())
		return got.GetMeta().GetLabels(), listed
	}

	DescribeTable("on a zone should hand out origin and zone recomputed on Get and List",
		func(name string, stored map[string]string, expected map[string]string) {
			s := newStore(resource_labels.ControlPlane{Mode: config_core.Zone, Zone: "zone-1"})
			input := maps.Clone(stored)

			Expect(s.Create(context.Background(), newResource(), store.CreateByKey(name, "default"), store.CreateWithLabels(input))).To(Succeed())

			got, listed := labelsOf(s, name)
			Expect(got).To(Equal(expected))
			Expect(listed).To(Equal(expected))
			Expect(input).To(Equal(stored))
		},
		Entry("absent labels are filled in", "mes-absent", nil, map[string]string{
			mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			mesh_proto.ZoneTag:             "zone-1",
		}),
		Entry("a local resource gets the local zone", "mes-local", map[string]string{
			mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			mesh_proto.ZoneTag:             "other-zone",
		}, map[string]string{
			mesh_proto.ResourceOriginLabel: string(mesh_proto.ZoneResourceOrigin),
			mesh_proto.ZoneTag:             "zone-1",
		}),
		Entry("an import from global is kept as stored", "mes-import", map[string]string{
			mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
			mesh_proto.ZoneTag:             "other-zone",
		}, map[string]string{
			mesh_proto.ResourceOriginLabel: string(mesh_proto.GlobalResourceOrigin),
			mesh_proto.ZoneTag:             "other-zone",
		}),
	)

	It("should hand out the stored labels as-is without a mode", func() {
		s := newStore(resource_labels.ControlPlane{})

		Expect(s.Create(context.Background(), newResource(), store.CreateByKey("mes-no-mode", "default"), store.CreateWithLabels(map[string]string{"app": "backend"}))).To(Succeed())

		got, listed := labelsOf(s, "mes-no-mode")
		Expect(got).To(Equal(map[string]string{"app": "backend"}))
		Expect(listed).To(Equal(map[string]string{"app": "backend"}))
	})
})
