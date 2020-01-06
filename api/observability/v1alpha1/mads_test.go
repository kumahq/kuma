package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/api/observability/v1alpha1"

	util_proto "github.com/Kong/kuma/api/internal/util/proto"
)

var _ = Describe("Monitoring Assignment Discovery Service", func() {

	Describe("MonitoringAssignment", func() {

		type testCase struct {
			input string
		}

		DescribeTable("should to be serializable to YAML",
			func(given testCase) {
				// given
				ma := &MonitoringAssignment{}

				By("deserializing from YAML")
				// when
				err := util_proto.FromYAML([]byte(given.input), ma)
				// then
				Expect(err).ToNot(HaveOccurred())

				By("validating")
				Expect(ma.Validate()).To(Succeed())

				By("serializing back to YAML")
				// when
				actual, err := util_proto.ToYAML(ma)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.input))
			},
			Entry("MonitoringAssignment per service", testCase{
				input: `
                name: /meshes/default/services/backend
                targets:
                - labels:
                    __address__: 192.168.0.1:8080
                    instance: backend-01
                - labels:
                    __address__: 192.168.0.2:8080
                    instance: backend-02
                labels:
                  job: backend
`,
			}),
			Entry("MonitoringAssignment per dataplane", testCase{
				input: `
                name: /meshes/default/dataplane/backend-01
                targets:
                - labels:
                    __address__: 192.168.0.1:8080
                labels:
                  job: backend
                  instance: backend-01
`,
			}),
		)

		It("should require a non-empty name", func() {
			// given
			ma := &MonitoringAssignment{}
			// expect
			Expect(ma.Validate()).To(HaveOccurred())
		})
	})
})
