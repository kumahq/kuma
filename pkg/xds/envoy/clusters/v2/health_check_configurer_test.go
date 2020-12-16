package clusters_test

import (
	"time"

	"github.com/golang/protobuf/ptypes"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("HealthCheckConfigurer", func() {

	type testCase struct {
		clusterName string
		healthCheck *mesh_core.HealthCheckResource
		expected    string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder(envoy.APIV2).
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.HealthCheck(given.healthCheck)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("HealthCheck with neither active nor passive checks", testCase{
			clusterName: "testCluster",
			healthCheck: mesh_core.NewHealthCheckResource(),
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
            name: testCluster
            type: EDS`,
		}),
		Entry("HealthCheck with active checks", testCase{
			clusterName: "testCluster",
			healthCheck: &mesh_core.HealthCheckResource{
				Spec: &mesh_proto.HealthCheck{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "redis"}},
					},
					Conf: &mesh_proto.HealthCheck_Conf{
						Interval:           ptypes.DurationProto(5 * time.Second),
						Timeout:            ptypes.DurationProto(4 * time.Second),
						UnhealthyThreshold: 3,
						HealthyThreshold:   2,
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
            healthChecks:
            - healthyThreshold: 2
              interval: 5s
              tcpHealthCheck: {}
              timeout: 4s
              unhealthyThreshold: 3
            name: testCluster
            type: EDS`,
		}),
	)
})
