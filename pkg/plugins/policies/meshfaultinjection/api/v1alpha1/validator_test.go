package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshfaultinjection_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
)

var _ = Describe("MeshFaultInjection", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mhcYAML string) {
				// setup
				meshFaultInjection := meshfaultinjection_proto.NewMeshFaultInjectionResource()

				// when
				err := core_model.FromYAML([]byte(mhcYAML), &meshFaultInjection.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshFaultInjection.Validate()

				// then
				Expect(verr).To(BeNil())
			},
			Entry("full example", `
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http:
        - abort:
            httpStatus: 503
            percentage: 50
          delay:
            value: 10s
            percentage: 5
        - delay:
            value: 5s
            percentage: 5
        - responseBandwidth:
            limit: 100mbps
            percentage: 5
        - abort:
            httpStatus: 500
            percentage: 50
      `),
			Entry("full example", `
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http: []
      `),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshFaultInjection := meshfaultinjection_proto.NewMeshFaultInjectionResource()

				// when
				err := core_model.FromYAML([]byte(given.inputYaml), &meshFaultInjection.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshFaultInjection.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("percentages are out of range and some values incorrect", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
from:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http:
      - abort:
          httpStatus: 677
          percentage: 111
      - delay: 
          value: 5s
          percentage: 1111
      - responseBandwidth:
          limit: 1000
          percentage: 1111
`,
				expected: `
violations:
  - field: spec.from[0].default.http.abort[0].httpStatus
    message: must be in range [100, 600)
  - field: spec.from[0].default.http.abort[0].percentage
    message: has to be in [0 - 100] range
  - field: spec.from[0].default.http.delay[1].percentage
    message: has to be in [0 - 100] range
  - field: spec.from[0].default.http.responseBandwidth[2].responseBandwidth
    message: has to be in kbps/mbps/gbps units
  - field: spec.from[0].default.http.responseBandwidth[2].percentage
    message: has to be in [0 - 100] range
`}),
			Entry("percentage is missing", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
from:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http:
      - abort:
          httpStatus: 677
      - delay: {}
      - responseBandwidth:
          limit: 1000
`,
				expected: `
violations:
  - field: spec.from[0].default.http.abort[0].httpStatus
    message: must be in range [100, 600)
  - field: spec.from[0].default.http.abort[0].percentage
    message: must be defined
  - field: spec.from[0].default.http.delay[1].percentage
    message: must be defined
  - field: spec.from[0].default.http.responseBandwidth[2].responseBandwidth
    message: has to be in kbps/mbps/gbps units
  - field: spec.from[0].default.http.responseBandwidth[2].percentage
    message: must be defined
`}),
		)
	})
})
