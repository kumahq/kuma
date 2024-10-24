package k8s_test

import (
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	. "github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("Config", func() {
	It("should be loadable from configuration file", func() {
		// given
		cfg := runtime_k8s.KubernetesRuntimeConfig{}
		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.ControlPlaneServiceName).To(Equal("custom-control-plane"))

		Expect(cfg.AdmissionServer.Address).To(Equal("127.0.0.2"))
		Expect(cfg.AdmissionServer.Port).To(Equal(uint32(8442)))
		Expect(cfg.AdmissionServer.CertDir).To(Equal("/var/secret/kuma-cp"))
		// and
		Expect(cfg.Injector.SidecarContainer.Image).To(Equal("kuma-sidecar:latest"))
		Expect(cfg.Injector.SidecarContainer.RedirectPortOutbound).To(Equal(uint32(1234)))
		Expect(cfg.Injector.SidecarContainer.RedirectPortInbound).To(Equal(uint32(1236)))
		Expect(cfg.Injector.SidecarContainer.IpFamilyMode).To(Equal("ipv4"))
		Expect(cfg.Injector.SidecarContainer.UID).To(Equal(int64(2345)))
		Expect(cfg.Injector.SidecarContainer.GID).To(Equal(int64(3456)))
		Expect(cfg.Injector.SidecarContainer.AdminPort).To(Equal(uint32(45678)))
		Expect(cfg.Injector.SidecarContainer.DrainTime.Duration).To(Equal(15 * time.Second))
		// and
		Expect(cfg.Injector.SidecarContainer.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(11)))
		Expect(cfg.Injector.SidecarContainer.ReadinessProbe.TimeoutSeconds).To(Equal(int32(13)))
		Expect(cfg.Injector.SidecarContainer.ReadinessProbe.PeriodSeconds).To(Equal(int32(15)))
		Expect(cfg.Injector.SidecarContainer.ReadinessProbe.SuccessThreshold).To(Equal(int32(11)))
		Expect(cfg.Injector.SidecarContainer.ReadinessProbe.FailureThreshold).To(Equal(int32(112)))
		// and
		Expect(cfg.Injector.SidecarContainer.LivenessProbe.InitialDelaySeconds).To(Equal(int32(260)))
		Expect(cfg.Injector.SidecarContainer.LivenessProbe.TimeoutSeconds).To(Equal(int32(23)))
		Expect(cfg.Injector.SidecarContainer.LivenessProbe.PeriodSeconds).To(Equal(int32(25)))
		Expect(cfg.Injector.SidecarContainer.LivenessProbe.FailureThreshold).To(Equal(int32(212)))
		// and
		Expect(cfg.Injector.SidecarContainer.StartupProbe.InitialDelaySeconds).To(Equal(int32(261)))
		Expect(cfg.Injector.SidecarContainer.StartupProbe.TimeoutSeconds).To(Equal(int32(24)))
		Expect(cfg.Injector.SidecarContainer.StartupProbe.PeriodSeconds).To(Equal(int32(26)))
		Expect(cfg.Injector.SidecarContainer.StartupProbe.FailureThreshold).To(Equal(int32(213)))
		// and
		Expect(cfg.Injector.SidecarContainer.Resources.Requests.CPU).To(Equal("150m"))
		Expect(cfg.Injector.SidecarContainer.Resources.Requests.Memory).To(Equal("164Mi"))
		Expect(cfg.Injector.SidecarContainer.Resources.Limits.CPU).To(Equal("1100m"))
		Expect(cfg.Injector.SidecarContainer.Resources.Limits.Memory).To(Equal("1512Mi"))
		// and
		Expect(cfg.Injector.InitContainer.Image).To(Equal("kuma-init:latest"))
		Expect(cfg.Injector.CNIEnabled).To(BeTrue())
		// and
		Expect(cfg.Injector.BuiltinDNS.Enabled).To(BeTrue())
		Expect(cfg.Injector.BuiltinDNS.Port).To(Equal(uint32(1253)))
		// and
		Expect(cfg.MarshalingCacheExpirationTime.Duration).To(Equal(1 * time.Second))
	})

	It("should have consistent defaults", func() {
		// given
		cfg := runtime_k8s.DefaultKubernetesRuntimeConfig()

		// when
		actual, err := config.ToYAML(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "default-config.golden.yaml")))
	})

	It("should have validators", func() {
		// given
		cfg := runtime_k8s.KubernetesRuntimeConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err.Error()).To(Equal("parsing configuration from file 'testdata/invalid-config.input.yaml' failed: configuration validation failed: .AdmissionServer is not valid: .Port must be in the range [0, 65535]; .CertDir should not be empty; .Injector is not valid: .SidecarContainer is not valid: .Image must be non-empty; .RedirectPortInbound must be in the range [0, 65535]; .RedirectPortOutbound must be in the range [0, 65535]; .AdminPort must be in the range [0, 65535]; .DrainTime must be positive; .ReadinessProbe is not valid: .InitialDelaySeconds must be >= 1; .TimeoutSeconds must be >= 1; .PeriodSeconds must be >= 1; .SuccessThreshold must be >= 1; .FailureThreshold must be >= 1; .LivenessProbe is not valid: .InitialDelaySeconds must be >= 1; .TimeoutSeconds must be >= 1; .PeriodSeconds must be >= 1; .FailureThreshold must be >= 1; .Resources is not valid: .Requests is not valid: .CPU is not valid: quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'; .Memory is not valid: quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'; .Limits is not valid: .CPU is not valid: quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'; .Memory is not valid: quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'; .InitContainer is not valid: .Image must be non-empty; .MarshalingCacheExpirationTime must be positive or equal to 0"))
	})
})
