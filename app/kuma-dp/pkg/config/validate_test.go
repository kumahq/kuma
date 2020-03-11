package config_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/app/kuma-dp/pkg/config"
)

var _ = Describe("ValidateTokenPath", func() {

	var tokenFile *os.File

	BeforeEach(func() {
		tf, err := ioutil.TempFile("", "")
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
		Expect(err).To(MatchError("could not read file: stat nonexistingfile: no such file or directory"))
	})

	It("should fail for empty file", func() {
		// when
		err := config.ValidateTokenPath(tokenFile.Name())

		// then
		Expect(err).To(MatchError(fmt.Sprintf("token under file %s is empty", tokenFile.Name())))
	})
})
