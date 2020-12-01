package multizone

import (
	"crypto/x509"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
)

var _ config.Config = &MultizoneConfig{}

// Global configuration
type GlobalConfig struct {
	PollTimeout time.Duration `yaml:"pollTimeout,omitempty" envconfig:"kuma_multizone_global_poll_timeout"`
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
		PollTimeout: 500 * time.Millisecond,
		KDS: &KdsServerConfig{
			GrpcPort:                 5685,
			RefreshInterval:          1 * time.Second,
			ZoneInsightFlushInterval: 10 * time.Second,
		},
	}
}

// Remote configuration
type RemoteConfig struct {
	// Kuma Zone name used to mark the remote dataplane resources
	Zone string `yaml:"zone,omitempty" envconfig:"kuma_multizone_remote_zone"`
	// GlobalAddress URL of Global Kuma CP
	GlobalAddress string `yaml:"globalAddress,omitempty" envconfig:"kuma_multizone_remote_global_address"`
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
	if r.GlobalAddress == "" {
		return errors.Errorf("GlobalAddress is mandatory in remote mode")
	}
	u, err := url.Parse(r.GlobalAddress)
	if err != nil {
		return errors.Wrapf(err, "unable to parse remote GlobaAddress.")
	}
	switch u.Scheme {
	case "grpc":
	case "grpcs":
		rootCaFile := r.KDS.RootCAFile
		if rootCaFile != "" {
			roots := x509.NewCertPool()
			caCert, err := ioutil.ReadFile(rootCaFile)
			if err != nil {
				return errors.Wrapf(err, "could not read certificate %s", rootCaFile)
			}
			ok := roots.AppendCertsFromPEM(caCert)
			if !ok {
				return errors.New("failed to parse root certificate")
			}
		}
	default:
		return errors.Errorf("unsupported scheme %q in remote GlobalAddress. Use one of %s", u.Scheme, []string{"grpc", "grpcs"})
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

// Multizone configuration
type MultizoneConfig struct {
	Global *GlobalConfig `yaml:"global,omitempty"`
	Remote *RemoteConfig `yaml:"remote,omitempty"`
}

func (m *MultizoneConfig) Sanitize() {
	m.Global.Sanitize()
	m.Remote.Sanitize()
}

func (m *MultizoneConfig) Validate() error {
	panic("not implemented. Call Global and Remote validators as needed.")
}

func DefaultMultizoneConfig() *MultizoneConfig {
	return &MultizoneConfig{
		Global: DefaultGlobalConfig(),
		Remote: DefaultRemoteConfig(),
	}
}
