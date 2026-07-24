package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

var _ = Describe("MeshTimeout", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtYAML string) {
				// setup
				meshTimeout := NewMeshTimeoutResource()

				// when
				err := core_model.FromYAML([]byte(mtYAML), &meshTimeout.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTimeout.Validate()

				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("full example", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: 10s
      idleTimeout: 1h
      http:
        requestTimeout: 1s
        streamIdleTimeout: 1h
        maxStreamDuration: 1h
        maxConnectionDuration: 1h`),
			Entry("only to targetRef", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: 10s
      idleTimeout: 1h
      http:
        requestTimeout: 1s
        streamIdleTimeout: 1h
        maxStreamDuration: 1h
        maxConnectionDuration: 1h`),
			Entry("minimal example", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http:
        requestTimeout: 1s`),
			Entry("example MeshExternalService", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshExternalService
      name: web-backend
    default:
      http:
        requestTimeout: 1s
  - targetRef:
      kind: MeshExternalService
      labels:
        kuma.io/display-name: web-backend
    default:
      http:
        requestTimeout: 1s`),
			Entry("example MeshHTTPRoute", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshHTTPRoute
      name: http-route-1
    default:
      http:
        requestTimeout: 1s`),
			Entry("example MeshHTTPRoute with labels", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshHTTPRoute
      labels:
        kuma.io/display-name: http-route-1
    default:
      http:
        requestTimeout: 1s`),
			Entry("matched inbound rule with route timeouts", `
targetRef:
  kind: Mesh
rules:
  - matches:
      - spiffeID:
          type: Exact
          value: spiffe://default/client
        sni:
          type: Exact
          value: backend.mesh
    default:
      http:
        requestTimeout: 1s
        streamIdleTimeout: 2s`),
			Entry("inbound rules and outbound to together", `
targetRef:
  kind: Mesh
rules:
  - default:
      connectionTimeout: 10s
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: 10s`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshTimeout := NewMeshTimeoutResource()

				// when
				err := core_model.FromYAML([]byte(given.inputYaml), &meshTimeout.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTimeout.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty 'to' array", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
`,
				expected: `
violations:
  - field: spec
    message: at least one of 'to' or 'rules' has to be defined`,
			}),
			Entry("unsupported kind in to selector", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshServiceSubset
    default:
      http:
        requestTimeout: 1s`,
				expected: `
violations:
  - field: spec.to[0].targetRef.kind
    message: value 'MeshServiceSubset' is not supported`,
			}),
			Entry("sectionName with outbound policy", testCase{
				inputYaml: `
targetRef:
  kind: Dataplane
  sectionName: test
to:
  - targetRef:
      kind: MeshServiceSubset
    default:
      http:
        requestTimeout: 1s`,
				expected: `
violations:
  - field: spec.targetRef.sectionName
    message: can only be used with inbound policies
  - field: spec.to[0].targetRef.kind
    message: value 'MeshServiceSubset' is not supported`,
			}),
			Entry("missing timeout configuration", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: web-backend`,
				expected: `
violations:
  - field: spec.to[0].default
    message: at least one timeout should be configured`,
			}),
			Entry("timeout cannot be negative", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      idleTimeout: -1s`,
				expected: `
violations:
  - field: spec.to[0].default.idleTimeout
    message: must not be negative when defined`,
			}),
			Entry("multiple timeout cannot be negative", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: -1s
      http:
        requestTimeout: -10s`,
				expected: `
violations:
  - field: spec.to[0].default.connectionTimeout
    message: must be greater than zero when defined
  - field: spec.to[0].default.http.requestTimeout
    message: must not be negative when defined`,
			}),
			Entry("multiple timeout cannot be negative", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http: {}`,
				expected: `
violations:
  - field: spec.to[0].default.http
    message: at least one timeout in this section should be configured`,
			}),
			Entry("MeshHTTPRoute with sectionName in to", testCase{
				inputYaml: `
to:
  - targetRef:
      kind: MeshHTTPRoute
      name: http-route-1
      sectionName: some-section
    default:
      http:
        requestTimeout: 1s`,
				expected: `
violations:
  - field: spec.to[0].targetRef.sectionName
    message: must not be set with kind MeshHTTPRoute`,
			}),
			Entry("top-level targetRef is referencing MeshHTTPRoute", testCase{
				inputYaml: `
targetRef:
  kind: MeshHTTPRoute
  name: route-1
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: 10s
      idleTimeout: 1h
      http:
        requestTimeout: 1s
        streamIdleTimeout: 1h
        maxStreamDuration: 1h
        maxConnectionDuration: 1h`,
				expected: `
violations:
  - field: spec.targetRef.kind
    message: value 'MeshHTTPRoute' is not supported
  - field: spec.to[0].targetRef.kind
    message: value 'MeshService' is not supported
  - field: spec.to[0].default.connectionTimeout
    message: can't be specified when top-level TargetRef is referencing MeshHTTPRoute
  - field: spec.to[0].default.idleTimeout
    message: can't be specified when top-level TargetRef is referencing MeshHTTPRoute
  - field: spec.to[0].default.http.maxStreamDuration
    message: can't be specified when top-level TargetRef is referencing MeshHTTPRoute
  - field: spec.to[0].default.http.maxConnectionDuration
    message: can't be specified when top-level TargetRef is referencing MeshHTTPRoute`,
			}),
			Entry("to TargetRef using labels and name for MeshExternalService", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshExternalService
      name: web-backend
      labels:
        kuma.io/display-name: web-backend
    default:
      connectionTimeout: 10s
      idleTimeout: 1h`,
				expected: `
violations:
  - field: spec.to[0].targetRef.labels
    message: either labels or name must be specified`,
			}),
			Entry("rules with empty spec", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - default: {}`,
				expected: `
violations:
  - field: spec.rules[0].default
    message: at least one timeout should be configured`,
			}),
			Entry("empty match entry", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - matches:
      - {}
    default:
      http:
        requestTimeout: 1s`,
				expected: `
violations:
  - field: spec.rules[0].matches[0]
    message: must specify at least one of 'spiffeID' or 'sni'`,
			}),
			Entry("spiffeID match without type", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - matches:
      - spiffeID:
          value: spiffe://default/client
    default:
      http:
        requestTimeout: 1s`,
				expected: `
violations:
  - field: spec.rules[0].matches[0].spiffeID.type
    message: must be set`,
			}),
			Entry("spiffeID match with unrecognized type", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - matches:
      - spiffeID:
          type: Regex
          value: spiffe://default/client
    default:
      http:
        requestTimeout: 1s`,
				expected: `
violations:
  - field: spec.rules[0].matches[0].spiffeID.type
    message: 'unrecognized type "Regex", supported values are: Exact, Prefix'`,
			}),
			Entry("spiffeID match with unsupported timeout fields", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - matches:
      - spiffeID:
          type: Exact
          value: spiffe://default/client
    default:
      connectionTimeout: 10s
      http:
        requestHeadersTimeout: 2s
        maxStreamDuration: 3s`,
				expected: `
violations:
  - field: spec.rules[0].default.connectionTimeout
    message: can't be specified when matches contain spiffeID because this field cannot be conditioned on source identity
  - field: spec.rules[0].default.http.requestHeadersTimeout
    message: can't be specified when matches contain spiffeID because this field cannot be conditioned on source identity
  - field: spec.rules[0].default.http.maxStreamDuration
    message: can't be specified when matches contain spiffeID because this field cannot be conditioned on source identity`,
			}),
		)
	})
})
