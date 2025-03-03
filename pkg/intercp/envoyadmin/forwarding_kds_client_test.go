package envoyadmin_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/intercp/catalog"
	"github.com/kumahq/kuma/pkg/intercp/envoyadmin"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/test/runtime"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Forwarding KDS Client", func() {
	const thisInstanceID = "instance-1"
	const otherInstanceID = "instance-2"
	t1, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")

	var forwardClient *countingForwardClient
	var adminClient *runtime.DummyEnvoyAdminClient
	var resManager manager.ResourceManager

	var forwardingClient admin.EnvoyAdminClient

	dp := samples.DataplaneBackendBuilder().
		WithName("east.backend").
		Build()

	BeforeEach(func() {
		forwardClient = &countingForwardClient{}
		adminClient = &runtime.DummyEnvoyAdminClient{}
		resManager = manager.NewResourceManager(memory.NewStore())
		cat := catalog.NewConfigCatalog(resManager)
		forwardingClient = envoyadmin.NewForwardingEnvoyAdminClient(
			resManager,
			cat,
			thisInstanceID,
			func(url string) (mesh_proto.InterCPEnvoyAdminForwardServiceClient, error) {
				return forwardClient, nil
			},
			adminClient,
		)

		_, err := cat.Replace(context.Background(), []catalog.Instance{
			{Id: thisInstanceID},
			{Id: otherInstanceID},
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(samples.MeshDefaultBuilder().Create(resManager)).To(Succeed())
		err = resManager.Create(context.Background(), dp, core_store.CreateByKey("east.backend", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	createZoneInsightConnectedToGlobal := func(insight string, globalInstanceID string, offline bool) {
		zoneInsight := system.NewZoneInsightResource()
		zoneInsight.Spec.EnvoyAdminStreams = &system_proto.EnvoyAdminStreams{
			ConfigDumpGlobalInstanceId: globalInstanceID,
			StatsGlobalInstanceId:      globalInstanceID,
			ClustersGlobalInstanceId:   globalInstanceID,
		}
		subscription := &system_proto.KDSSubscription{
			Id:               "1",
			GlobalInstanceId: globalInstanceID,
			ConnectTime:      util_proto.MustTimestampProto(t1),
		}
		if offline {
			subscription.DisconnectTime = util_proto.MustTimestampProto(t1.Add(1 * time.Hour))
		}
		zoneInsight.Spec.Subscriptions = append(zoneInsight.Spec.Subscriptions, subscription)
		err := resManager.Create(context.Background(), zoneInsight, core_store.CreateByKey(insight, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	}

	type testCase struct {
		globalInstanceID  string
		forwardedRequests int
		executedRequests  int
	}

	DescribeTable("when request for config dump is executed",
		func(given testCase) {
			// given
			createZoneInsightConnectedToGlobal("east", given.globalInstanceID, false)

			// when
			_, err := forwardingClient.ConfigDump(context.Background(), dp, false)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(forwardClient.xdsConfigCalled).To(Equal(given.forwardedRequests))
			Expect(adminClient.ConfigDumpCalled).To(Equal(given.executedRequests))
		},
		Entry("should forward request when the zone of dp is connected to other instance of Global CP", testCase{
			globalInstanceID:  otherInstanceID,
			forwardedRequests: 1,
			executedRequests:  0,
		}),
		Entry("should execute request when the zone of dp is connected to this instance of Global CP", testCase{
			globalInstanceID:  thisInstanceID,
			forwardedRequests: 0,
			executedRequests:  1,
		}),
	)

	DescribeTable("when request for stats is executed",
		func(given testCase) {
			// given
			createZoneInsightConnectedToGlobal("east", given.globalInstanceID, false)

			// when
			_, err := forwardingClient.Stats(context.Background(), dp, mesh_proto.AdminOutputFormat_TEXT)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(forwardClient.statsCalled).To(Equal(given.forwardedRequests))
			Expect(adminClient.StatsCalled).To(Equal(given.executedRequests))
		},
		Entry("should forward request when the zone of dp is connected to other instance of Global CP", testCase{
			globalInstanceID:  otherInstanceID,
			forwardedRequests: 1,
			executedRequests:  0,
		}),
		Entry("should execute request when the zone of dp is connected to this instance of Global CP", testCase{
			globalInstanceID:  thisInstanceID,
			forwardedRequests: 0,
			executedRequests:  1,
		}),
	)

	DescribeTable("when request for clusters is executed",
		func(given testCase) {
			// given
			createZoneInsightConnectedToGlobal("east", given.globalInstanceID, false)

			// when
			_, err := forwardingClient.Clusters(context.Background(), dp, mesh_proto.AdminOutputFormat_TEXT)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(forwardClient.clustersCalled).To(Equal(given.forwardedRequests))
			Expect(adminClient.ClustersCalled).To(Equal(given.executedRequests))
		},
		Entry("should forward request when the zone of dp is connected to other instance of Global CP", testCase{
			globalInstanceID:  otherInstanceID,
			forwardedRequests: 1,
			executedRequests:  0,
		}),
		Entry("should execute request when the zone of dp is connected to this instance of Global CP", testCase{
			globalInstanceID:  thisInstanceID,
			forwardedRequests: 0,
			executedRequests:  1,
		}),
	)

	It("should return an error when zone is offline", func() {
		// given
		createZoneInsightConnectedToGlobal("east", thisInstanceID, true)

		// when
		_, err := forwardingClient.Clusters(context.Background(), dp, mesh_proto.AdminOutputFormat_TEXT)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("couldn't execute Clusters operation, zone is offline"))
	})
})

type countingForwardClient struct {
	xdsConfigCalled int
	statsCalled     int
	clustersCalled  int
}

var _ mesh_proto.InterCPEnvoyAdminForwardServiceClient = &countingForwardClient{}

func (c *countingForwardClient) XDSConfig(ctx context.Context, in *mesh_proto.XDSConfigRequest, opts ...grpc.CallOption) (*mesh_proto.XDSConfigResponse, error) {
	c.xdsConfigCalled++
	return &mesh_proto.XDSConfigResponse{
		Result: &mesh_proto.XDSConfigResponse_Config{
			Config: []byte("forwarded"),
		},
	}, nil
}

func (c *countingForwardClient) Stats(ctx context.Context, in *mesh_proto.StatsRequest, opts ...grpc.CallOption) (*mesh_proto.StatsResponse, error) {
	c.statsCalled++
	return &mesh_proto.StatsResponse{
		Result: &mesh_proto.StatsResponse_Stats{
			Stats: []byte("forwarded"),
		},
	}, nil
}

func (c *countingForwardClient) Clusters(ctx context.Context, in *mesh_proto.ClustersRequest, opts ...grpc.CallOption) (*mesh_proto.ClustersResponse, error) {
	c.clustersCalled++
	return &mesh_proto.ClustersResponse{
		Result: &mesh_proto.ClustersResponse_Clusters{
			Clusters: []byte("forwarded"),
		},
	}, nil
}
