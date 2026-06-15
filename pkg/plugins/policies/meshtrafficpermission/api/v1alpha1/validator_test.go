package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	meshtrafficpermissions_proto "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

var _ = Describe("MeshTrafficPermission", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtpYAML string) {
				// setup
				mtp := meshtrafficpermissions_proto.NewMeshTrafficPermissionResource()

				// when
				err := core_model.FromYAML([]byte(mtpYAML), &mtp.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := mtp.Validate()

				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("allow or deny all possible kinds of clients", `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      action: Allow
  - targetRef:
      kind: MeshSubset
      tags:
        kuma.io/zone: us-east
        env: dev
    default:
      action: Deny
  - targetRef:
      kind: MeshService
      name: backend
    default:
      action: Allow
  - targetRef:
      kind: MeshServiceSubset
      name: backend
      tags:
        version: v1
    default:
      action: Deny
`),
			Entry("allow MeshSubset at top-level targetRef", `
targetRef:
  kind: MeshSubset
  tags:
    env: prod
from:
  - targetRef:
      kind: Mesh
    default:
      action: Deny
`),
			Entry("allow MeshService at top-level targetRef", `
targetRef:
  kind: MeshService
  name: backend
from:
  - targetRef:
      kind: Mesh
    default:
      action: Deny
`),
			Entry("allow MeshServiceSubset at top-level targetRef", `
targetRef:
  kind: MeshServiceSubset
  name: backend
  tags:
    version: v2
from:
  - targetRef:
      kind: Mesh
    default:
      action: Deny
`),
			Entry("full rules example", `
targetRef:
  kind: Mesh
rules:
  - default:
      deny:
        - spiffeID:
            type: Exact
            value: spiffe://trust.domain/service
      allow:
        - spiffeID:
            type: Prefix
            value: spiffe://trust.domain
      allowWithShadowDeny:
        - spiffeID:
            type: Exact
            value: spiffe://trust.domain/service-2
`),
			Entry("sni-only allow", `
targetRef:
  kind: Dataplane
  sectionName: ze-port
rules:
  - default:
      allow:
        - sni:
            type: Exact
            value: sni.extsvc.default.zone-1.aws-aurora.8443
`),
			Entry("spiffeID and sni combined in the same match", `
targetRef:
  kind: Dataplane
  sectionName: ze-port
rules:
  - default:
      allow:
        - spiffeID:
            type: Exact
            value: spiffe://default/ns/backend-ns/sa/backend
          sni:
            type: Exact
            value: sni.extsvc.default.zone-1.aws-aurora.8443
`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				mtp := meshtrafficpermissions_proto.NewMeshTrafficPermissionResource()

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
			Entry("empty 'from' array", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
from: []
`,
				expected: `
violations:
  - field: spec
    message: at least one of 'from' or 'rules' has to be defined`,
			}),
			Entry("empty 'rules' array", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
rules: []
`,
				expected: `
violations:
  - field: spec
    message: at least one of 'from' or 'rules' has to be defined`,
			}),
			Entry("sectionName without from or rules", testCase{
				inputYaml: `
targetRef:
  kind: Dataplane
  sectionName: test
to: []
`,
				expected: `
violations:
- field: spec.targetRef.sectionName
  message: can only be used with inbound policies
- field: spec
  message: at least one of 'from' or 'rules' has to be defined`,
			}),
			Entry("not supported kinds in 'from' array", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
from: 
  - targetRef:
      kind: MeshGatewayRoute
      name: mgr-1
    default:
      action: Allow
`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value 'MeshGatewayRoute' is not supported
`,
			}),
			Entry("default is nil", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
`,
				expected: `
violations:
  - field: spec.from[0].default.action
    message: must be defined 
`,
			}),
			Entry("rules default is nil", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - default:
      deny: []
`,
				expected: `
violations:
  - field: spec.rules[0]
    message: at least one of 'allow', 'allowWithShadowDeny', 'deny' has to be defined
`,
			}),
			Entry("matches with invalid sni", testCase{
				inputYaml: `
targetRef:
  kind: Dataplane
  sectionName: ze-port
rules:
  - default:
      allow:
        - sni:
            type: Exact
            value: ""
      deny:
        - sni:
            type: Exact
      allowWithShadowDeny:
        - sni:
            type: Exact
            value: "not_valid"
`,
				expected: `
violations:
  - field: spec.rules[0].allow[0].sni.value
    message: must be set
  - field: spec.rules[0].allowWithShadowDeny[0].sni.value
    message: a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')
  - field: spec.rules[0].deny[0].sni.value
    message: must be set
`,
			}),
			Entry("matches with invalid spiffe id", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - default:
      deny: 
        - spiffeID:
            type: Exact
            value: some-service
      allow:
        - spiffeID:
            type: Prefix
            value: wrong
      allowWithShadowDeny:
        - spiffeID:
            type: Exact
            value: test
`,
				expected: `
violations:
  - field: spec.rules[0].allow[0].spiffeID
    message: 'must be a valid Spiffe ID: scheme is missing or invalid'
  - field: spec.rules[0].allowWithShadowDeny[0].spiffeID
    message: 'must be a valid Spiffe ID: scheme is missing or invalid'
  - field: spec.rules[0].deny[0].spiffeID
    message: 'must be a valid Spiffe ID: scheme is missing or invalid'
`,
			}),
		)
	})
})
