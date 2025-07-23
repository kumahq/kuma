package envoy_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
)

var _ = Describe("Bootstrap File", func() {
	Describe("GenerateBootstrapFile(..)", func() {
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

		It("should create Envoy bootstrap file on disk", func() {
			// given
			config := `node:
  id: example
`
			// and
			runtime := kuma_dp.DataplaneRuntime{
				WorkDir: configDir,
			}

			// when
			filename, err := envoy.GenerateBootstrapFile(runtime, []byte(config))
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(filename).To(Equal(filepath.Join(configDir, "bootstrap.yaml")))

			// when
			actual, err := os.ReadFile(filename)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(`
            node:
              id: example
`))
		})
	})
})
