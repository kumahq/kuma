package clusters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
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
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3, given.clusterName).
				Configure(clusters.EdsCluster()).
				Configure(clusters.ClientSideTLS(given.endpoints)).
				Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
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
              sni: httpbin.org
        type: EDS
`,
		}),
		Entry("cluster with mTLS and empty SNI because target is an IP address", testCase{
			clusterName: "testCluster",
			endpoints: []xds.Endpoint{
				{
					Target: "192.168.0.1",
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
          name: 192.168.0.1
          transportSocket:
            name: envoy.transport_sockets.tls
            typedConfig:
              '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
        type: EDS
`,
		}),
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
						ServerName:         "custom",
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
                      matchTypedSubjectAltNames:
                      - matcher:
                          exact: custom
                        sanType: DNS
                      - matcher:
                          exact: custom
                        sanType: IP_ADDRESS
                      trustedCa:
                        inlineBytes: Y2FjZXJ0
                  sni: custom
            type: EDS
`,
		}),
		Entry("cluster with mTLS and certs but skipHostnameVerification", testCase{
			clusterName: "testCluster",
			endpoints: []xds.Endpoint{
				{
					Target: "httpbin.org",
					Port:   3000,
					Tags:   nil,
					Weight: 100,
					ExternalService: &xds.ExternalService{
						TLSEnabled:               true,
						CaCert:                   []byte("cacert"),
						ClientCert:               []byte("clientcert"),
						ClientKey:                []byte("clientkey"),
						AllowRenegotiation:       true,
						SkipHostnameVerification: true,
						ServerName:               "",
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
                      trustedCa:
                        inlineBytes: Y2FjZXJ0
                  sni: httpbin.org
            type: EDS
`,
		}),
	)
})
