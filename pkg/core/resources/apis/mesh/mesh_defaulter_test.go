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

			// do
			mesh.Default()

			// when
			actual, err := util_proto.ToYAML(&mesh.Spec)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		}

		DescribeTable("should apply defaults on a target MeshResource",
			applyDefaultsScenario,
			Entry("when `mtls` field is not set", testCase{
				input: ``,
				expected: `
                mtls:
                  ca:
                    builtin: {}
`,
			}),
			Entry("when `mtls.ca` field is not set", testCase{
				input: `
                mtls: {}
`,
				expected: `
                mtls:
                  ca:
                    builtin: {}
`,
			}),
			Entry("when `mtls.ca.type` field is not set", testCase{
				input: `
                mtls:
                  ca: {}
`,
				expected: `
                mtls:
                  ca:
                    builtin: {}
`,
			}),
			Entry("when both `metrics.prometheus.port` and `metrics.prometheus.path` are not set", testCase{
				input: `
                metrics:
                  prometheus: {}
`,
				expected: `
                mtls:
                  ca:
                    builtin: {}
                metrics:
                  prometheus:
                    port: 5670
                    path: /metrics
`,
			}),
			Entry("when `metrics.prometheus.port` is not set", testCase{
				input: `
                metrics:
                  prometheus:
                    path: /non-standard-path
`,
				expected: `
                mtls:
                  ca:
                    builtin: {}
                metrics:
                  prometheus:
                    port: 5670
                    path: /non-standard-path
`,
			}),
			Entry("when `metrics.prometheus.path` is not set", testCase{
				input: `
                metrics:
                  prometheus:
                    port: 1234
`,
				expected: `
                mtls:
                  ca:
                    builtin: {}
                metrics:
                  prometheus:
                    port: 1234
                    path: /metrics
`,
			}),
		)

		DescribeTable("should not override user-defined configuration in a target MeshResource",
			applyDefaultsScenario,
			Entry("when `mtls` field is set to `provided` CA", testCase{
				input: `
                mtls:
                  ca:
                    provided: {}
`,
				expected: `
                mtls:
                  ca:
                    provided: {}
`,
			}),
			Entry("when `metrics` field is not set", testCase{
				input: ``,
				expected: `
                mtls:
                  ca:
                    builtin: {}
`,
			}),
			Entry("when `metrics.prometheus` field is not set", testCase{
				input: `
                metrics: {}
`,
				expected: `
                mtls:
                  ca:
                    builtin: {}
                metrics: {}
`,
			}),
			Entry("when `mtls.ca.type` field is not set", testCase{
				input: `
                metrics:
                  prometheus:
                    port: 1234
                    path: /non-standard-path
`,
				expected: `
                mtls:
                  ca:
                    builtin: {}
                metrics:
                  prometheus:
                    port: 1234
                    path: /non-standard-path
`,
			}),
		)
	})
})
