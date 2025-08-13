package v1alpha1_test

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/origin"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/plugin/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
)

var _ = Describe("Cluster modifications", func() {
	type testCaseCluster struct {
		yaml   string
		origin origin.Origin
	}

	type testCase struct {
		clusters      []testCaseCluster
		modifications []string
		expected      string
	}

	DescribeTable("should apply modifications",
		func(given testCase) {
			// given
			set := core_xds.NewResourceSet()
			for _, c := range given.clusters {
				cluster := &envoy_cluster.Cluster{}
				err := util_proto.FromYAML([]byte(c.yaml), cluster)
				Expect(err).ToNot(HaveOccurred())
				origin := c.origin
				if origin == "" {
					origin = metadata.OriginInbound
				}
				set.Add(&core_xds.Resource{
					Name:     cluster.Name,
					Origin:   origin,
					Resource: cluster,
				})
			}

			var mods []api.Modification
			for _, modificationYAML := range given.modifications {
				modification := api.Modification{}
				err := yaml.Unmarshal([]byte(modificationYAML), &modification)
				Expect(err).ToNot(HaveOccurred())
				mods = append(mods, modification)
			}

			// when
			err := plugin.ApplyMods(set, mods)

			// then
			Expect(err).ToNot(HaveOccurred())
			resp, err := set.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("should add cluster", testCase{
			modifications: []string{
				`
                cluster:
                   operation: Add
                   value: |
                     edsClusterConfig:
                       edsConfig:
                         ads: {}
                     name: test:cluster
                     type: EDS`,
			},
			expected: `
            resources:
            - name: test:cluster
              resource:
                '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
                edsClusterConfig:
                  edsConfig:
                    ads: {}
                name: test:cluster
                type: EDS`,
		}),
		Entry("should replace cluster", testCase{
			clusters: []testCaseCluster{
				{
					yaml: `
                    connectTimeout: 5s
                    lbPolicy: CLUSTER_PROVIDED
                    name: test:cluster
                    type: ORIGINAL_DST
                    `,
				},
			},
			modifications: []string{
				`
                cluster:
                   operation: Add
                   value: |
                     edsClusterConfig:
                       edsConfig:
                         ads: {}
                     name: test:cluster
                     type: EDS`,
			},
			expected: `
            resources:
            - name: test:cluster
              resource:
                '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
                edsClusterConfig:
                  edsConfig:
                    ads: {}
                name: test:cluster
                type: EDS`,
		}),
		Entry("should remove cluster matching all", testCase{
			clusters: []testCaseCluster{
				{
					yaml: `
                    connectTimeout: 5s
                    lbPolicy: CLUSTER_PROVIDED
                    name: test:cluster
                    type: ORIGINAL_DST
`,
				},
			},
			modifications: []string{
				`
                cluster:
                   operation: Remove`,
			},
			expected: `{}`,
		}),
		Entry("should remove cluster matching name", testCase{
			clusters: []testCaseCluster{
				{
					yaml: `
                    connectTimeout: 5s
                    lbPolicy: CLUSTER_PROVIDED
                    name: test:cluster
                    type: ORIGINAL_DST
`,
				},
				{
					yaml: `
                    connectTimeout: 5s
                    lbPolicy: CLUSTER_PROVIDED
                    name: test:cluster2
                    type: ORIGINAL_DST
`,
				},
			},
			modifications: []string{
				`
                cluster:
                   operation: Remove
                   match:
                     name: test:cluster`,
			},
			expected: `
            resources:
            - name: test:cluster2
              resource:
                '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
                connectTimeout: 5s
                lbPolicy: CLUSTER_PROVIDED
                name: test:cluster2
                type: ORIGINAL_DST`,
		}),
		Entry("should remove all inbound clusters", testCase{
			clusters: []testCaseCluster{
				{
					yaml: `
                    connectTimeout: 5s
                    lbPolicy: CLUSTER_PROVIDED
                    name: test:cluster
                    type: ORIGINAL_DST
`,
				},
			},
			modifications: []string{
				`
                cluster:
                   operation: Remove
                   match:
                     origin: inbound`,
			},
			expected: `{}`,
		}),
		Entry("should patch cluster matching name", testCase{
			clusters: []testCaseCluster{
				{
					yaml: `
                    lbPolicy: CLUSTER_PROVIDED
                    name: test:cluster
                    outlierDetection:
                      enforcingConsecutive5xx: 100
                      enforcingConsecutiveGatewayFailure: 0
                      enforcingConsecutiveLocalOriginFailure: 0
                      enforcingFailurePercentage: 0
                      enforcingSuccessRate: 0
                    type: ORIGINAL_DST
`,
				},
			},
			modifications: []string{
				`
                cluster:
                   operation: Patch
                   match:
                     name: test:cluster
                   value: |
                     connectTimeout: 5s
                     httpProtocolOptions:
                       acceptHttp10: true
                     outlierDetection:
                       enforcingSuccessRate: 100`,
			},
			expected: `
            resources:
            - name: test:cluster
              resource:
                '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
                connectTimeout: 5s
                httpProtocolOptions:
                  acceptHttp10: true
                lbPolicy: CLUSTER_PROVIDED
                name: test:cluster
                outlierDetection:
                  enforcingConsecutive5xx: 100
                  enforcingConsecutiveGatewayFailure: 0
                  enforcingConsecutiveLocalOriginFailure: 0
                  enforcingFailurePercentage: 0
                  enforcingSuccessRate: 100
                type: ORIGINAL_DST`,
		}),
		Entry("should patch cluster matching origins with JsonPatch", testCase{
			clusters: []testCaseCluster{
				{
					yaml: `
                    name: foo-service
                    type: EDS
                    connectTimeout: 5s
                    transportSocket:
                      name: envoy.transport_sockets.tls
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
                        commonTlsContext:
                          alpnProtocols:
                            - kuma
                        sni: foo-service{mesh=default}
                    `,
					origin: metadata.OriginOutbound,
				},
				{
					yaml: `
                    name: bar-service
                    type: ORIGINAL_DST
                    connectTimeout: 15s
                    transportSocket:
                      name: envoy.transport_sockets.tls
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
                        sni: bar-service{mesh=default}
                    `,
					origin: metadata.OriginInbound,
				},
				{
					yaml: `
                    name: baz-service
                    type: ORIGINAL_DST
                    connectTimeout: 11s
                    transportSocket:
                      name: envoy.transport_sockets.tls
                      typedConfig:
                        '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
                        sni: baz-service{mesh=default}
                    `,
					origin: metadata.OriginOutbound,
				},
			},
			modifications: []string{
				`
                cluster:
                  operation: Patch
                  match:
                    origin: outbound
                  jsonPatches:
                  - op: add
                    path: /transportSocket/typedConfig/commonTlsContext/tlsParams
                    value: { "tlsMinimumProtocolVersion": "TLSv1_2" }
                  - op: add
                    path: /transportSocket/typedConfig/commonTlsContext/tlsParams/tlsMaximumProtocolVersion
                    value: TLSv1_2
                  - op: replace
                    path: /connectTimeout
                    value: 77s
                `,
			},
			expected: `
            resources:
            - name: bar-service
              resource:
                '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
                name: bar-service
                type: ORIGINAL_DST
                connectTimeout: 15s
                transportSocket:
                  name: envoy.transport_sockets.tls
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
                    sni: bar-service{mesh=default}
            - name: baz-service
              resource:
                '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
                name: baz-service
                type: ORIGINAL_DST
                connectTimeout: 77s
                transportSocket:
                  name: envoy.transport_sockets.tls
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
                    commonTlsContext:
                      tlsParams:
                        tlsMinimumProtocolVersion: TLSv1_2
                        tlsMaximumProtocolVersion: TLSv1_2
                    sni: baz-service{mesh=default}
            - name: foo-service
              resource:
                '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
                name: foo-service
                type: EDS
                connectTimeout: 77s
                transportSocket:
                  name: envoy.transport_sockets.tls
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
                    commonTlsContext:
                      alpnProtocols:
                        - kuma
                      tlsParams:
                        tlsMinimumProtocolVersion: TLSv1_2
                        tlsMaximumProtocolVersion: TLSv1_2
                    sni: foo-service{mesh=default}
            `,
		}),
	)
})
