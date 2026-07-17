package kuma_cp_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v3/pkg/config/core"
)

var _ = Describe("Config Validate", func() {
	It("should reject global mode with kubernetes environment", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.Mode = core.Global
		cfg.Environment = core.KubernetesEnvironment

		// when
		err := cfg.Validate()

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Kubernetes-native Global Control Plane is not supported"))
		Expect(err.Error()).To(ContainSubstring("mode=global"))
		Expect(err.Error()).To(ContainSubstring("environment=kubernetes"))
		Expect(err.Error()).To(ContainSubstring("environment=universal"))
	})

	It("should allow global mode with universal environment", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.Mode = core.Global
		cfg.Environment = core.UniversalEnvironment

		// when
		err := cfg.Validate()

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reject global mode with an invalid environment", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		cfg.Mode = core.Global
		cfg.Environment = "not-a-real-environment"

		// when
		err := cfg.Validate()

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Environment should be either"))
	})
})
