package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	_ "github.com/kumahq/kuma/pkg/plugins/policies"
	meshtrace_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	meshtrafficpermissions_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

var _ = Describe("plugins validation", func() {
	Describe("ValidateResourceSchema()", func() {
		type testCase struct {
			inputYaml    string
			resourceType core_model.ResourceType
			expected     string
		}
		DescribeTable("valid MeshTrafficPermission should pass validation",
			func(given testCase) {
				obj, err := registry.Global().NewObject(given.resourceType)
				Expect(err).To(Not(HaveOccurred()))

				err = core_model.FromYAML([]byte(given.inputYaml), obj.GetSpec())
				Expect(err).To(Not(HaveOccurred()))

				verr := obj.(core_model.ResourceValidator).Validate()
				Expect(verr).To(BeNil())
			},
			Entry("valid MeshTrafficPermission", testCase{
				resourceType: meshtrafficpermissions_proto.MeshTrafficPermissionType,
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      action: ALLOW
`,
			}),
			Entry("valid MeshTrace", testCase{
				resourceType: meshtrace_proto.MeshTraceType,
				inputYaml: `
targetRef:
  kind: Mesh
default:
  backends: []
  tags: null
`,
			}),
		)

		DescribeTable("should validate schema and return as accurate errors as possible",
			func(given testCase) {
				// and
				mtp := meshtrafficpermissions_proto.NewMeshTrafficPermissionResource()
				err := core_model.FromYAML([]byte(given.inputYaml), mtp.Spec)
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
  - field: spec
    message: from[0].default.action in body should be one of [ALLOW DENY ALLOW_WITH_SHADOW_DENY]`,
			}),
		)
	})
})
