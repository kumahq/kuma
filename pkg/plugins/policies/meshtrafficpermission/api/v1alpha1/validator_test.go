package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshtrafficpermissions_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
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
        - spiffeId:
            type: Exact
            value: spiffe://trust.domain/service
      allow:
        - spiffeId:
            type: PathPrefix
            value: spiffe://trust.domain/service
      allowWithShadowDeny:
        - spiffeId:
            type: Exact
            value: spiffe://trust.domain/service-2
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
    message: value is not supported
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
			Entry("matches with invalid spiffe id", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - default:
      deny: 
        - spiffeId:
            type: Exact
            value: some-service
      allow:
        - spiffeId:
            type: PathPrefix
            value: wrong
      allowWithShadowDeny:
        - spiffeId:
            type: Exact
            value: test
`,
				expected: `
violations:
  - field: spec.rules[0].allow[0].spiffeId
    message: must be a valid Spiffe ID
  - field: spec.rules[0].allowWithShadowDeny[0].spiffeId
    message: must be a valid Spiffe ID
  - field: spec.rules[0].deny[0].spiffeId
    message: must be a valid Spiffe ID
`,
			}),
		)
	})
})
