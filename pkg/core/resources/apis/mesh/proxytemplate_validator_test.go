package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("ProxyTemplate", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(spec string) {
				proxyTemplate := mesh.ProxyTemplateResource{}

				// when
				err := util_proto.FromYAML([]byte(spec), &proxyTemplate.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				err = proxyTemplate.Validate()
				// then
				Expect(err).ToNot(HaveOccurred())
			},
			Entry("full example", `
                selectors:
                - match:
                    service: backend
                conf:
                  imports:
                  - default-proxy
                  resources:
                  - name: additional
                    version: v1
                    resource: | 
                      '@type': type.googleapis.com/envoy.api.v2.Cluster
                      connectTimeout: 5s
                      loadAssignment:
                        clusterName: localhost:8443
                        endpoints:
                          - lbEndpoints:
                              - endpoint:
                                  address:
                                    socketAddress:
                                      address: 127.0.0.1
                                      portValue: 8443
                      name: localhost:8443
                      type: STATIC`,
			),
			Entry("empty conf", `
                selectors:
                - match:
                    service: backend`,
			),
		)

		type testCase struct {
			proxyTemplate string
			expected      string
		}
		DescribeTable("should validate fields",
			func(given testCase) {
				// given
				proxyTemplate := mesh.ProxyTemplateResource{}

				// when
				err := util_proto.FromYAML([]byte(given.proxyTemplate), &proxyTemplate.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := proxyTemplate.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty import", testCase{
				proxyTemplate: `
                conf:
                  imports:
                  - ""
                selectors:
                - match:
                    service: backend`,
				expected: `
                violations:
                - field: conf.imports[0]
                  message: cannot be empty`,
			}),
			Entry("unknown profile", testCase{
				proxyTemplate: `
                conf:
                  imports:
                  - unknown-profile
                selectors:
                - match:
                    service: backend`,
				expected: `
                violations:
                - field: conf.imports[0]
                  message: 'profile not found. Available profiles: default-proxy'`,
			}),
			Entry("resources empty fields", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    service: backend
                conf:
                  resources:
                  - name:
                    version:
                    resource:`,
				expected: `
                violations:
                - field: conf.resources[0].name
                  message: cannot be empty
                - field: conf.resources[0].version
                  message: cannot be empty
                - field: conf.resources[0].resource
                  message: cannot be empty`,
			}),
			Entry("selector without tags", testCase{
				proxyTemplate: `
                selectors:
                - match:`,
				expected: `
                violations:
                - field: selectors[0].match
                  message: must have at least one tag
                - field: selectors[0].match
                  message: mandatory tag "service" is missing`,
			}),
			Entry("empty tag", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    "": asdf`,
				expected: `
                violations:
                - field: selectors[0].match
                  message: tag name must be non-empty
                - field: selectors[0].match
                  message: mandatory tag "service" is missing`,
			}),
			Entry("empty tag value", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    service:`,
				expected: `
                violations:
                - field: 'selectors[0].match["service"]'
                  message: tag value must be non-empty`,
			}),
			Entry("validation error from envoy protobuf resource", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    service: backend
                conf:
                  resources:
                  - name: additional
                    version: v1
                    resource: | 
                      '@type': type.googleapis.com/envoy.api.v2.Cluster
                      loadAssignment:
                        clusterName: localhost:8443
                        endpoints:
                          - lbEndpoints:
                              - endpoint:
                                  address:
                                    socketAddress:
                                      address: 127.0.0.1
                                      portValue: 8443`,
				expected: `
                violations:
                - field: conf.resources[0].resource
                  message: 'native Envoy resource is not valid: invalid Cluster.Name: value length must be at least 1 bytes'`,
			}),
			Entry("invalid envoy resource", testCase{
				proxyTemplate: `
                selectors:
                - match:
                    service: backend
                conf:
                  resources:
                  - name: additional
                    version: v1
                    resource: not-envoy-resource`,
				expected: `
                violations:
                - field: conf.resources[0].resource
                  message: 'native Envoy resource is not valid: json: cannot unmarshal string into Go value of type map[string]*json.RawMessage'`,
			}),
		)
	})
})
