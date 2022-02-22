package k8s_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	. "github.com/kumahq/kuma/pkg/plugins/common/k8s"
)

var _ = Describe("ResourceNameExtensions()", func() {
	type testCase struct {
		namespace string
		name      string
		expected  core_model.ResourceNameExtensions
	}

	DescribeTable("should return correct k8s name extensions",
		func(given testCase) {
			// when
			actual := ResourceNameExtensions(given.namespace, given.name)
			// then
			Expect(actual).To(HaveKey("k8s.kuma.io/namespace"), "DO NOT change extension identifiers lightly! They are considered a part of user-facing Kuma API")
			Expect(actual).To(HaveKey("k8s.kuma.io/name"), "DO NOT change extension identifiers lightly! They are considered a part of user-facing Kuma API")
			// then
			Expect(actual).To(Equal(given.expected))
		},
		Entry("namespace-scoped k8s resource", testCase{
			namespace: "my-namespace",
			name:      "my-policy",
			expected: core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": "my-namespace",
				"k8s.kuma.io/name":      "my-policy",
			},
		}),
		Entry("cluster-scoped k8s resource", testCase{
			namespace: "",
			name:      "my-policy",
			expected: core_model.ResourceNameExtensions{
				"k8s.kuma.io/namespace": "",
				"k8s.kuma.io/name":      "my-policy",
			},
		}),
	)
})
