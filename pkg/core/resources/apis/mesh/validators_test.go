package mesh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
)

var _ = Describe("AllowedValuesHint()", func() {

	type testCase struct {
		values   []string
		expected string
	}

	DescribeTable("should generate a proper hint",
		func(given testCase) {
			Expect(AllowedValuesHint(given.values...)).To(Equal(given.expected))
		},
		Entry("nil list", testCase{
			values:   nil,
			expected: `Allowed values: (none)`,
		}),
		Entry("empty list", testCase{
			values:   []string{},
			expected: `Allowed values: (none)`,
		}),
		Entry("one-item list", testCase{
			values:   []string{"http"},
			expected: `Allowed values: http`,
		}),
		Entry("multi-item list", testCase{
			values:   []string{"grpc", "http", "http2", "mongo", "mysql", "redis", "tcp"},
			expected: `Allowed values: grpc, http, http2, mongo, mysql, redis, tcp`,
		}),
	)
})
