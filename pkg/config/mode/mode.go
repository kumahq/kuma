package mode

import (
	"net"
	"net/url"
	"strconv"

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
	_, err := url.ParseRequestURI(z.Remote.Address)
	if err != nil {
		return errors.Wrapf(err, "Invalid remote address for zone %s", z.Remote)
	}
	_, port, err := net.SplitHostPort(z.Ingress.Address)
	if err != nil {
		return errors.Wrapf(err, "Invalid ingress address for zone %s", z.Ingress)
	}
	_, err = strconv.ParseUint(port, 10, 32)
	if err != nil {
		return errors.Wrapf(err, "Invalid ingress port %s", port)
	}

	return nil
}

// Global configuration
type GlobalConfig struct {
	LBAddress string        `yaml:"lbaddress,omitempty"`
	Zones     []*ZoneConfig `yaml:"zones"`
}

func (g *GlobalConfig) Sanitize() {
}

func (g *GlobalConfig) Validate() error {
	if len(g.Zones) > 0 {
		_, err := url.ParseRequestURI(g.LBAddress)
		if err != nil {
			return errors.Wrapf(err, "Invalid LB address")
		}
	}
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
	Global *GlobalConfig `yaml:"global,omitempty"`
	Remote *RemoteConfig `yaml:"remote,omitempty"`
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
