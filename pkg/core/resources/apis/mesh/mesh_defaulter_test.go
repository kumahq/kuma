package mesh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/core/resources/apis/mesh"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("MeshResource", func() {

	Describe("Default()", func() {

		type testCase struct {
			input    string
			expected string
		}

		applyDefaultsScenario := func(given testCase) {
			// given
			mesh := &MeshResource{}
			err := util_proto.FromYAML([]byte(given.input), &mesh.Spec)
			Expect(err).ToNot(HaveOccurred())

			// when
			err = mesh.Default()

			// then
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(&mesh.Spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		}

		DescribeTable("should apply defaults on a target MeshResource",
			applyDefaultsScenario,
			Entry("when defaults are not set", testCase{
				input: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
`,
				expected: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      port: 5670
                      path: /metrics
                      tags:
                        service: dataplane-metrics
`,
			}),
			Entry("when defaults are set", testCase{
				input: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      path: /non-standard-path
                      port: 1234
                      tags:
                        service: dataplane-metrics
                      skipMTLS: true
`,
				expected: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      path: /non-standard-path
                      port: 1234
                      tags:
                        service: dataplane-metrics
                      skipMTLS: true
`,
			}),
		)

		DescribeTable("should not override user-defined configuration in a target MeshResource",
			applyDefaultsScenario,
			Entry("when `metrics.prometheus` field is not set", testCase{
				input: `
                metrics: {}
`,
				expected: `
                metrics: {}
`,
			}),
		)
	})
})
