package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/api/mesh/v1alpha1"

	util_proto "github.com/Kong/kuma/api/internal/util/proto"
)

var _ = Describe("HealthCheck", func() {

	Context("valid configurations", func() {
		type testCase struct {
			input string
		}

		DescribeTable("Validate() should return a nil",
			func(given testCase) {
				// setup
				check := &HealthCheck{}

				// when
				err := util_proto.FromYAML([]byte(given.input), check)
				// then
				Expect(err).ToNot(HaveOccurred())

				// expect
				Expect(check.Validate()).To(Succeed())
			},
			Entry("conf with active health checks", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
                conf:
                  activeChecks:
                    interval: 10s
                    timeout: 2s
                    unhealthyThreshold: 3
                    healthyThreshold: 1
`,
			}),
			Entry("conf with passive health checks", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
                conf:
                  passiveChecks:
                    unhealthyThreshold: 3
                    penaltyInterval: 5s
`,
			}),
			Entry("conf with both active and passive health checks", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
                conf:
                  activeChecks:
                    interval: 10s
                    timeout: 2s
                    unhealthyThreshold: 3
                    healthyThreshold: 1
                  passiveChecks:
                    unhealthyThreshold: 3
                    penaltyInterval: 5s
`,
			}),
			Entry("no conf", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
`,
			}),
		)
	})

	Context("invalid configurations", func() {
		type testCase struct {
			input       string
			expectedErr interface{}
		}

		DescribeTable("Validate() should return an error",
			func(given testCase) {
				// setup
				check := &HealthCheck{}

				// when
				err := util_proto.FromYAML([]byte(given.input), check)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				err = check.Validate()
				// then
				Expect(err.Error()).To(Equal(given.expectedErr))
			},
			Entry("0 sources", testCase{
				input:       ``,
				expectedErr: `invalid HealthCheck.Sources: value must contain at least 1 item(s)`,
			}),
			Entry("0 destinations", testCase{
				input: `
                sources:
                - match:
                    service: web
`,
				expectedErr: `invalid HealthCheck.Destinations: value must contain at least 1 item(s)`,
			}),
			Entry("incomplete active checks conf", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
                conf:
                  activeChecks: {}
`,
				expectedErr: `invalid HealthCheck.Conf: embedded message failed validation | caused by: invalid HealthCheck_Conf.ActiveChecks: embedded message failed validation | caused by: invalid HealthCheck_Conf_Active.Interval: value is required`,
			}),
			Entry("incomplete passive checks conf", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
                conf:
                  passiveChecks: {}
`,
				expectedErr: `invalid HealthCheck.Conf: embedded message failed validation | caused by: invalid HealthCheck_Conf.PassiveChecks: embedded message failed validation | caused by: invalid HealthCheck_Conf_Passive.UnhealthyThreshold: value must be greater than 0`,
			}),
		)
	})

})
