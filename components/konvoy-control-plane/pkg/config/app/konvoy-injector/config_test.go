package konvoyinjector_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	konvoy_injector "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-injector"
)

var _ = Describe("Config", func() {

	It("should be loadable from configuration file", func() {
		// given
		cfg := konvoy_injector.Config{}

		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.WebHookServer.Address).To(Equal("127.0.0.2"))
		Expect(cfg.WebHookServer.Port).To(Equal(uint32(8442)))
		Expect(cfg.WebHookServer.CertDir).To(Equal("/var/secret/konvoy-injector"))
	})

	It("should have consistent defaults", func() {
		// given
		cfg := konvoy_injector.DefaultConfig()

		// when
		actual, err := config.ToYAML(&cfg)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		expected, err := ioutil.ReadFile(filepath.Join("testdata", "default-config.golden.yaml"))
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(expected))
	})

	It("should have validators", func() {
		// given
		cfg := konvoy_injector.Config{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err).To(MatchError(`Invalid configuration: .WebHookServer is not valid: .Address must be either empty or a valid IPv4/IPv6 address; .Port must be in the range [0, 65535]; .CertDir must be non-empty`))
	})
})
