package v1alpha1_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshtrace_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("MeshTrace", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtpYAML string) {
				// setup
				meshTrace := meshtrace_proto.NewMeshTraceResource()

				// when
				err := util_proto.FromYAML([]byte(mtpYAML), meshTrace.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTrace.Validate()

				// then
				Expect(verr).To(BeNil())
			},
			Entry("full example", `
apiVersion: kuma.io/v1alpha1
kind: MeshTrace
metadata:
  name: all
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshService
    name: backend
  default:
    backend:
      zipkin:
        url: http://jaeger-collector.mesh-observability:9411/api/v2/spans
    tags:
	  - name: team
		literal: core
	  - name: env
        header:
		  name: x-env
		  default: prod
    sampling:
      overall: 80
      random: 60
      client: 40
`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshTrace := meshtrace_proto.NewMeshTraceResource()

				// when
				err := util_proto.FromYAML([]byte(given.inputYaml), meshTrace.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTrace.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty 'from' and 'to' array", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
  name: default
`,
				expected: `
violations:
  - field: spec
    message: at least one of 'from', 'to' has to be defined`,
			}),
		)
	})
})
