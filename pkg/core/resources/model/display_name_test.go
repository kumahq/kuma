package model_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var _ = Describe("PluralType", func() {
	type testCase struct {
		name     string
		expected string
	}

	DescribeTable("should compute the correct plural name",
		func(given testCase) {
			pluralResourceName := core_model.PluralType(given.name)
			Expect(pluralResourceName).To(Equal(given.expected))
		},
		Entry("when S ending", testCase{
			name:     "MeshTLS",
			expected: "MeshTLSes",
		}),
		Entry("when y ending", testCase{
			name:     "Meshy",
			expected: "Meshies",
		}),
		Entry("when sh ending", testCase{
			name:     "Mesh",
			expected: "Meshes",
		}),
		Entry("when s ending", testCase{
			name:     "MeshTls",
			expected: "MeshTlses",
		}),
		Entry("when ch ending", testCase{
			name:     "Mech",
			expected: "Meches",
		}),
		Entry("when Y ending", testCase{
			name:     "MesY",
			expected: "Mesies",
		}),
	)
})
