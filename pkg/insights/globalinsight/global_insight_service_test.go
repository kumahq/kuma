package globalinsight_test

import (
	"context"
	"encoding/json"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/insights/globalinsight"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Global Insight", func() {
	var rm manager.ResourceManager
	var rs store.ResourceStore

	BeforeEach(func() {
		rs = memory.NewStore()
		rm = manager.NewResourceManager(rs)
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should compute global insight", func() {
		// given
		globalInsightService := globalinsight.NewDefaultGlobalInsightService(rm)

		err := createMeshInsight("default", rs)
		Expect(err).ToNot(HaveOccurred())
		err = createMeshInsight("payments", rs)
		Expect(err).ToNot(HaveOccurred())
		err = createServiceInsight("si-1", "default", rs)
		Expect(err).ToNot(HaveOccurred())
		err = createServiceInsight("si-2", "payments", rs)
		Expect(err).ToNot(HaveOccurred())
		err = createZoneInsight("zi-1", true, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createZoneInsight("zi-2", false, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createZoneIngressInsight("zii-1", "default", true, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createZoneIngressInsight("zii-2", "payments", false, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createZoneEgressInsight("zei-1", "default", true, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createZoneEgressInsight("zei-1", "payments", false, rs)
		Expect(err).ToNot(HaveOccurred())

		// when
		globalInsight, err := globalInsightService.GetGlobalInsight(context.Background())
		Expect(err).ToNot(HaveOccurred())

		// overwrite arbitrary CreatedAt so we can check equality
		globalInsight.CreatedAt = time.Time{}

		// then
		result, err := json.Marshal(globalInsight)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(matchers.MatchGoldenJSON(path.Join("testdata", "full_global_insight.golden.json")))
	})
})

func createMeshInsight(name string, rs store.ResourceStore) error {
	return builders.MeshInsight().
		WithName(name).
		WithStandardDataplaneStats(1, 1, 1, 3).
		WithBuiltinGatewayDataplaneStats(1, 0, 0, 1).
		WithDelegatedGatewayDataplaneStats(2, 1, 0, 3).
		AddPolicyStats("MeshTimeout", 2).
		AddPolicyStats("MeshRetry", 1).
		AddResourceStats("MeshTimeout", 2).
		AddResourceStats("MeshRetry", 1).
		Create(rs)
}

func createServiceInsight(name string, mesh string, rs store.ResourceStore) error {
	return builders.ServiceInsight().
		WithName(name).
		WithMesh(mesh).
		AddService("test-service", &mesh_proto.ServiceInsight_Service{
			ServiceType: mesh_proto.ServiceInsight_Service_internal,
			Status:      mesh_proto.ServiceInsight_Service_online,
		}).
		AddService("test-service-2", &mesh_proto.ServiceInsight_Service{
			ServiceType: mesh_proto.ServiceInsight_Service_internal,
			Status:      mesh_proto.ServiceInsight_Service_offline,
		}).
		AddService("test-external-service", &mesh_proto.ServiceInsight_Service{
			ServiceType: mesh_proto.ServiceInsight_Service_external,
		}).
		AddService("test-builtin-gateway", &mesh_proto.ServiceInsight_Service{
			ServiceType: mesh_proto.ServiceInsight_Service_gateway_builtin,
			Status:      mesh_proto.ServiceInsight_Service_partially_degraded,
		}).
		AddService("test-delegated-gateway", &mesh_proto.ServiceInsight_Service{
			ServiceType: mesh_proto.ServiceInsight_Service_gateway_delegated,
			Status:      mesh_proto.ServiceInsight_Service_offline,
		}).
		Create(rs)
}

func createZoneInsight(name string, online bool, rs store.ResourceStore) error {
	builder := builders.ZoneInsight().WithName(name)

	if online {
		builder.AddSubscription(&system_proto.KDSSubscription{
			ConnectTime: util_proto.MustTimestampProto(time.Unix(1694779925, 0)),
		})
	} else {
		builder.AddSubscription(&system_proto.KDSSubscription{
			ConnectTime:    util_proto.MustTimestampProto(time.Unix(1694779805, 0)),
			DisconnectTime: util_proto.MustTimestampProto(time.Unix(1694779925, 0)),
		})
	}

	return builder.Create(rs)
}

func createZoneIngressInsight(name string, mesh string, online bool, rs store.ResourceStore) error {
	builder := builders.ZoneIngressInsight().WithName(name).WithMesh(mesh)

	if online {
		builder.AddSubscription(&mesh_proto.DiscoverySubscription{
			ConnectTime: util_proto.MustTimestampProto(time.Unix(1694779805, 0)),
		})
	} else {
		builder.AddSubscription(&mesh_proto.DiscoverySubscription{
			ConnectTime:    util_proto.MustTimestampProto(time.Unix(1694779805, 0)),
			DisconnectTime: util_proto.MustTimestampProto(time.Unix(1694779925, 0)),
		})
	}

	return builder.Create(rs)
}

func createZoneEgressInsight(name string, mesh string, online bool, rs store.ResourceStore) error {
	builder := builders.ZoneEgressInsight().WithName(name).WithMesh(mesh)

	if online {
		builder.AddSubscription(&mesh_proto.DiscoverySubscription{
			ConnectTime: util_proto.MustTimestampProto(time.Unix(1694779805, 0)),
		})
	} else {
		builder.AddSubscription(&mesh_proto.DiscoverySubscription{
			ConnectTime:    util_proto.MustTimestampProto(time.Unix(1694779805, 0)),
			DisconnectTime: util_proto.MustTimestampProto(time.Unix(1694779925, 0)),
		})
	}

	return builder.Create(rs)
}
