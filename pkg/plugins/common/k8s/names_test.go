package k8s_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/plugins/common/k8s"

	core_model "github.com/Kong/kuma/pkg/core/resources/model"
)

var _ = Describe("DimensionalResourceName()", func() {
	type testCase struct {
		namespace string
		name      string
		expected  core_model.DimensionalResourceName
	}

	DescribeTable("should return a correct dimensional name",
		func(given testCase) {
			// when
			actual := DimensionalResourceName(given.namespace, given.name)
			// then
			Expect(actual).To(HaveKey("k8s.kuma.io/namespace"), "DO NOT change dimension identifiers lightly! They are considered a part of user-facing Kuma API")
			Expect(actual).To(HaveKey("k8s.kuma.io/name"), "DO NOT change dimension identifiers lightly! They are considered a part of user-facing Kuma API")
			// then
			Expect(actual).To(Equal(given.expected))
		},
		Entry("namespace-scoped k8s resource", testCase{
			namespace: "my-namespace",
			name:      "my-policy",
			expected: core_model.DimensionalResourceName{
				"k8s.kuma.io/namespace": "my-namespace",
				"k8s.kuma.io/name":      "my-policy",
			},
		}),
		Entry("cluster-scoped k8s resource", testCase{
			namespace: "",
			name:      "my-policy",
			expected: core_model.DimensionalResourceName{
				"k8s.kuma.io/namespace": "",
				"k8s.kuma.io/name":      "my-policy",
			},
		}),
	)
})
