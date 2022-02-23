package config_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/config"
)

var _ = Describe("ValidateTokenPath", func() {

	var tokenFile *os.File

	BeforeEach(func() {
		tf, err := os.CreateTemp("", "")
		Expect(err).ToNot(HaveOccurred())
		tokenFile = tf
	})

	It("should pass validation for empty path", func() {
		// when
		err := config.ValidateTokenPath("")

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should pass validation for empty path", func() {
		// given
		_, err := tokenFile.WriteString("sampletoken")
		Expect(err).ToNot(HaveOccurred())

		// when
		err = config.ValidateTokenPath("")

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should fail for non existing file", func() {
		// when
		err := config.ValidateTokenPath("nonexistingfile")

		// then
		Expect(err).To(MatchError("could not read file nonexistingfile: stat nonexistingfile: no such file or directory"))
	})

	It("should fail for empty file", func() {
		// when
		err := config.ValidateTokenPath(tokenFile.Name())

		// then
		Expect(err).To(MatchError(fmt.Sprintf("token under file %s is empty", tokenFile.Name())))
	})

	Context("should valicate token", func() {
		type testCase struct {
			token    string
			expected string
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

				// then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(given.expected))
			},
			Entry("can't parse token", testCase{
				token:    "yJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoidGVzdCIsIk1lc2giOiJkZWZhdWx0IiwiVGFncyI6e30sIlR5cGUiOiIifQ.rdQ6l_6hzT93Kbk9kO-kZYY7BaexUH8QknvbdRy_f6s",
				expected: "not valid JWT token. Can't parse it.: invalid character 'È' looking for beginning of value",
			}),
			Entry("need 3 segments", testCase{
				token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoidGVzdCIsIk1lc2giOiJkZWZhdWx0IiwiVGFncyI6e30sIlR5cGUiOiIifQ",
				expected: "not valid JWT token. Can't parse it.: token contains an invalid number of segments",
			}),
			Entry("new line in the end ", testCase{
				token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoidGVzdCIsIk1lc2giOiJkZWZhdWx0IiwiVGFncyI6e30sIlR5cGUiOiIifQ.rdQ6l_6hzT93Kbk9kO-kZYY7BaexUH8QknvbdRy_f6s\n",
				expected: "The file cannot have blank characters like empty lines. Example how to get rid of non-printable characters: sed -i '' '/^$/d' token.file",
			}),
		)

	})
})
