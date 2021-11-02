package clusters_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("DNSClusterConfigurer", func() {

	type testCase struct {
		clusterName string
		address     string
		port        int32
		expected    string
		isHttps     bool
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
				Configure(clusters.DNSCluster(given.clusterName, given.address, uint32(given.port), given.isHttps)).
				Configure(clusters.Timeout(core_mesh.ProtocolTCP, DefaultTimeout())).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("should generate proper Envoy config", testCase{
			// given
			clusterName: "test:cluster",
			address:     "google.com",
			port:        80,
			isHttps:     false,
			expected: `
        altStatName: test_cluster
        connectTimeout: 5s
        loadAssignment:
          clusterName: test:cluster
          endpoints:
          - lbEndpoints:
            - endpoint:
                address:
                  socketAddress:
                    address: google.com
                    portValue: 80
        name: test:cluster
        type: STRICT_DNS`,
		}),
		Entry("should generate proper Envoy config for https", testCase{
			// given
			clusterName: "test:cluster",
			address:     "google.com",
			port:        80,
			isHttps:     true,
			expected: `
        altStatName: test_cluster
        connectTimeout: 5s
        loadAssignment:
          clusterName: test:cluster
          endpoints:
          - lbEndpoints:
            - endpoint:
                address:
                  socketAddress:
                    address: google.com
                    portValue: 80
        name: test:cluster
        type: STRICT_DNS
        transportSocket:
          name: envoy.transport_sockets.tls
          typedConfig:
            '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
            allowRenegotiation: true
            sni: google.com
        `,
		}),
	)
})
