package v1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/api/observability/v1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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
mesh: default
service: backend
targets:
- address: 192.168.0.1:8080
  name: backend-01
  scheme: http
- address: 192.168.0.2:8080
  name: backend-02
  scheme: http
labels:
  team: infra`,
			}),
		)
	})
})
