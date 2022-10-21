package v1alpha1_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshtrafficpermissions_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("MeshTrafficPermission", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtpYAML string) {
				// setup
				meshTrafficPermission := meshtrafficpermissions_proto.NewMeshTrafficPermissionResource()

				// when
				err := util_proto.FromYAML([]byte(mtpYAML), meshTrafficPermission.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTrafficPermission.Validate()

				// then
				Expect(verr).To(BeNil())
			},
			Entry("allow or deny all possible kinds of clients", `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      action: ALLOW
  - targetRef:
      kind: MeshSubset
      tags:
        kuma.io/zone: us-east
        env: dev
    default:
      action: DENY
  - targetRef:
      kind: MeshService
      name: backend
    default:
      action: ALLOW
  - targetRef:
      kind: MeshServiceSubset
      name: backend
      tags:
        version: v1
    default:
      action: DENY
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
      action: DENY
`),
			Entry("allow MeshService at top-level targetRef", `
targetRef:
  kind: MeshService
  name: backend
from:
  - targetRef:
      kind: Mesh
    default:
      action: DENY
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
      action: DENY
`),
			Entry("allow MeshGatewayRoute at top-level targetRef", `
targetRef:
  kind: MeshGatewayRoute
  name: backend-gateway-route
from:
  - targetRef:
      kind: Mesh
    default:
      action: DENY
`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshTrafficPermission := meshtrafficpermissions_proto.NewMeshTrafficPermissionResource()

				// when
				err := util_proto.FromYAML([]byte(given.inputYaml), meshTrafficPermission.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTrafficPermission.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("disallow MeshHTTPRoute at top-level targetRef", testCase{
				inputYaml: `
targetRef:
  kind: MeshHTTPRoute
  name: backend-http-route
from:
  - targetRef:
      kind: Mesh
    default:
      action: DENY
`,
				expected: `
violations:
  - field: spec.targetRef.kind
    message: MeshHTTPRoute is not yet supported`,
			}),
			Entry("empty 'from' array", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
from: []
`,
				expected: `
violations:
  - field: spec.from
    message: needs at least one item`,
			}),
			Entry("empty 'from' array", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
from: []
`,
				expected: `
violations:
  - field: spec.from
    message: needs at least one item`,
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
      action: ALLOW
  - targetRef:
      kind: MeshHTTPRoute
      name: mhr-1
    default:
      action: ALLOW
`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value is not supported
  - field: spec.from[1].targetRef.kind
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
  - field: spec.from[0].default
    message: must be defined 
`,
			}),
		)
	})
})
