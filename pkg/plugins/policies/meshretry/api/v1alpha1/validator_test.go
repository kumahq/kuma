package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshretry_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
)

var _ = Describe("MeshRetry", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtpYAML string) {
				// setup
				mtp := meshretry_proto.NewMeshRetryResource()

				// when
				err := core_model.FromYAML([]byte(mtpYAML), &mtp.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := mtp.Validate()

				// then
				Expect(verr).To(BeNil())
			},
			Entry("full example", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      tcp:
        maxConnectAttempt: 5
  - targetRef:
      kind: MeshService
      name: backend
    default:
      tcp:
        maxConnectAttempt: 5
      http:
        numRetries: 10
        perTryTimeout: 20s
        backOff:
          baseInterval: 15s
          maxInterval: 20m
        rateLimitedBackOff:
          resetHeaders:
            - name: retry-after-http
              format: Seconds
            - name: x-retry-after-http
              format: UnixTimestamp
          maxInterval: 21m
        retryOn:
          - 5xx
          - GatewayError
          - Reset
          - Retriable4xx
          - ConnectFailure
          - EnvoyRatelimited
          - RefusedStream
          - Http3PostConnectFailure
          - HttpMethodConnect
          - HttpMethodDelete
          - HttpMethodGet
          - HttpMethodHead
          - HttpMethodOptions
          - HttpMethodPatch
          - HttpMethodPost
          - HttpMethodPut
          - HttpMethodTrace
          - 500
          - 409
          - 503
        retriableRequestHeaders:
          - type: PREFIX
            name: x-my-header-1
            value: kuma-value-
          - type: REGULAR_EXPRESSION
            name: x-my-header-2
            value: ".*"
          - type: EXACT
            name: x-my-header-3
            value: exact-value
        retriableResponseHeaders:
          - type: PREFIX
            name: x-my-header-4
            value: kuma-value-
          - type: REGULAR_EXPRESSION
            name: x-my-header-5
            value: ".*"
          - type: EXACT
            name: x-my-header-6
            value: exact-value
      grpc:
        numRetries: 10
        perTryTimeout: 20s
        backOff:
          baseInterval: 15s
          maxInterval: 20m
        rateLimitedBackOff:
          resetHeaders:
            - name: retry-after-grpc
              format: Seconds
            - name: x-retry-after-grpc
              format: UnixTimestamp
          maxInterval: 21m
        retryOn:
          - Canceled
          - DeadlineExceeded
          - Internal
          - ResourceExhausted
          - Unavailable
`),
			Entry("minimalistic http retry", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      http:
        numRetries: 5
`),
			Entry("empty http arrays", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      http:
        numRetries: 5
        retryOn: []
        retriableRequestHeaders: []
        retriableResponseHeaders: []
`),
			Entry("minimalistic grpc retry", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      grpc:
        numRetries: 5
`),
			Entry("empty grpc arrays", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      grpc:
        numRetries: 5
        retryOn: []
`),
			Entry("http.retryOn with status codes", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      http:
        retryOn: 
          - 500
          - 409
`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				mtp := meshretry_proto.NewMeshRetryResource()

				// when
				err := core_model.FromYAML([]byte(given.inputYaml), &mtp.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := mtp.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty 'to' array", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
to: []
`,
				expected: `
violations:
  - field: spec.to
    message: needs at least one item`,
			}),
			Entry("unsupported targetRef kinds in 'to'", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshSubset
      tags:
        version: v1
    default:
      http:
        numRetries: 1
  - targetRef:
      kind: MeshServiceSubset
      name: backend
      tags:
        kuma.io/zone: us-east
    default:
      grpc:
        numRetries: 1
`,
				expected: `
violations:
  - field: spec.to[0].targetRef.kind
    message: value is not supported
  - field: spec.to[1].targetRef.kind
    message: value is not supported
`,
			}),
			Entry("empty 'default", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.conf
    message: at least one of the 'tcp', 'http', 'grpc' must be specified`,
			}),
			Entry("empty root sections in 'default'", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      tcp: {}
      http: {}
      grpc: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.conf.tcp
    message: must not be empty
  - field: spec.to[0].default.conf.http
    message: must not be empty
  - field: spec.to[0].default.conf.grpc
    message: must not be empty`,
			}),
			Entry("empty http.backOff", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      http: 
        backOff: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.conf.http.backOff
    message: must not be empty`,
			}),
			Entry("empty http.rateLimitedBackOff.resetHeaders", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      http: 
        rateLimitedBackOff: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.conf.http.rateLimitedBackOff.resetHeaders
    message: must be defined`,
			}),
			Entry("empty grpc.backOff", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      grpc: 
        backOff: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.conf.grpc.backOff
    message: must not be empty`,
			}),
			Entry("empty grpc.rateLimitedBackOff.resetHeaders", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      grpc: 
        rateLimitedBackOff: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.conf.grpc.rateLimitedBackOff.resetHeaders
    message: must be defined`,
			}),
			Entry("http.retryOn with not allowed values", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      http: 
        retryOn: 
          - WRONG_VALUE
          - 5XX
          - reset
          - DEADLINE_EXCEEDED
          - 123
          - 952
`,
				expected: `
violations:
  - field: spec.to[0].default.conf.http.retryOn[0]
    message: unknown item 'WRONG_VALUE'
  - field: spec.to[0].default.conf.http.retryOn[1]
    message: unknown item '5XX'
  - field: spec.to[0].default.conf.http.retryOn[2]
    message: unknown item 'reset'
  - field: spec.to[0].default.conf.http.retryOn[3]
    message: unknown item 'DEADLINE_EXCEEDED'
  - field: spec.to[0].default.conf.http.retryOn[4]
    message: unknown item '123'
  - field: spec.to[0].default.conf.http.retryOn[5]
    message: unknown item '952'`,
			}),
			Entry("grpc.retryOn with not allowed values", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      grpc: 
        retryOn: 
          - 500
          - reset
          - wrong
`,
				expected: `
violations:
  - field: spec.to[0].default.conf.grpc.retryOn[0]
    message: unknown item '500'
  - field: spec.to[0].default.conf.grpc.retryOn[1]
    message: unknown item 'reset'
  - field: spec.to[0].default.conf.grpc.retryOn[2]
    message: unknown item 'wrong'`,
			}),
			Entry("http retriableHeaders wrong values", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      http: 
        retriableRequestHeaders: 
          - type: WRONG_TYPE
            name: headername
            value: headervalue
        retriableResponseHeaders: 
          - type: ANOTHER_WRONG_TYPE
            name: headername
            value: headervalue
`,
				expected: `
violations:
  - field: spec
    message: to[0].default.http.retriableRequestHeaders[0].type in body should be one of [REGULAR_EXPRESSION EXACT PREFIX]
  - field: spec
    message: to[0].default.http.retriableResponseHeaders[0].type in body should be one of [REGULAR_EXPRESSION EXACT PREFIX]`,
			}),

			Entry("empty http.rateLimitedBackOff.resetHeaders.name and wrong format", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
    default:
      http: 
        rateLimitedBackOff:
          resetHeaders:
            - name: ""
`,
				expected: `
violations:
  - field: spec.to[0].default.conf.http.rateLimitedBackOff.resetHeaders[0].name
    message: must be defined
  - field: spec.to[0].default.conf.http.rateLimitedBackOff.resetHeaders[0].format
    message: 'must be only one of: Seconds, UnixTimestamp'`,
			}),
		)
	})
})
