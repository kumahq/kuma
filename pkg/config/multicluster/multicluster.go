package multicluster

import (
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = &MulticlusterConfig{}

// Global configuration
type GlobalConfig struct {
	KDS *KdsServerConfig `yaml:"kds,omitempty"`
}

func (g *GlobalConfig) Sanitize() {
	g.KDS.Sanitize()
}

func (g *GlobalConfig) Validate() error {
	return g.KDS.Validate()
}

func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		KDS: &KdsServerConfig{
			GrpcPort:        5685,
			RefreshInterval: 1 * time.Second,
		},
	}
}

// Remote configuration
type RemoteConfig struct {
	// Kuma Zone name used to mark the remote dataplane resources
	Zone string `yaml:"zone,omitempty" envconfig:"kuma_multicluster_remote_zone"`
	// GlobalAddress URL of Global Kuma CP
	GlobalAddress string `yaml:"globalAddress,omitempty" envconfig:"kuma_multicluster_remote_global_address"`
	// KDS Configuration
	KDS *KdsClientConfig `yaml:"kds,omitempty"`
}

func (r *RemoteConfig) Sanitize() {
	r.KDS.Sanitize()
}

func (r *RemoteConfig) Validate() error {
	if r.Zone == "" {
		return errors.Errorf("Zone is mandatory in remote mode")
	} else if !govalidator.IsDNSName(r.Zone) {
		return errors.Errorf("Wrong zone name %s", r.Zone)
	}
	return r.KDS.Validate()
}

func DefaultRemoteConfig() *RemoteConfig {
	return &RemoteConfig{
		GlobalAddress: "",
		Zone:          "",
		KDS: &KdsClientConfig{
			RefreshInterval: 1 * time.Second,
		},
	}
}

// Multicluster configuration
type MulticlusterConfig struct {
	Global *GlobalConfig `yaml:"global,omitempty"`
	Remote *RemoteConfig `yaml:"remote,omitempty"`
}

func (m *MulticlusterConfig) Sanitize() {
	if m.Global != nil {
		m.Global.Sanitize()
	}
	if m.Remote != nil {
		m.Remote.Sanitize()
	}
}

func (m *MulticlusterConfig) Validate() error {
	var result error
	if m.Global != nil {
		err := m.Global.Validate()
		if err != nil {
			_ = multierror.Append(result, err)
		}
	}

	if m.Remote != nil {
		err := m.Remote.Validate()
		if err != nil {
			_ = multierror.Append(result, err)
		}
	}

	return result
}

func DefaultMulticlusterConfig() *MulticlusterConfig {
	return &MulticlusterConfig{
		Global: DefaultGlobalConfig(),
		Remote: DefaultRemoteConfig(),
	}
}
