package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("TrafficRoute", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(yaml string) {
				// given
				resource := NewTrafficRouteResource()
				err := util_proto.FromYAML([]byte(yaml), resource.Spec)
				Expect(err).ToNot(HaveOccurred())

				// when
				err = resource.Validate()

				// then
				Expect(err).ToNot(HaveOccurred())
			},
			Entry("example with split", `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  loadBalancer:
                    roundRobin: {}
                  split:
                  - weight: 100
                    destination:
                      kuma.io/service: offers
                  - weight: 0
                    destination:
                      kuma.io/service: backend`,
			),
			Entry("example with destination", `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  loadBalancer:
                    leastRequest: {}
                  destination:
                    kuma.io/service: offers`,
			),
			Entry("example with http", `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  loadBalancer:
                    ringHash:
                      hashFunction: XX_HASH
                  http:
                  - match:
                      path:
                        prefix: "/offers"
                      method:
                        exact: "GET"
                      headers:
                        x-custom-header:
                          regex: "^xyz$"
                    modify:
                      path:
                        regex:
                          pattern: "^/service/([^/]+)(/.*)$"
                          substitution: "\\2/instance/\\1"
                        host:
                          fromPath:
                            pattern: "^/(.+)/.+$"
                            substitution: "\\1"
                    destination:
                      kuma.io/service: offers
                  destination:
                    kuma.io/service: backend`,
			),
			Entry("example with http split", `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  http:
                  - match:
                      path:
                        prefix: "/offers"
                    modify:
                      path:
                        rewritePrefix: "/not-offers"
                      host:
                        value: xyz
                      requestHeaders:
                        add:
                          - name: x-custom-header
                            value: xyz
                            append: true 
                        remove:
                          - name: x-something
                      responseHeaders:
                        add:
                          - name: x-custom-header
                            value: xyz
                            append: true
                        remove:
                          - name: x-something
                    split:
                      - weight: 20
                        destination:
                          kuma.io/service: offers
                  split:
                    - weight: 100
                      destination:
                        kuma.io/service: backend`,
			),
		)

		type testCase struct {
			route    string
			expected string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				route := NewTrafficRouteResource()

				// when
				err := util_proto.FromYAML([]byte(given.route), route.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := route.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty spec", testCase{
				route: ``,
				expected: `
                violations:
                - field: sources
                  message: must have at least one element
                - field: destinations
                  message: must have at least one element
                - field: conf
                  message: cannot be empty
`,
			}),
			Entry("selectors without tags", testCase{
				route: `
                sources:
                - match: {}
                destinations:
                - match: {}
                conf:
                  split:
                  - destination: {}
`,
				expected: `
                violations:
                - field: sources[0].match
                  message: must have at least one tag
                - field: sources[0].match
                  message: mandatory tag "kuma.io/service" is missing
                - field: destinations[0].match
                  message: must consist of exactly one tag "kuma.io/service"
                - field: destinations[0].match
                  message: mandatory tag "kuma.io/service" is missing
                - field: conf.split[0].weight
                  message: needs to be defined
                - field: conf.split[0].destination
                  message: must have at least one tag
                - field: conf.split[0].destination
                  message: mandatory tag "kuma.io/service" is missing
                - field: conf.split
                  message: there must be at least one split entry with weight above 0
`,
			}),
			Entry("selectors with empty tags values", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service:
                    region:
                destinations:
                - match:
                    kuma.io/service:
                    region:
                conf:
                  split:
                  - weight: 1
                    destination:
                      kuma.io/service:
                      region:
`,
				expected: `
                violations:
                - field: sources[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: sources[0].match["region"]
                  message: tag value must be non-empty
                - field: destinations[0].match
                  message: must consist of exactly one tag "kuma.io/service"
                - field: destinations[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: destinations[0].match["region"]
                  message: tag "region" is not allowed
                - field: destinations[0].match["region"]
                  message: tag value must be non-empty
                - field: conf.split[0].destination["kuma.io/service"]
                  message: tag value must be non-empty
                - field: conf.split[0].destination["region"]
                  message: tag value must be non-empty
`,
			}),
			Entry("multiple selectors", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service:
                    region:
                - match: {}
                destinations:
                - match:
                    kuma.io/service:
                    region:
                - match: {}
                conf:
                  destination:
                    region:
`,
				expected: `
                violations:
                - field: sources[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: sources[0].match["region"]
                  message: tag value must be non-empty
                - field: sources[1].match
                  message: must have at least one tag
                - field: sources[1].match
                  message: mandatory tag "kuma.io/service" is missing
                - field: destinations[0].match
                  message: must consist of exactly one tag "kuma.io/service"
                - field: destinations[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: destinations[0].match["region"]
                  message: tag "region" is not allowed
                - field: destinations[0].match["region"]
                  message: tag value must be non-empty
                - field: destinations[1].match
                  message: must consist of exactly one tag "kuma.io/service"
                - field: destinations[1].match
                  message: mandatory tag "kuma.io/service" is missing
                - field: conf.destination["region"]
                  message: tag value must be non-empty
                - field: conf.destination
                  message: mandatory tag "kuma.io/service" is missing
`,
			}),
			Entry("wrong ring hash function in the load balancer", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: '*'
                destinations:
                - match:
                    kuma.io/service: '*'
                conf:
                  split:
                  - weight: 100
                    destination:
                      kuma.io/service: 'backend'
                  loadBalancer:
                    ringHash:
                      hashFunction: 'INVALID_HASH_FUNCTION'
`,
				expected: `
                violations:
                - field: conf.loadBalancer.ringHash.hashFunction
                  message: must have a valid hash function
`,
			}),
			Entry("cannot define both split and destination", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  split:
                  - weight: 100
                    destination:
                      kuma.io/service: offers
                  destination:
                    kuma.io/service: offers
`,
				expected: `
                violations:
                - field: conf
                  message: '"split" cannot be defined at the same time as "destination"'
`,
			}),
			Entry("requires either split or a destination", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf: {}
`,
				expected: `
                violations:
                - field: conf
                  message: 'requires either "destination" or "split"'
`,
			}),
			Entry("http - cannot define both split and destination", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  http:
                  - match:
                      path:
                        prefix: "/offers"
                    split:
                    - weight: 100
                      destination:
                        kuma.io/service: offers
                    destination:
                      kuma.io/service: offers
                  destination:
                    kuma.io/service: offers
`,
				expected: `
                violations:
                - field: conf.http[0]
                  message: '"split" cannot be defined at the same time as "destination"'
`,
			}),
			Entry("http - requires either split or destination", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  http:
                  - match:
                      path:
                        prefix: "/offers"
                  destination:
                    kuma.io/service: offers
`,
				expected: `
                violations:
                - field: conf.http[0]
                  message: 'requires either "destination" or "split"'
`,
			}),
			Entry("http - match must be present", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  http:
                  - destination:
                      kuma.io/service: offers
                  destination:
                    kuma.io/service: offers
`,
				expected: `
                violations:
                - field: conf.http[0].match
                  message: 'must be present and contain at least one of the elements: "method", "path" or "headers"'
`,
			}),
			Entry("http - invalid match values", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  http:
                  - match:
                      method: {}
                      path: {}
                      headers: {}
                    destination:
                      kuma.io/service: offers
                  - match:
                      method:
                        regex: ""
                      path:
                        prefix: ""
                      headers:
                        "": {}
                    destination:
                      kuma.io/service: offers
                  destination:
                    kuma.io/service: offers
`,
				expected: `
                violations:
                - field: conf.http[0].match.method
                  message: 'cannot be empty. Available options: "exact", "split" or "regex"'
                - field: conf.http[0].match.path
                  message: 'cannot be empty. Available options: "exact", "split" or "regex"'
                - field: conf.http[0].match.headers
                  message: must contain at least one element
                - field: conf.http[1].match.method.regex
                  message: cannot be empty
                - field: conf.http[1].match.path.prefix
                  message: cannot be empty
                - field: conf.http[1].match.headers[""]
                  message: cannot be empty
                - field: conf.http[1].match.headers[""]
                  message: 'cannot be empty. Available options: "exact", "split" or "regex"'
`,
			}),
			Entry("split with all entries that sums to 0", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  split:
                  - weight: 0
                    destination:
                      kuma.io/service: offers`,
				expected: `
                violations:
                - field: conf.split
                  message: there must be at least one split entry with weight above 0`,
			}),
			Entry("http - modify empty values", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  http:
                  - match:
                      path:
                        prefix: "/offers"
                    modify:
                      path:
                        rewritePrefix: ""
                      host:
                        value: ""
                      requestHeaders:
                        add:
                          - name: ""
                            value: ""
                        remove:
                          - name: ""
                      responseHeaders:
                        add:
                          - name: ""
                            value: ""
                        remove:
                          - name: ""
                    destination:
                      kuma.io/service: offers
                  - match:
                      path:
                        prefix: "/offers1"
                    modify:
                      path:
                        regex:
                          pattern: ""
                          substitution: ""
                      host:
                        fromPath:
                          pattern: ""
                          substitution: ""
                    destination:
                      kuma.io/service: offers
                  - match:
                      path:
                        prefix: "/offers1"
                    modify:
                      path: {}
                      host: {}
                    destination:
                      kuma.io/service: offers
                  destination:
                    kuma.io/service: backend`,
				expected: `
                violations:
                - field: conf.http[0].modify.path.rewritePrefix
                  message: cannot be empty
                - field: conf.http[0].modify.host.value
                  message: cannot be empty
                - field: conf.http[0].modify.requestHeaders.add[0].name
                  message: cannot be empty
                - field: conf.http[0].modify.requestHeaders.add[0].value
                  message: cannot be empty
                - field: conf.http[0].modify.requestHeaders.remove[0].name
                  message: cannot be empty
                - field: conf.http[0].modify.responseHeaders.add[0].name
                  message: cannot be empty
                - field: conf.http[0].modify.responseHeaders.add[0].value
                  message: cannot be empty
                - field: conf.http[0].modify.responseHeaders.remove[0].name
                  message: cannot be empty
                - field: conf.http[1].modify.path.regex.pattern
                  message: cannot be empty
                - field: conf.http[1].modify.path.regex.substitution
                  message: cannot be empty
                - field: conf.http[1].modify.host.fromPath.pattern
                  message: cannot be empty
                - field: conf.http[1].modify.host.fromPath.substitution
                  message: cannot be empty
                - field: conf.http[2].modify.path
                  message: either "rewritePrefix" or "regex" has to be set
                - field: conf.http[2].modify.host
                  message: either "value" or "fromPath" has to be set`,
			}),
			Entry("http - modify path prefix cannot be used if match path prefix is not used", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  http:
                  - match:
                      path:
                        exact: "/offers"
                    modify:
                      path:
                        rewritePrefix: "/not-offers"
                    destination:
                      kuma.io/service: offers
                  destination:
                    kuma.io/service: backend`,
				expected: `
                violations:
                - field: conf.http[0].modify.path.rewritePrefix
                  message: can only be set when .http.match.path.prefix is not empty`,
			}),
			Entry("http - not allow some headers to be modified", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: web
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  http:
                  - match:
                      path:
                        exact: "/offers"
                    modify:
                      requestHeaders:
                        add:
                        - name: 'host'
                          value: xyz
                        - name: 'Host'
                          value: xyz
                        - name: ':path'
                          value: xyz
                        remove:
                        - name: 'host'
                          value: xyz
                      responseHeaders:
                        add:
                        - name: 'host'
                          value: xyz
                        - name: 'Host'
                          value: xyz
                        - name: ':path'
                          value: xyz
                        remove:
                        - name: 'host'
                          value: xyz
                    destination:
                      kuma.io/service: offers
                  destination:
                    kuma.io/service: backend`,
				expected: `
                violations:
                - field: conf.http[0].modify.requestHeaders.add[0].name
                  message: host header and HTTP/2 pseudo-headers are not allowed to be modified
                - field: conf.http[0].modify.requestHeaders.add[1].name
                  message: host header and HTTP/2 pseudo-headers are not allowed to be modified
                - field: conf.http[0].modify.requestHeaders.add[2].name
                  message: host header and HTTP/2 pseudo-headers are not allowed to be modified
                - field: conf.http[0].modify.requestHeaders.remove[0].name
                  message: host header and HTTP/2 pseudo-headers are not allowed to be modified
                - field: conf.http[0].modify.responseHeaders.add[0].name
                  message: host header and HTTP/2 pseudo-headers are not allowed to be modified
                - field: conf.http[0].modify.responseHeaders.add[1].name
                  message: host header and HTTP/2 pseudo-headers are not allowed to be modified
                - field: conf.http[0].modify.responseHeaders.add[2].name
                  message: host header and HTTP/2 pseudo-headers are not allowed to be modified
                - field: conf.http[0].modify.responseHeaders.remove[0].name
                  message: host header and HTTP/2 pseudo-headers are not allowed to be modified`,
			}),
		)
	})
})
