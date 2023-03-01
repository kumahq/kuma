package v1alpha1_test

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	plugin "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/plugin/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("Cluster modifications", func() {
	type testCase struct {
		clusters      []string
		modifications []string
		expected      string
	}

	DescribeTable("should apply modifications",
		func(given testCase) {
			// given
			set := core_xds.NewResourceSet()
			for _, clusterYAML := range given.clusters {
				cluster := &envoy_cluster.Cluster{}
				err := util_proto.FromYAML([]byte(clusterYAML), cluster)
				Expect(err).ToNot(HaveOccurred())
				set.Add(&core_xds.Resource{
					Name:     cluster.Name,
					Origin:   generator.OriginInbound,
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
			clusters: []string{
				`
                connectTimeout: 5s
                lbPolicy: CLUSTER_PROVIDED
                name: test:cluster
                type: ORIGINAL_DST`,
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
			clusters: []string{
				`
                connectTimeout: 5s
                lbPolicy: CLUSTER_PROVIDED
                name: test:cluster
                type: ORIGINAL_DST`,
			},
			modifications: []string{
				`
                cluster:
                   operation: Remove`,
			},
			expected: `{}`,
		}),
		Entry("should remove cluster matching name", testCase{
			clusters: []string{
				`
                connectTimeout: 5s
                lbPolicy: CLUSTER_PROVIDED
                name: test:cluster
                type: ORIGINAL_DST`,
				`
                connectTimeout: 5s
                lbPolicy: CLUSTER_PROVIDED
                name: test:cluster2
                type: ORIGINAL_DST`,
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
			clusters: []string{
				`
                connectTimeout: 5s
                lbPolicy: CLUSTER_PROVIDED
                name: test:cluster
                type: ORIGINAL_DST`,
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
			clusters: []string{
				`
                lbPolicy: CLUSTER_PROVIDED
                name: test:cluster
                outlierDetection:
                  enforcingConsecutive5xx: 100
                  enforcingConsecutiveGatewayFailure: 0
                  enforcingConsecutiveLocalOriginFailure: 0
                  enforcingFailurePercentage: 0
                  enforcingSuccessRate: 0
                type: ORIGINAL_DST`,
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
	)
})
