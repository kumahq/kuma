package validation_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshtrace_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("plugins validation", func() {
	Describe("ValidateResourceSchema()", func() {
		DescribeTable("valid MeshTrafficPermission should pass validation",
			func(resourceYaml string) {
				mtp := meshtrace_proto.NewMeshTrafficPermissionResource()
				err := proto.FromYAML([]byte(resourceYaml), mtp.Spec)
				Expect(err).To(Not(HaveOccurred()))
				verr := mtp.Validate()

				Expect(verr).To(BeNil())
			},
			Entry("valid MeshTrafficPermission", `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      action: ALLOW
`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate schema and return as accurate errors as possible",
			func(given testCase) {
				// and
				mtp := meshtrace_proto.NewMeshTrafficPermissionResource()
				err := proto.FromYAML([]byte(given.inputYaml), mtp.Spec)
				Expect(err).To(Not(HaveOccurred()))
				verr := mtp.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("unknown fields", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      action: foo
`,
				expected: `
violations:
  - field: from.0.default.action
    message: 'from.0.default.action must be one of the following: "ALLOW", "DENY", "ALLOW_WITH_SHADOW_DENY", "DENY_WITH_SHADOW_ALLOW"'`,
			}),
		)
	})
})
