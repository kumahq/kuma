package k8s

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
)

func DefaultKubernetesRuntimeConfig() *KubernetesRuntimeConfig {
	return &KubernetesRuntimeConfig{
		AdmissionServer: AdmissionServerConfig{
			Address: "", // all addresses
			Port:    5443,
		},
	}
}

// Kubernetes-specific configuration
type KubernetesRuntimeConfig struct {
	// Admission WebHook Server implemented by the Control Plane.
	AdmissionServer AdmissionServerConfig `yaml:"admissionServer"`
}

// Configuration of the Admission WebHook Server implemented by the Control Plane.
type AdmissionServerConfig struct {
	// Address the Admission WebHook Server should be listening on.
	Address string `yaml:"address" envconfig:"kuma_kubernetes_admission_server_address"`
	// Port the Admission WebHook Server should be listening on.
	Port uint32 `yaml:"port" envconfig:"kuma_kubernetes_admission_server_port"`
	// Directory with a TLS cert and private key for the Admission WebHook Server.
	// TLS certificate file must be named `tls.crt`.
	// TLS key file must be named `tls.key`.
	CertDir string `yaml:"certDir" envconfig:"kuma_kubernetes_admission_server_cert_dir"`
}

var _ config.Config = &KubernetesRuntimeConfig{}

func (c *KubernetesRuntimeConfig) Validate() error {
	if err := c.AdmissionServer.Validate(); err != nil {
		return errors.Wrap(err, "Admission Server validation failed")
	}
	return nil
}

var _ config.Config = &AdmissionServerConfig{}

func (c *AdmissionServerConfig) Validate() error {
	if 65535 < c.Port {
		return errors.New("Port must be in the range [0, 65535]")
	}
	if c.CertDir == "" {
		return errors.New("CertDir should not be empty")
	}
	return nil
}
