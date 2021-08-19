package runtime

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/config/plugins/runtime/universal"
)

func DefaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		Kubernetes: k8s.DefaultKubernetesRuntimeConfig(),
		Universal:  universal.DefaultUniversalRuntimeConfig(),
	}
}

// Environment-specific configuration
type RuntimeConfig struct {
	// Kubernetes-specific configuration
	Kubernetes *k8s.KubernetesRuntimeConfig `yaml:"kubernetes"`
	// Universal-specific configuration
	Universal *universal.UniversalRuntimeConfig `yaml:"universal"`
}

func (c *RuntimeConfig) Sanitize() {
	c.Kubernetes.Sanitize()
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
