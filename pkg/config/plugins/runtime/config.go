package runtime

import (
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/v2/pkg/config/plugins/runtime/universal"
)

func DefaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		OtelEnvEnabled: false,
		Kubernetes:     k8s.DefaultKubernetesRuntimeConfig(),
		Universal:      universal.DefaultUniversalRuntimeConfig(),
	}
}

// RuntimeConfig defines Environment-specific configuration
type RuntimeConfig struct {
	// OtelEnvEnabled is the control-plane authoritative global guard for OTEL
	// exporter env-var support. When false, CP plans OTEL env reuse as disabled
	// even if a dataplane advertises the capability.
	OtelEnvEnabled bool `json:"otelEnvEnabled" envconfig:"kuma_runtime_otel_env_enabled"`
	// Kubernetes-specific configuration
	Kubernetes *k8s.KubernetesRuntimeConfig `json:"kubernetes"`
	// Universal-specific configuration
	Universal *universal.UniversalRuntimeConfig `json:"universal"`
}

func (c *RuntimeConfig) Sanitize() {
	c.Kubernetes.Sanitize()
}

func (c *RuntimeConfig) PostProcess() error {
	return multierr.Combine(
		c.Kubernetes.PostProcess(),
		c.Universal.PostProcess(),
	)
}

func (c *RuntimeConfig) Validate(env core.EnvironmentType) error {
	switch env {
	case core.KubernetesEnvironment:
		if err := c.Kubernetes.Validate(); err != nil {
			return errors.Wrap(err, "Kubernetes validation failed")
		}
	case core.UniversalEnvironment:
	default:
		return errors.Errorf("unknown environment type %q", env)
	}
	return nil
}
