package dnsserver

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
)

var _ = Describe("Corefile", func() {
	Describe("WriteCorefile(..)", func() {
		var configDir string

		BeforeEach(func() {
			var err error
			configDir, err = os.MkdirTemp("", "")
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			if configDir != "" {
				// when
				err := os.RemoveAll(configDir)
				// then
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("should create DNS Server config file on disk", func() {
			// given
			config := `. {
	errors
}`
			// and
			dnsConfig := kuma_dp.DNS{
				ConfigDir: configDir,
			}

			// when
			filename, err := WriteCorefile(dnsConfig, []byte(config))
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(filename).To(Equal(filepath.Join(configDir, "Corefile")))

			// when
			actual, err := os.ReadFile(filename)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(Equal([]byte(`. {
	errors
}`)))
		})
	})
})
