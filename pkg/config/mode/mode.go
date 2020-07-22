package mode

import (
	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = &ModeConfig{}

// Control Plane mode

type CpMode = string

const (
	Standalone CpMode = "standalone"
	Remote     CpMode = "remote"
	Global     CpMode = "global"
)

// ValidateCpMode to check modes of kuma-cp
func ValidateCpMode(mode CpMode) error {
	if mode != Standalone && mode != Remote && mode != Global {
		return errors.Errorf("invalid mode. Available modes: %s, %s, %s", Standalone, Remote, Global)
	}
	return nil
}

// Remote configuration
type RemoteConfig struct {
	// Kuma Zone name used to mark the remote dataplane resources
	Zone string `yaml:"zone,omitempty" envconfig:"kuma_mode_remote_zone"`
}

func (r *RemoteConfig) Sanitize() {
}

func (r *RemoteConfig) Validate() error {
	if r.Zone == "" {
		return errors.Errorf("Zone is mandatory in remote mode")
	} else if !govalidator.IsDNSName(r.Zone) {
		return errors.Errorf("Wrong zone name %s", r.Zone)
	}
	return nil
}

func DefaultRemoteConfig() *RemoteConfig {
	return &RemoteConfig{
		Zone: "",
	}
}

// Mode configuration
type ModeConfig struct {
	Mode   CpMode        `yaml:"mode" envconfig:"kuma_mode_mode"`
	Remote *RemoteConfig `yaml:"remote,omitempty"`
}

func (m *ModeConfig) Sanitize() {
}

func (m *ModeConfig) Validate() error {
	switch m.Mode {
	case Standalone:
	case Global:
	case Remote:
		return m.Remote.Validate()
	default:
		return errors.Errorf("Unsupported mode %s", m.Mode)
	}
	return nil
}

func DefaultModeConfig() *ModeConfig {
	return &ModeConfig{
		Mode:   Standalone,
		Remote: DefaultRemoteConfig(),
	}
}
