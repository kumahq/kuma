package clusters_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("CircuitBreakerConfigurer", func() {

	type testCase struct {
		clusterName    string
		circuitBreaker *mesh_core.CircuitBreakerResource
		expected       string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.CircuitBreaker(given.circuitBreaker)).
				Configure(clusters.Timeout(mesh_core.ProtocolTCP, &mesh_proto.Timeout_Conf{ConnectTimeout: durationpb.New(5 * time.Second)})).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("CircuitBreaker with thresholds", testCase{
			circuitBreaker: &mesh_core.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Thresholds: &mesh_proto.CircuitBreaker_Conf_Thresholds{
							MaxConnections:     &wrapperspb.UInt32Value{Value: 2},
							MaxPendingRequests: &wrapperspb.UInt32Value{Value: 3},
							MaxRequests:        &wrapperspb.UInt32Value{Value: 4},
							MaxRetries:         &wrapperspb.UInt32Value{Value: 5},
						},
					},
				},
			},
			expected: `
        circuitBreakers:
          thresholds:
          - maxConnections: 2
            maxPendingRequests: 3
            maxRequests: 4
            maxRetries: 5
        connectTimeout: 5s
        edsClusterConfig:
          edsConfig:
            ads: {}
            resourceApiVersion: V3
        type: EDS`,
		}),
	)
})
