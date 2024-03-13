package kuma_cp_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
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
		cfg := kuma_cp.Config{}

		// when
		err := config.Load("kuma-cp.defaults.yaml", &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).To(Equal(kuma_cp.DefaultConfig()), "The default config generated by Kuma and kuma-cp.defaults.yaml config file are different. Please update the kuma-cp.defaults.yaml file.")
	})

	It("kuma-cp.defaults.yaml should not have extra keys", func() {
		cfg := kuma_cp.Config{}
		err := config.LoadWithOption("kuma-cp.defaults.yaml", &cfg, true, false, false)
		Expect(err).ToNot(HaveOccurred())
	})
})
