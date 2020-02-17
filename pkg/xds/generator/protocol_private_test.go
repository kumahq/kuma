package generator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
)

var _ = Describe("getCommonProtocol()", func() {
	type testCase struct {
		one      mesh_core.Protocol
		another  mesh_core.Protocol
		expected mesh_core.Protocol
	}

	DescribeTable("should correctly determine common protocol",
		func(given testCase) {
			// when
			actual := getCommonProtocol(given.one, given.another)
			// then
			Expect(actual).To(Equal(given.expected))
		},
		Entry("`unknown` and `unknown`", testCase{
			one:      mesh_core.ProtocolUnknown,
			another:  mesh_core.ProtocolUnknown,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`unknown` and `http`", testCase{
			one:      mesh_core.ProtocolUnknown,
			another:  mesh_core.ProtocolHTTP,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`http` and `unknown`", testCase{
			one:      mesh_core.ProtocolHTTP,
			another:  mesh_core.ProtocolUnknown,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`unknown` and `tcp`", testCase{
			one:      mesh_core.ProtocolUnknown,
			another:  mesh_core.ProtocolTCP,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`tcp` and `unknown`", testCase{
			one:      mesh_core.ProtocolTCP,
			another:  mesh_core.ProtocolUnknown,
			expected: mesh_core.ProtocolUnknown,
		}),
		Entry("`http` and `tcp`", testCase{
			one:      mesh_core.ProtocolHTTP,
			another:  mesh_core.ProtocolTCP,
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("`tcp` and `http`", testCase{
			one:      mesh_core.ProtocolTCP,
			another:  mesh_core.ProtocolHTTP,
			expected: mesh_core.ProtocolTCP,
		}),
		Entry("`http` and `http`", testCase{
			one:      mesh_core.ProtocolHTTP,
			another:  mesh_core.ProtocolHTTP,
			expected: mesh_core.ProtocolHTTP,
		}),
		Entry("`tcp` and `tcp`", testCase{
			one:      mesh_core.ProtocolTCP,
			another:  mesh_core.ProtocolTCP,
			expected: mesh_core.ProtocolTCP,
		}),
	)
})
