package multizone

import (
	"crypto/x509"
	"net/url"
	"os"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

var _ config.Config = &MultizoneConfig{}

// GlobalConfig defines Global configuration
type GlobalConfig struct {
	// KDS Configuration
	KDS *KdsServerConfig `json:"kds,omitempty"`
}

func (g *GlobalConfig) Sanitize() {
	g.KDS.Sanitize()
}

func (g *GlobalConfig) PostProcess() error {
	return multierr.Combine(g.KDS.PostProcess())
}

func (g *GlobalConfig) Validate() error {
	return g.KDS.Validate()
}

func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		KDS: &KdsServerConfig{
			GrpcPort:                 5685,
			RefreshInterval:          config_types.Duration{Duration: 1 * time.Second},
			ZoneInsightFlushInterval: config_types.Duration{Duration: 10 * time.Second},
			TlsEnabled:               true,
			MaxMsgSize:               10 * 1024 * 1024,
			MsgSendTimeout:           config_types.Duration{Duration: 60 * time.Second},
			TlsMinVersion:            "TLSv1_2",
			TlsCipherSuites:          []string{},
			NackBackoff:              config_types.Duration{Duration: 5 * time.Second},
			DisableSOTW:              false,
		},
	}
}

var _ config.Config = &ZoneConfig{}

// ZoneConfig defines zone configuration
type ZoneConfig struct {
	// Kuma Zone name used to mark the zone dataplane resources
	Name string `json:"name,omitempty" envconfig:"kuma_multizone_zone_name"`
	// GlobalAddress URL of Global Kuma CP
	GlobalAddress string `json:"globalAddress,omitempty" envconfig:"kuma_multizone_zone_global_address"`
	// KDS Configuration
	KDS *KdsClientConfig `json:"kds,omitempty"`
<<<<<<< HEAD
=======
	// DisableOriginLabelValidation disables validation of the origin label when applying resources on Zone CP
	DisableOriginLabelValidation bool `json:"disableOriginLabelValidation,omitempty" envconfig:"kuma_multizone_zone_disable_origin_label_validation"`
	// IngressUpdateInterval is the interval between the CP updating the list of
	// available services on ZoneIngress.
	IngressUpdateInterval config_types.Duration `json:"ingressUpdateInterval,omitempty" envconfig:"kuma_multizone_zone_ingress_update_interval"`
>>>>>>> 08bb1373c (fix(kuma-cp): consistently update ZoneIngress available services (#10426))
}

func (r *ZoneConfig) Sanitize() {
	r.KDS.Sanitize()
}

func (r *ZoneConfig) PostProcess() error {
	return multierr.Combine(r.KDS.PostProcess())
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
			RefreshInterval: config_types.Duration{Duration: 1 * time.Second},
			MaxMsgSize:      10 * 1024 * 1024,
			MsgSendTimeout:  config_types.Duration{Duration: 60 * time.Second},
			NackBackoff:     config_types.Duration{Duration: 5 * time.Second},
		},
<<<<<<< HEAD
=======
		DisableOriginLabelValidation: false,
		IngressUpdateInterval:        config_types.Duration{Duration: 1 * time.Second},
>>>>>>> 08bb1373c (fix(kuma-cp): consistently update ZoneIngress available services (#10426))
	}
}

// MultizoneConfig defines multizone configuration
type MultizoneConfig struct {
	Global *GlobalConfig `json:"global,omitempty"`
	Zone   *ZoneConfig   `json:"zone,omitempty"`
}

func (m *MultizoneConfig) Sanitize() {
	m.Global.Sanitize()
	m.Zone.Sanitize()
}

func (m *MultizoneConfig) PostProcess() error {
	return multierr.Combine(
		m.Global.PostProcess(),
		m.Zone.PostProcess(),
	)
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
