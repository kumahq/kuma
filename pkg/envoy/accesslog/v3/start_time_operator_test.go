package v3_test

import (
	"time"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("StartTimeOperator", func() {

	Describe("FormatHttpLogEntry() and FormatTcpLogEntry()", func() {
		type testCase struct {
			timeFormat       string
			commonProperties *accesslog_data.AccessLogCommon
			expected         string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// setup
				fragment := StartTimeOperator(given.timeFormat)

				// when
				actual, err := fragment.FormatHttpLogEntry(&accesslog_data.HTTPAccessLogEntry{
					CommonProperties: given.commonProperties,
				})
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))

				// when
				actual, err = fragment.FormatTcpLogEntry(&accesslog_data.TCPAccessLogEntry{
					CommonProperties: given.commonProperties,
				})
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))
			},
			Entry("start time is `nil`", testCase{
				timeFormat:       "", // default time format
				commonProperties: nil,
				expected:         ``,
			}),
			Entry("start time is valid", testCase{
				timeFormat: "", // default time format
				commonProperties: &accesslog_data.AccessLogCommon{
					StartTime: util_proto.MustTimestampProto(time.Unix(1582062737, 987654321)),
				},
				expected: `2020-02-18T21:52:17.987Z`,
			}),
			Entry("user-defined time format", testCase{
				timeFormat: "%s.%3f", // not supported yet
				commonProperties: &accesslog_data.AccessLogCommon{
					StartTime: util_proto.MustTimestampProto(time.Unix(1582062737, 987654321)),
				},
				expected: `2020-02-18T21:52:17.987Z`,
			}),
		)

		It("should fail if start time is not valid", func() {
			// setup
			fragment := StartTimeOperator("")

			// given
			commonProperties := &accesslog_data.AccessLogCommon{
				StartTime: &timestamppb.Timestamp{
					Nanos: -1, // is considered invalid
				},
			}

			// when
			_, err := fragment.FormatHttpLogEntry(&accesslog_data.HTTPAccessLogEntry{
				CommonProperties: commonProperties,
			})
			// then
			Expect(err).To(HaveOccurred())

			// when
			_, err = fragment.FormatTcpLogEntry(&accesslog_data.TCPAccessLogEntry{
				CommonProperties: commonProperties,
			})
			// then
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("String()", func() {
		type testCase struct {
			timeFormat string
			expected   string
		}

		DescribeTable("should return correct canonical representation",
			func(given testCase) {
				// setup
				fragment := StartTimeOperator(given.timeFormat)

				// when
				actual := fragment.String()
				// then
				Expect(actual).To(Equal(given.expected))

			},
			Entry("%START_TIME%", testCase{
				timeFormat: "", // default time format
				expected:   `%START_TIME%`,
			}),
			Entry("%START_TIME(%Y/%m/%dT%H:%M:%S%z %s)%", testCase{
				timeFormat: "%Y/%m/%dT%H:%M:%S%z %s",
				expected:   `%START_TIME(%Y/%m/%dT%H:%M:%S%z %s)%`,
			}),
			Entry("%START_TIME(%s.%3f)%", testCase{
				timeFormat: "%s.%3f",
				expected:   `%START_TIME(%s.%3f)%`,
			}),
		)
	})
})
