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
			Entry("when both `metrics.prometheus.port` and `metrics.prometheus.path` are not set", testCase{
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
`,
			}),
			Entry("when `metrics.prometheus.port` is not set", testCase{
				input: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      path: /non-standard-path
`,
				expected: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      port: 5670
                      path: /non-standard-path
`,
			}),
			Entry("when `metrics.prometheus.path` is not set", testCase{
				input: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      port: 1234
`,
				expected: `
                metrics:
                  enabledBackend: prometheus-1
                  backends:
                  - name: prometheus-1
                    type: prometheus
                    conf:
                      port: 1234
                      path: /metrics
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
