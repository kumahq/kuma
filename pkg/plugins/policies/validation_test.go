package policies_test

import (
    "github.com/ghodss/yaml"
    "github.com/kumahq/kuma/pkg/plugins/policies"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("policies validation", func() {
    Describe("ValidateSchema()", func() {
        DescribeTable("valid MeshTrafficPermission should pass validation",
            func(resourceYaml string) {
                json, err := yaml.YAMLToJSON([]byte(resourceYaml))
                Expect(err).To(Not(HaveOccurred()))
                verr := policies.ValidateSchema(string(json), "MeshTrafficPermission")

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
                json, err := yaml.YAMLToJSON([]byte(given.inputYaml))
                Expect(err).To(Not(HaveOccurred()))
                verr := policies.ValidateSchema(string(json), "MeshTrafficPermission")
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
