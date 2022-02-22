package table_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
)

var _ = Describe("HumanDuration", func() {
	type testCase struct {
		input  time.Duration
		output string
	}
	DescribeTable("should return the correct human readable duration", func(given testCase) {
		Expect(table.Duration(given.input)).To(Equal(given.output))
	},
		Entry("should return never for invalid time", testCase{
			input:  -2 * time.Second,
			output: "never",
		}),
		Entry("should return 0s if seconds is 0", testCase{
			input:  0,
			output: "0s",
		}),
		Entry("should return second", testCase{
			input:  time.Second,
			output: "1s",
		}),
		Entry("should return seconds if duration is less than a minute", testCase{
			input:  time.Minute - time.Millisecond,
			output: "59s",
		}),
		Entry("should return minute if duration is minute", testCase{
			input:  time.Minute,
			output: "1m",
		}),
		Entry("should return minutes if duration is less than a milliseconds for hour", testCase{
			input:  time.Hour - time.Millisecond,
			output: "59m",
		}),
		Entry("should return the correct year if duration is less than a seconds for hour", testCase{
			input:  3*time.Hour - time.Millisecond,
			output: "2h",
		}),
		Entry("should return the day", testCase{
			input:  24 * time.Hour,
			output: "1d",
		}),
		Entry("should return the year", testCase{
			input:  365 * 24 * time.Hour,
			output: "1y",
		}),
		Entry("should return the year if duration is less than a hour for a year", testCase{
			input:  10 * 365 * 24 * time.Hour,
			output: "10y",
		}),
	)
})
