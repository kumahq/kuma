package multizone

import (
	"crypto/x509"
	"net/url"
	"os"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = &MultizoneConfig{}

// Global configuration
type GlobalConfig struct {
	// KDS Configuration
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
			GrpcPort:                 5685,
			RefreshInterval:          1 * time.Second,
			ZoneInsightFlushInterval: 10 * time.Second,
			MaxMsgSize:               10 * 1024 * 1024,
		},
	}
}

// Zone configuration
type ZoneConfig struct {
	// Kuma Zone name used to mark the zone dataplane resources
	Name string `yaml:"name,omitempty" envconfig:"kuma_multizone_zone_name"`
	// GlobalAddress URL of Global Kuma CP
	GlobalAddress string `yaml:"globalAddress,omitempty" envconfig:"kuma_multizone_zone_global_address"`
	// KDS Configuration
	KDS *KdsClientConfig `yaml:"kds,omitempty"`
}

func (r *ZoneConfig) Sanitize() {
	r.KDS.Sanitize()
}

func (r *ZoneConfig) Validate() error {
	if r.Name == "" {
		return errors.Errorf("Name is mandatory in Zone mode")
	} else if !govalidator.IsDNSName(r.Name) {
		return errors.Errorf("Wrong zone name %s", r.Name)
	}
	if r.GlobalAddress == "" {
		return errors.Errorf("GlobalAddress is mandatory in Zone mode")
	}
	u, err := url.Parse(r.GlobalAddress)
	if err != nil {
		return errors.Wrapf(err, "unable to parse zone GlobalAddress.")
	}
	switch u.Scheme {
	case "grpc":
	case "grpcs":
		rootCaFile := r.KDS.RootCAFile
		if rootCaFile != "" {
			roots := x509.NewCertPool()
			caCert, err := os.ReadFile(rootCaFile)
			if err != nil {
				return errors.Wrapf(err, "could not read certificate %s", rootCaFile)
			}
			ok := roots.AppendCertsFromPEM(caCert)
			if !ok {
				return errors.New("failed to parse root certificate")
			}
		}
	default:
		return errors.Errorf("unsupported scheme %q in zone GlobalAddress. Use one of %s", u.Scheme, []string{"grpc", "grpcs"})
	}
	return r.KDS.Validate()
}

func DefaultZoneConfig() *ZoneConfig {
	return &ZoneConfig{
		GlobalAddress: "",
		Name:          "",
		KDS: &KdsClientConfig{
			RefreshInterval: 1 * time.Second,
			MaxMsgSize:      10 * 1024 * 1024,
		},
	}
}

// Multizone configuration
type MultizoneConfig struct {
	Global *GlobalConfig `yaml:"global,omitempty"`
	Zone   *ZoneConfig   `yaml:"zone,omitempty"`
}

func (m *MultizoneConfig) Sanitize() {
	m.Global.Sanitize()
	m.Zone.Sanitize()
}

func (m *MultizoneConfig) Validate() error {
	panic("not implemented. Call Global and Zone validators as needed.")
}

func DefaultMultizoneConfig() *MultizoneConfig {
	return &MultizoneConfig{
		Global: DefaultGlobalConfig(),
		Zone:   DefaultZoneConfig(),
	}
}
