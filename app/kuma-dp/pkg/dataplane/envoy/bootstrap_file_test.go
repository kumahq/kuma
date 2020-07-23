package envoy

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_bootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v2"

	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
)

var _ = Describe("Bootstrap File", func() {
	Describe("GenerateBootstrapFile(..)", func() {

		var configDir string

		BeforeEach(func() {
			var err error
			configDir, err = ioutil.TempDir("", "")
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
			config := &envoy_bootstrap.Bootstrap{
				Node: &envoy_core.Node{
					Id: "example",
				},
			}
			// and
			runtime := kuma_dp.DataplaneRuntime{
				ConfigDir: configDir,
			}

			// when
			filename, err := GenerateBootstrapFile(runtime, config)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(filename).To(Equal(filepath.Join(configDir, "bootstrap.yaml")))

			// when
			actual, err := ioutil.ReadFile(filename)
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
