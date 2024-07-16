package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewValueOrRangeList", func() {
	DescribeTable("should create ValueOrRangeList",
		func(input interface{}, expected string) {
			// when
			var result ValueOrRangeList

			switch v := input.(type) {
			case []uint16:
				result = NewValueOrRangeList(v)
			case uint16:
				result = NewValueOrRangeList(v)
			case string:
				result = NewValueOrRangeList(v)
			}

			// then
			Expect(string(result)).To(Equal(expected))
		},
		Entry("from uint16 slice", []uint16{80, 443}, "80,443"),
		Entry("from single uint16", uint16(8080), "8080"),
		Entry("from string", "1000-2000", "1000-2000"),
	)
})
