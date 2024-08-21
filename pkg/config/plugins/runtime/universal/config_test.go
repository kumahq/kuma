package universal_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/plugins/runtime/universal"
)

var _ = Describe("Config", func() {
	It("should be loadable from configuration file", func() {
		// given
		cfg := universal.UniversalRuntimeConfig{}
		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.DataplaneCleanupAge.Duration).To(Equal(5 * time.Hour))
		Expect(cfg.VIPRefreshInterval.Duration).To(Equal(5 * time.Second))
	})

	It("should have consistent defaults", func() {
		// given
		cfg := universal.DefaultUniversalRuntimeConfig()

		// when
		actual, err := config.ToYAML(cfg)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		expected, err := os.ReadFile(filepath.Join("testdata", "default-config.golden.yaml"))
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(expected))
	})
	//
	It("should have validators", func() {
		// given
		cfg := universal.UniversalRuntimeConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("parsing configuration from file 'testdata/invalid-config.input.yaml' failed: configuration validation failed: .DataplaneCleanupAge must be positive; .VIPRefreshInterval must be positive"))
	})
})
