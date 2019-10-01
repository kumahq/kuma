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

		DescribeTable("should apply defaults on a target MeshResource",
			func(given testCase) {
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
			},
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
		)
	})
})
