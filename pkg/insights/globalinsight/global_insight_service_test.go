package globalinsight_test

import (
	"context"
	"encoding/json"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/insights/globalinsight"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
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
		err = createMeshService("svc-1-online", "default", 1, 1, 1, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createMeshService("svc-1-offline", "default", 0, 0, 1, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createMeshService("svc-2-online", "payments", 1, 1, 1, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createMeshService("svc-2-offline", "payments", 0, 0, 1, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createMeshService("svc-2-partial", "payments", 1, 1, 2, rs)
		Expect(err).ToNot(HaveOccurred())
		// connected to the CP but no inbound is ready, so the service is not serving traffic
		err = createMeshService("svc-2-unhealthy", "payments", 2, 0, 2, rs)
		Expect(err).ToNot(HaveOccurred())
		// all proxies connected, only some of them with ready inbounds
		err = createMeshService("svc-2-degraded", "payments", 2, 1, 2, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createMeshExternalService("es-1", "default", rs)
		Expect(err).ToNot(HaveOccurred())
		err = createMeshExternalService("es-2", "payments", rs)
		Expect(err).ToNot(HaveOccurred())
		err = createHostnameGenerator("default-hg", rs)
		Expect(err).ToNot(HaveOccurred())
		err = createHostnameGenerator("payments-hg", rs)
		Expect(err).ToNot(HaveOccurred())
		err = createZoneInsight("zi-1", true, rs)
		Expect(err).ToNot(HaveOccurred())
		err = createZoneInsight("zi-2", false, rs)
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
		WithDataplaneStats(3, 2, 1, 6).
		AddResourceStats("MeshTimeout", 2).
		AddResourceStats("MeshRetry", 1).
		Create(rs)
}

func createMeshService(name string, mesh string, connected, healthy, total int, rs store.ResourceStore) error {
	return builders.MeshService().
		WithName(name).
		WithMesh(mesh).
		WithDataplaneProxies(connected, healthy, total).
		Create(rs)
}

func createMeshExternalService(name string, mesh string, rs store.ResourceStore) error {
	return builders.MeshExternalService().
		WithName(name).
		WithMesh(mesh).
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
