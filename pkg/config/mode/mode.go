package mode

import (
	"net"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config"
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

type EndpointConfig struct {
	Address string `yaml:"address"`
}

type ZoneConfig struct {
	Remote  EndpointConfig `yaml:"remote,omitempty"`
	Ingress EndpointConfig `yaml:"ingress,omitempty"`
}

func (z *ZoneConfig) Sanitize() {
}

func (z *ZoneConfig) Validate() error {
	_, _, err := net.SplitHostPort(z.Remote.Address)
	if err != nil {
		return errors.Wrapf(err, "Invalid remote url for cluster %s", z.Remote)
	}
	_, _, err = net.SplitHostPort(z.Ingress.Address)
	if err != nil {
		return errors.Wrapf(err, "Invalid ingress url for cluster %s", z.Ingress)
	}

	return nil
}

// Global configuration
type GlobalConfig struct {
	Zones     []*ZoneConfig `yaml:"zones"`
	LBAddress string        `yaml:"lbaddress,omitempty"`
}

func (g *GlobalConfig) Sanitize() {
}

func (g *GlobalConfig) Validate() error {
	for _, zone := range g.Zones {
		err := zone.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Zones:     []*ZoneConfig{},
		LBAddress: "",
	}
}

// Remote configuration
type RemoteConfig struct {
	// Kuma Zone name used to mark the remote dataplane resources
	Zone string `yaml:"zone,omitempty"`
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
	Mode   CpMode        `yaml:"mode"`
	Global *GlobalConfig `yaml:"global,omitempty"`
	Remote *RemoteConfig `yaml:"local,omitempty"`
}

func (m *ModeConfig) Sanitize() {
}

func (m *ModeConfig) Validate() error {
	switch m.Mode {
	case Standalone:
	case Global:
		return m.Global.Validate()
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
		Global: DefaultGlobalConfig(),
		Remote: DefaultRemoteConfig(),
	}
}
