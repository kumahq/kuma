package v1alpha1_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshaccesslog_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("MeshAccessLog", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtpYAML string) {
				// setup
				meshAccessLog := meshaccesslog_proto.NewMeshAccessLogResource()

				// when
				err := util_proto.FromYAML([]byte(mtpYAML), meshAccessLog.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshAccessLog.Validate()

				// then
				Expect(verr).To(BeNil())
			},
			Entry("example", `
targetRef:
  kind: MeshService
  name: web-backend
  tags:
    kuma.io/zone: us-east
from:
  - targetRef:
      kind: MeshService
      name: web-frontend
    default:
      backends:
        - tcp:
            conf:
              format:
                json:
                  - key: "start_time"
                    value: "%START_TIME%"
              address: 127.0.0.1:5000
        - reference:
            conf: 
              kind: MeshAccessLogBackend
              name: file-backend
`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshAccessLog := meshaccesslog_proto.NewMeshAccessLogResource()

				// when
				err := util_proto.FromYAML([]byte(given.inputYaml), meshAccessLog.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshAccessLog.Validate()
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
  - field: spec.from
    message: cannot be empty`,
			}),
		)
	})
})
