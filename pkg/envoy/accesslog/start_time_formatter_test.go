package accesslog_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/envoy/accesslog"

	"github.com/golang/protobuf/ptypes/timestamp"

	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("StartTimeFormatter", func() {

	Describe("FormatHttpLogEntry() and FormatTcpLogEntry()", func() {
		type testCase struct {
			timeFormat       string
			commonProperties *accesslog_data.AccessLogCommon
			expected         string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// setup
				formatter := StartTimeFormatter(given.timeFormat)

				// when
				actual, err := formatter.FormatHttpLogEntry(&accesslog_data.HTTPAccessLogEntry{
					CommonProperties: given.commonProperties,
				})
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))

				// when
				actual, err = formatter.FormatTcpLogEntry(&accesslog_data.TCPAccessLogEntry{
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
			formatter := StartTimeFormatter("")

			// given
			commonProperties := &accesslog_data.AccessLogCommon{
				StartTime: &timestamp.Timestamp{
					Nanos: -1, // is considered invalid
				},
			}

			// when
			_, err := formatter.FormatHttpLogEntry(&accesslog_data.HTTPAccessLogEntry{
				CommonProperties: commonProperties,
			})
			// then
			Expect(err).To(HaveOccurred())

			// when
			_, err = formatter.FormatTcpLogEntry(&accesslog_data.TCPAccessLogEntry{
				CommonProperties: commonProperties,
			})
			// then
			Expect(err).To(HaveOccurred())
		})
	})
})
