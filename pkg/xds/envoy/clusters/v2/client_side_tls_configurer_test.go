package clusters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("ClientSideTLSConfigurer", func() {

	type testCase struct {
		clusterName string
		endpoints   []xds.Endpoint
		expected    string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder(envoy.APIV2).
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.ClientSideTLS(given.endpoints)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("cluster with mTLS", testCase{
			clusterName: "testCluster",
			endpoints: []xds.Endpoint{
				{
					Target: "httpbin.org",
					Port:   3000,
					Tags:   nil,
					Weight: 100,
					ExternalService: &xds.ExternalService{
						TLSEnabled: true,
					},
				},
			},

			expected: `
        connectTimeout: 5s
        edsClusterConfig:
          edsConfig:
            ads: {}
        name: testCluster
        transportSocketMatches:
        - match: {}
          name: httpbin.org
          transportSocket:
            name: envoy.transport_sockets.tls
            typedConfig:
              '@type': type.googleapis.com/envoy.api.v2.auth.UpstreamTlsContext
              commonTlsContext: {}
        type: EDS
`}),
	)
})
