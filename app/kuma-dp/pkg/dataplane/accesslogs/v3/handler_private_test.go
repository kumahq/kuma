package v3

import (
	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	accesslog "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("handler", func() {

	Describe("Handle()", func() {
		Describe("happy path", func() {

			sampleFormat := `[%START_TIME%] "%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_RECEIVED% %BYTES_SENT% %DURATION% %RESP(CONTENT-TYPE)%
` // intentional newline at the end

			type testCase struct {
				format   string
				msg      string
				expected []string
			}

			DescribeTable("should handle valid messages",
				func(given testCase) {
					By("doing setup")
					fakeSender := fakeSender{}
					// when
					format, err := accesslog.ParseFormat(given.format)
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					handler := &handler{format: format, sender: &fakeSender}

					// when
					msg := &envoy_accesslog.StreamAccessLogsMessage{}
					// and
					err = util_proto.FromYAML([]byte(given.msg), msg)
					// then
					Expect(err).ToNot(HaveOccurred())

					By("doing test")
					// when
					err = handler.Handle(msg)
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect([]string(fakeSender)).To(Equal(given.expected))
				},
				Entry("empty HTTP log entry", testCase{
					format: sampleFormat,
					msg: `
                    http_logs:
                      log_entry:
                      - {}
`,
					expected: []string{"[-] \"- - -\" 0 - 0 0 - -\n"},
				}),
				Entry("1 HTTP log entry", testCase{
					format: sampleFormat,
					msg: `
                    http_logs:
                      log_entry:
                      - common_properties:
                          start_time: 2020-02-11T12:34:56.123Z
                        protocol_version: HTTP11
                        request:
                          request_method: POST
                          authority: backend.internal:8080
                          path: /api
                          request_body_bytes: 234
                        response:
                          response_code: 200
                          response_headers:
                            content-type: application/json
                          response_body_bytes: 567
`,
					expected: []string{"[2020-02-11T12:34:56.123Z] \"POST /api HTTP/1.1\" 200 - 234 567 - application/json\n"},
				}),
				Entry("2 HTTP log entries", testCase{
					format: sampleFormat,
					msg: `
                    http_logs:
                      log_entry:
                      - common_properties:
                          start_time: 2020-02-11T12:34:56.123Z
                        protocol_version: HTTP11
                        request:
                          request_method: POST
                          authority: backend.internal:8080
                          path: /api
                          request_body_bytes: 234
                        response:
                          response_code: 200
                          response_headers:
                            content-type: application/json
                          response_body_bytes: 567
                      - common_properties:
                          start_time: 2020-02-18T23:45:07.456Z
                        protocol_version: HTTP2
                        request:
                          request_method: GET
                          authority: web.internal:8080
                          path: /index.html
                          request_body_bytes: 0
                        response:
                          response_code: 301
                          response_headers:
                            content-type: text/html
                          response_body_bytes: 89012
`,
					expected: []string{
						"[2020-02-11T12:34:56.123Z] \"POST /api HTTP/1.1\" 200 - 234 567 - application/json\n",
						"[2020-02-18T23:45:07.456Z] \"GET /index.html HTTP/2\" 301 - 0 89012 - text/html\n",
					},
				}),
				Entry("empty TCP log entry", testCase{
					format: sampleFormat,
					msg: `
                    tcp_logs:
                      log_entry:
                      - {}
`,
					expected: []string{"[-] \"- - -\" 0 - 0 0 - -\n"},
				}),
				Entry("1 HTTP log entry", testCase{
					format: sampleFormat,
					msg: `
                    tcp_logs:
                      log_entry:
                      - common_properties:
                          start_time: 2020-02-11T12:34:56.123Z
                        connection_properties:
                          received_bytes: 234
                          sent_bytes: 567
`,
					expected: []string{"[2020-02-11T12:34:56.123Z] \"- - -\" 0 - 234 567 - -\n"},
				}),
				Entry("2 TCP log entries", testCase{
					format: sampleFormat,
					msg: `
                    tcp_logs:
                      log_entry:
                      - common_properties:
                          start_time: 2020-02-11T12:34:56.123Z
                        connection_properties:
                          received_bytes: 234
                          sent_bytes: 567
                      - common_properties:
                          start_time: 2020-02-18T23:45:07.456Z
                        connection_properties:
                          received_bytes: 0
                          sent_bytes: 89012
`,
					expected: []string{
						"[2020-02-11T12:34:56.123Z] \"- - -\" 0 - 234 567 - -\n",
						"[2020-02-18T23:45:07.456Z] \"- - -\" 0 - 0 89012 - -\n",
					},
				}),
			)
		})

		Describe("error path", func() {
			type testCase struct {
				format      string
				msg         string
				expectedErr string
			}

			DescribeTable("should reject invalid mesasages",
				func(given testCase) {
					By("doing setup")
					fakeSender := fakeSender{}
					// when
					format, err := accesslog.ParseFormat(given.format)
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					handler := &handler{format: format, sender: &fakeSender}

					// when
					msg := &envoy_accesslog.StreamAccessLogsMessage{}
					// and
					err = util_proto.FromYAML([]byte(given.msg), msg)
					// then
					Expect(err).ToNot(HaveOccurred())

					By("doing test")
					// when
					err = handler.Handle(msg)
					// then
					Expect(err).To(HaveOccurred())
					// and
					Expect(err.Error()).To(Equal(given.expectedErr))
				},
				Entry("no log entries", testCase{
					expectedErr: `unknown type of log entries: <nil>`,
				}),
			)
		})
	})
})
