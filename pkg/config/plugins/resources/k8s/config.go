package k8s

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

func DefaultKubernetesStoreConfig() *KubernetesStoreConfig {
	return &KubernetesStoreConfig{
		SystemNamespace: "kuma-system",
	}
}

// Kubernetes store configuration
type KubernetesStoreConfig struct {
	// Namespace where Control Plane is installed to.
	SystemNamespace string `yaml:"systemNamespace" envconfig:"kuma_store_kubernetes_system_namespace"`
}

var _ config.Config = &KubernetesStoreConfig{}

func (p *KubernetesStoreConfig) Sanitize() {
}

func (p *KubernetesStoreConfig) Validate() error {
	if len(p.SystemNamespace) < 1 {
		return errors.New("SystemNamespace should not be empty")
	}
	return nil
}
