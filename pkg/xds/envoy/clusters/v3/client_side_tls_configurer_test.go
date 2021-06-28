package clusters_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

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
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.ClientSideTLS(given.endpoints)).
				Configure(clusters.Timeout(mesh_core.ProtocolTCP, &mesh_proto.Timeout_Conf{ConnectTimeout: durationpb.New(5 * time.Second)})).
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
            resourceApiVersion: V3
        name: testCluster
        transportSocketMatches:
        - match: {}
          name: httpbin.org
          transportSocket:
            name: envoy.transport_sockets.tls
            typedConfig:
              '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
              commonTlsContext: {}
              sni: httpbin.org:3000
        type: EDS
`}),
		Entry("cluster with mTLS and certs", testCase{
			clusterName: "testCluster",
			endpoints: []xds.Endpoint{
				{
					Target: "httpbin.org",
					Port:   3000,
					Tags:   nil,
					Weight: 100,
					ExternalService: &xds.ExternalService{
						TLSEnabled:         true,
						CaCert:             []byte("cacert"),
						ClientCert:         []byte("clientcert"),
						ClientKey:          []byte("clientkey"),
						AllowRenegotiation: true,
					},
				},
			},

			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: testCluster
            transportSocketMatches:
            - match: {}
              name: httpbin.org
              transportSocket:
                name: envoy.transport_sockets.tls
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
                  allowRenegotiation: true
                  commonTlsContext:
                    tlsCertificates:
                    - certificateChain:
                        inlineBytes: Y2xpZW50Y2VydA==
                      privateKey:
                        inlineBytes: Y2xpZW50a2V5
                    validationContext:
                      matchSubjectAltNames:
                      - exact: httpbin.org
                      trustedCa:
                        inlineBytes: Y2FjZXJ0
                  sni: httpbin.org:3000
            type: EDS
`}),
	)
})
