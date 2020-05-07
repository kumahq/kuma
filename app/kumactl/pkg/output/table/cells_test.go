package table_test

import (
	"time"

	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/Kong/kuma/app/kumactl/pkg/output/table"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HumanDuration", func() {
	type testCase struct {
		input  time.Duration
		output string
	}
	entries := []TableEntry{
		Entry("For Duration", testCase{
			input:  -2 * time.Second,
			output: "never",
		}),
		Entry("For Duration", testCase{
			input:  0,
			output: "0s",
		}),
		Entry("For Duration", testCase{
			input:  time.Second,
			output: "1s",
		}),
		Entry("For Duration", testCase{
			input:  time.Minute - time.Millisecond,
			output: "59s",
		}),
		Entry("For Duration", testCase{
			input:  time.Minute,
			output: "1m",
		}),
		Entry("For Duration", testCase{
			input:  time.Hour - time.Millisecond,
			output: "59m",
		}),
		Entry("For Duration", testCase{
			input:  3*time.Hour - time.Millisecond,
			output: "2h",
		}),
		Entry("For Duration", testCase{
			input:  24 * time.Hour,
			output: "1d",
		}),
		Entry("For Duration", testCase{
			input:  365 * 24 * time.Hour,
			output: "1y",
		}),
		Entry("For Duration", testCase{
			input:  10 * 365 * 24 * time.Hour,
			output: "10y",
		}),
	}
	DescribeTable("should return the correct human readable duration", func(given testCase) {
		Expect(table.Duration(given.input)).To(Equal(given.output))
	},
		entries...,
	)
})
