package modifications_test

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/modifications"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Virtual Host modifications", func() {

	type testCase struct {
		routeCfgs     []string
		modifications []string
		expected      string
	}

	DescribeTable("should apply modifications",
		func(given testCase) {
			// given
			set := core_xds.NewResourceSet()
			for _, routeCfgYAML := range given.routeCfgs {
				routeCfg := &envoy_api.RouteConfiguration{}
				err := util_proto.FromYAML([]byte(routeCfgYAML), routeCfg)
				Expect(err).ToNot(HaveOccurred())
				set.Add(&core_xds.Resource{
					Name:     routeCfg.Name,
					Origin:   generator.OriginOutbound,
					Resource: routeCfg,
				})
			}

			var mods []*mesh_proto.ProxyTemplate_Modifications
			for _, modificationYAML := range given.modifications {
				modification := &mesh_proto.ProxyTemplate_Modifications{}
				err := util_proto.FromYAML([]byte(modificationYAML), modification)
				Expect(err).ToNot(HaveOccurred())
				mods = append(mods, modification)
			}

			// when
			err := modifications.Apply(set, mods)

			// then
			Expect(err).ToNot(HaveOccurred())
			resp, err := set.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("should add virtual host", testCase{
			routeCfgs: []string{
				`
                name: outbound:backend
                virtualHosts:
                - name: backend
                  domains:
                  - "*"
                  routes:
                  - match:
                      prefix: /
                    route:
                      cluster: backend
`,
			},
			modifications: []string{`
                virtualHost:
                   operation: add
                   value: |
                     name: backend
                     domains:
                     - backend.com
                     routes:
                     - match:
                         prefix: /
                       route:
                         cluster: backend`,
			},
			expected: `
            resources:
            - name: outbound:backend
              resource:
                '@type': type.googleapis.com/envoy.api.v2.RouteConfiguration
                name: outbound:backend
                virtualHosts:
                - domains:
                  - '*'
                  name: backend
                  routes:
                  - match:
                      prefix: /
                    route:
                      cluster: backend
                - domains:
                  - backend.com
                  name: backend
                  routes:
                  - match:
                      prefix: /
                    route:
                      cluster: backend`,
		}),
		Entry("should remove virtual host", testCase{
			routeCfgs: []string{
				`
                name: outbound:backend
                virtualHosts:
                - name: backend
                  domains:
                  - "*"
                  routes:
                  - match:
                      prefix: /backend
                    route:
                      cluster: backend
                virtualHosts:
                - name: web
                  domains:
                  - "*"
                  routes:
                  - match:
                      prefix: /web
                    route:
                      cluster: web
`,
			},
			modifications: []string{`
                virtualHost:
                   operation: remove
                   match:
                     name: backend`,
			},
			expected: `
            resources:
            - name: outbound:backend
              resource:
                '@type': type.googleapis.com/envoy.api.v2.RouteConfiguration
                name: outbound:backend
                virtualHosts:
                - domains:
                  - '*'
                  name: web
                  routes:
                  - match:
                      prefix: /web
                    route:
                      cluster: web`,
		}),
		Entry("should patch a virtual host", testCase{
			routeCfgs: []string{
				`
                name: outbound:backend
                virtualHosts:
                - name: backend
                  domains:
                  - "*"
                  routes:
                  - match:
                      prefix: /
                    route:
                      cluster: backend
`,
			},
			modifications: []string{`
                virtualHost:
                   operation: patch
                   match:
                     origin: outbound
                   value: |
                     retryPolicy:
                       retryOn: 5xx
                       numRetries: 3`,
			},
			expected: `
            resources:
            - name: outbound:backend
              resource:
                '@type': type.googleapis.com/envoy.api.v2.RouteConfiguration
                name: outbound:backend
                virtualHosts:
                - domains:
                  - '*'
                  name: backend
                  retryPolicy:
                    numRetries: 3
                    retryOn: 5xx
                  routes:
                  - match:
                      prefix: /
                    route:
                      cluster: backend`,
		}),
		Entry("should patch a virtual host adding new route", testCase{
			routeCfgs: []string{
				`
                name: outbound:backend
                virtualHosts:
                - name: backend
                  domains:
                  - "*"
                  routes:
                  - match:
                      prefix: /
                    route:
                      cluster: backend
`,
			},
			modifications: []string{`
                virtualHost:
                   operation: patch
                   match:
                     routeConfigurationName: outbound:backend
                   value: |
                     routes:
                     - match:
                         prefix: /web
                       route:
                         cluster: web`,
			},
			expected: `
            resources:
            - name: outbound:backend
              resource:
                '@type': type.googleapis.com/envoy.api.v2.RouteConfiguration
                name: outbound:backend
                virtualHosts:
                - domains:
                  - '*'
                  name: backend
                  routes:
                  - match:
                      prefix: /
                    route:
                      cluster: backend
                  - match:
                      prefix: /web
                    route:
                      cluster: web`,
		}),
	)
})
