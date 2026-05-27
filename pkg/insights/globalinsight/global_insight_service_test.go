package globalinsight_test

import (
	"context"
	"encoding/json"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/insights/globalinsight"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
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
		err = createHostnameGenerator("default-hg", rs)
		Expect(err).ToNot(HaveOccurred())
		err = createHostnameGenerator("payments-hg", rs)
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

	It("should include mesh-scoped zone proxy dataplanes in zone stats", func() {
		// given
		globalInsightService := globalinsight.NewDefaultGlobalInsightService(rm)

		err := createMeshScopedZoneProxyDataplane("zi-online", "default", []mesh_proto.Dataplane_Networking_Listener_Type{
			mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
		}, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createDataplaneInsight("zi-online", "default", true, rs)
		Expect(err).ToNot(HaveOccurred())

		err = createMeshScopedZoneProxyDataplane("zi-ze-offline", "payments", []mesh_proto.Dataplane_Networking_Listener_Type{
			mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
			mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
		}, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createDataplaneInsight("zi-ze-offline", "payments", false, rs)
		Expect(err).ToNot(HaveOccurred())

		err = createMeshScopedZoneProxyDataplane("ze-no-insight", "payments", []mesh_proto.Dataplane_Networking_Listener_Type{
			mesh_proto.Dataplane_Networking_Listener_ZoneEgress,
		}, rs)
		Expect(err).ToNot(HaveOccurred())

		err = builders.Dataplane().
			WithName("standard-dp").
			WithMesh("default").
			WithServices("backend").
			Create(rs)
		Expect(err).ToNot(HaveOccurred())

		// when
		globalInsight, err := globalInsightService.GetGlobalInsight(context.Background())
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(globalInsight.Zones.ZoneIngresses.Total).To(Equal(2))
		Expect(globalInsight.Zones.ZoneIngresses.Online).To(Equal(1))
		Expect(globalInsight.Zones.ZoneEgresses.Total).To(Equal(2))
		Expect(globalInsight.Zones.ZoneEgresses.Online).To(Equal(0))
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

func createHostnameGenerator(name string, rs store.ResourceStore) error {
	return builders.HostnameGenerator().
		WithName(name).
		WithTemplate("{{ .Name }}.mesh").
		WithMeshServiceMatchLabels(map[string]string{
			"kuma.io/service": name,
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

func createMeshScopedZoneProxyDataplane(
	name string,
	mesh string,
	listenerTypes []mesh_proto.Dataplane_Networking_Listener_Type,
	rs store.ResourceStore,
) error {
	listeners := make([]*mesh_proto.Dataplane_Networking_Listener, 0, len(listenerTypes))
	for i, listenerType := range listenerTypes {
		listeners = append(listeners, &mesh_proto.Dataplane_Networking_Listener{
			Type:    listenerType,
			Address: "127.0.0.1",
			Port:    10001 + uint32(i),
		})
	}

	dataplane := builders.Dataplane().
		WithName(name).
		WithMesh(mesh).
		With(func(dp *core_mesh.DataplaneResource) {
			dp.Spec.Networking.Listeners = listeners
		}).
		Build()

	return rs.Create(
		context.Background(),
		dataplane,
		store.CreateByKey(name, mesh),
		store.CreateWithLabels(meshScopedZoneProxyLabels(listenerTypes)),
	)
}

func createDataplaneInsight(name string, mesh string, online bool, rs store.ResourceStore) error {
	builder := builders.DataplaneInsight().WithName(name).WithMesh(mesh)

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

func meshScopedZoneProxyLabels(
	listenerTypes []mesh_proto.Dataplane_Networking_Listener_Type,
) map[string]string {
	labels := map[string]string{}

	for _, listenerType := range listenerTypes {
		switch listenerType {
		case mesh_proto.Dataplane_Networking_Listener_ZoneIngress:
			labels[mesh_proto.ListenerZoneIngressLabel] = "enabled"
		case mesh_proto.Dataplane_Networking_Listener_ZoneEgress:
			labels[mesh_proto.ListenerZoneEgressLabel] = "enabled"
		}
	}

	return labels
}
