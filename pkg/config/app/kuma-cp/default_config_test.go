package kuma_cp

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
)

var _ = Describe("Default config", func() {
	It("should be check against the kuma-cp.defaults.yaml file", func() {
		backupEnvVars := os.Environ()
		os.Clearenv()
		defer func() {
			for _, envVar := range backupEnvVars {
				parts := strings.SplitN(envVar, "=", 2)
				Expect(os.Setenv(parts[0], parts[1])).To(Succeed())
			}
		}()
		// given
		cfg := Config{}

		// when
		err := config.Load("kuma-cp.defaults.yaml", &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(DefaultConfig()).To(Equal(cfg), "The default config generated by Kuma and kuma-cp.defaults.yaml config file are different. Please update the kuma-cp.defaults.yaml file.")
	})

	It("kuma-cp.defaults.yaml should not have extra keys", func() {
		cfg := Config{}
		err := config.LoadWithOption("kuma-cp.defaults.yaml", &cfg, true, false, false)
		Expect(err).ToNot(HaveOccurred())
	})
})
