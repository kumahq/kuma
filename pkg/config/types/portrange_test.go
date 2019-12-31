package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/config/types"
)

var _ = Describe("ParsePortRange()", func() {
	Describe("happy paths", func() {
		type testCase struct {
			input    string
			expected PortRange
		}

		DescribeTable("should parse valid text representations",
			func(given testCase) {
				// when
				actual, err := ParsePortRange(given.input)
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(*actual).To(Equal(given.expected))
			},
			Entry("", testCase{
				input:    ``,
				expected: MustPortRange(0, 0),
			}),
			Entry("-", testCase{
				input:    `-`,
				expected: MustPortRange(0, 0),
			}),
			Entry("8080", testCase{
				input:    `8080`,
				expected: MustPortRange(8080, 8080),
			}),
			Entry("8080-8081", testCase{
				input:    `8080-8081`,
				expected: MustPortRange(8080, 8081),
			}),
			Entry("-8080", testCase{
				input:    `-8080`,
				expected: MustPortRange(1, 8080),
			}),
			Entry("8080-", testCase{
				input:    `8080-`,
				expected: MustPortRange(8080, 65535),
			}),
		)
	})

	Describe("error paths", func() {
		type testCase struct {
			input       string
			expectedErr string
		}

		DescribeTable("should fail to parse invalid text representations",
			func(given testCase) {
				// when
				actual, err := ParsePortRange(given.input)
				// then
				Expect(actual).To(BeNil())
				// and
				Expect(err.Error()).To(ContainSubstring(given.expectedErr))
			},
			Entry("0", testCase{
				input:       `0`,
				expectedErr: `invalid value "0". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("0-", testCase{
				input:       `0-`,
				expectedErr: `invalid value "0-". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("-0", testCase{
				input:       `-0`,
				expectedErr: `invalid value "-0". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("0-8080", testCase{
				input:       `0-8080`,
				expectedErr: `invalid value "0-8080". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("8080-0", testCase{
				input:       `8080-0`,
				expectedErr: `invalid value "8080-0". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("82345", testCase{
				input:       `82345`,
				expectedErr: `invalid value "82345". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("-1-2", testCase{
				input:       `-1-2`,
				expectedErr: `invalid value "-1-2". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("1-2-", testCase{
				input:       `1-2-`,
				expectedErr: `invalid value "1-2-". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("1and2", testCase{
				input:       `1and2`,
				expectedErr: `invalid value "1and2". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("1one-2", testCase{
				input:       `1one-2`,
				expectedErr: `invalid value "1one-2". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("1-2two", testCase{
				input:       `1-2two`,
				expectedErr: `invalid value "1-2two". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
			Entry("8081-8080", testCase{
				input:       `8081-8080`,
				expectedErr: `invalid value "8081-8080". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
			}),
		)
	})
})

var _ = Describe("MustExactPort()", func() {
	Describe("happy paths", func() {
		type testCase struct {
			port uint32
		}

		DescribeTable("should create a range that consists of a single port",
			func(given testCase) {
				// when
				actual := MustExactPort(given.port)
				// then
				Expect(actual.Empty()).To(Equal(given.port == 0))
				Expect(actual.Lowest()).To(Equal(given.port))
				Expect(actual.Highest()).To(Equal(given.port))
			},
			Entry("0", testCase{
				port: 0,
			}),
			Entry("1", testCase{
				port: 1,
			}),
			Entry("8080", testCase{
				port: 8080,
			}),
		)
	})

	Describe("error paths", func() {
		type testCase struct {
			port uint32
		}

		DescribeTable("should fail to create a range of invalid port",
			func(given testCase) {
				Expect(func() { MustExactPort(given.port) }).To(Panic())
			},
			Entry("78901", testCase{
				port: 78901,
			}),
		)
	})
})

var _ = Describe("MustPortRange()", func() {
	Describe("happy paths", func() {
		type testCase struct {
			lowest  uint32
			highest uint32
		}

		DescribeTable("should create a valid range",
			func(given testCase) {
				// when
				actual := MustPortRange(given.lowest, given.highest)
				// then
				Expect(actual.Empty()).To(Equal(given.lowest == 0 && given.highest == 0))
				Expect(actual.Lowest()).To(Equal(given.lowest))
				Expect(actual.Highest()).To(Equal(given.highest))
			},
			Entry("0, 0", testCase{
				lowest:  0,
				highest: 0,
			}),
			Entry("8080, 8080", testCase{
				lowest:  8080,
				highest: 8080,
			}),
			Entry("8080, 8081", testCase{
				lowest:  8080,
				highest: 8081,
			}),
		)
	})

	Describe("error paths", func() {
		type testCase struct {
			lowest  uint32
			highest uint32
		}

		DescribeTable("should fail to create an invalid range",
			func(given testCase) {
				Expect(func() { MustPortRange(given.lowest, given.highest) }).To(Panic())
			},
			Entry("8081, 8080", testCase{
				lowest:  8081,
				highest: 8080,
			}),
			Entry("1, 70000", testCase{
				lowest:  1,
				highest: 70000,
			}),
		)
	})
})

var _ = Describe("PortRange", func() {
	Describe("String()", func() {
		type testCase struct {
			lowest   uint32
			highest  uint32
			expected string
		}

		DescribeTable("should format port range properly",
			func(given testCase) {
				// given
				r := MustPortRange(given.lowest, given.highest)
				// when
				actual := r.String()
				// then
				Expect(actual).To(Equal(given.expected))
			},
			Entry("0, 0", testCase{
				lowest:   0,
				highest:  0,
				expected: ``,
			}),
			Entry("8080, 8080", testCase{
				lowest:   8080,
				highest:  8080,
				expected: `8080`,
			}),
			Entry("8080, 8081", testCase{
				lowest:   8080,
				highest:  8081,
				expected: `8080-8081`,
			}),
		)
	})

	Describe("UnmarshalText()", func() {
		Describe("happy paths", func() {
			type testCase struct {
				input    string
				expected PortRange
			}

			DescribeTable("should unmarshal a valid range",
				func(given testCase) {
					// given
					r := PortRange{}
					// when
					err := r.UnmarshalText([]byte(given.input))
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(r).To(Equal(given.expected))
				},
				Entry("", testCase{
					input:    ``,
					expected: PortRange{},
				}),
				Entry("8080", testCase{
					input:    `8080`,
					expected: MustExactPort(8080),
				}),
				Entry("8080-8081", testCase{
					input:    `8080-8081`,
					expected: MustPortRange(8080, 8081),
				}),
			)
		})

		Describe("error paths", func() {
			type testCase struct {
				input       string
				expectedErr string
			}

			DescribeTable("should unmarshal a valid range",
				func(given testCase) {
					// given
					r := PortRange{}
					// when
					err := r.UnmarshalText([]byte(given.input))
					// then
					Expect(err.Error()).To(ContainSubstring(given.expectedErr))
				},
				Entry("0", testCase{
					input:       `0`,
					expectedErr: `invalid value "0". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
				}),
				Entry("8081-8080", testCase{
					input:       `8081-8080`,
					expectedErr: `invalid value "8081-8080". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
				}),
				Entry("78901", testCase{
					input:       `78901`,
					expectedErr: `invalid value "78901". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
				}),
			)
		})
	})

	Describe("Set()", func() {
		Describe("happy paths", func() {
			type testCase struct {
				input    string
				expected PortRange
			}

			DescribeTable("should unmarshal a valid range",
				func(given testCase) {
					// given
					r := PortRange{}
					// when
					err := r.Set(given.input)
					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(r).To(Equal(given.expected))
				},
				Entry("", testCase{
					input:    ``,
					expected: PortRange{},
				}),
				Entry("8080", testCase{
					input:    `8080`,
					expected: MustExactPort(8080),
				}),
				Entry("8080-8081", testCase{
					input:    `8080-8081`,
					expected: MustPortRange(8080, 8081),
				}),
			)
		})

		Describe("error paths", func() {
			type testCase struct {
				input       string
				expectedErr string
			}

			DescribeTable("should unmarshal a valid range",
				func(given testCase) {
					// given
					r := PortRange{}
					// when
					err := r.Set(given.input)
					// then
					Expect(err.Error()).To(ContainSubstring(given.expectedErr))
				},
				Entry("0", testCase{
					input:       `0`,
					expectedErr: `invalid value "0". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
				}),
				Entry("8081-8080", testCase{
					input:       `8081-8080`,
					expectedErr: `invalid value "8081-8080". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
				}),
				Entry("78901", testCase{
					input:       `78901`,
					expectedErr: `invalid value "78901". Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`,
				}),
			)
		})
	})

	Describe("Type()", func() {
		type testCase struct {
			input PortRange
		}

		DescribeTable("should always return `portOrRange`",
			func(given testCase) {
				Expect(given.input.Type()).To(Equal(`portOrRange`))
			},
			Entry("", testCase{
				input: PortRange{},
			}),
			Entry("8080", testCase{
				input: MustExactPort(8080),
			}),
			Entry("8080-8081", testCase{
				input: MustPortRange(8080, 8081),
			}),
		)
	})
})
