package config_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/config"
)

var _ = Describe("ValidateTokenPath", func() {
	It("should pass validation for empty path", func() {
		// when
		err := config.ValidateTokenPath("")

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should fail for non existing file", func() {
		// when
		err := config.ValidateTokenPath("nonexistingfile")

		// then
		Expect(err).To(MatchError("could not read file nonexistingfile: stat nonexistingfile: no such file or directory"))
	})

	Context("should validate token", func() {
		type testCase struct {
			token         string
			expectedError string
		}
		DescribeTable("should fail with invalid token",
			func(given testCase) {
				// setup
				invalidTokenFile, err := os.CreateTemp("", "")
				Expect(err).ToNot(HaveOccurred())

				_, err = invalidTokenFile.Write([]byte(given.token))
				Expect(err).ToNot(HaveOccurred())

				// when
				err = config.ValidateTokenPath(invalidTokenFile.Name())

				if given.expectedError == "" {
					Expect(err).ToNot(HaveOccurred())
				} else {
					// then
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(strings.ReplaceAll(given.expectedError, "{}", invalidTokenFile.Name())))
				}
			},
			Entry("empty file", testCase{
				token:         "",
				expectedError: "token under file {} is empty",
			}),
			Entry("can't parse token", testCase{
				token:         "yJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoidGVzdCIsIk1lc2giOiJkZWZhdWx0IiwiVGFncyI6e30sIlR5cGUiOiIifQ.rdQ6l_6hzT93Kbk9kO-kZYY7BaexUH8QknvbdRy_f6s",
				expectedError: "not valid JWT token. Can't parse it.: token is malformed: could not JSON decode header: invalid character 'È' looking for beginning of value",
			}),
			Entry("need 3 segments", testCase{
				token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoidGVzdCIsIk1lc2giOiJkZWZhdWx0IiwiVGFncyI6e30sIlR5cGUiOiIifQ",
				expectedError: "not valid JWT token. Can't parse it.: token is malformed: token contains an invalid number of segments",
			}),
			Entry("new line in the end", testCase{
				token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoidGVzdCIsIk1lc2giOiJkZWZhdWx0IiwiVGFncyI6e30sIlR5cGUiOiIifQ.rdQ6l_6hzT93Kbk9kO-kZYY7BaexUH8QknvbdRy_f6s\n",
			}),
			Entry("new line at the start", testCase{
				token: "\neyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoidGVzdCIsIk1lc2giOiJkZWZhdWx0IiwiVGFncyI6e30sIlR5cGUiOiIifQ.rdQ6l_6hzT93Kbk9kO-kZYY7BaexUH8QknvbdRy_f6s\n",
			}),
			Entry("new line in the middle", testCase{
				token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpX\nVCJ9.eyJOYW1lIjoidGVzdCIsIk1lc2giOiJkZWZhdWx0IiwiVGFncyI6e30sIlR5cGUiOiIifQ.rdQ6l_6hzT93Kbk9kO-kZYY7BaexUH8QknvbdRy_f6s\n",
				expectedError: "Token shouldn't contain line breaks within the token, only at the start or end",
			}),
		)
	})
})
