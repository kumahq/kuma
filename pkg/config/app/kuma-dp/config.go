package kumadp

import (
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/Kong/kuma/pkg/config"
	config_types "github.com/Kong/kuma/pkg/config/types"
)

func DefaultConfig() Config {
	return Config{
		ControlPlane: ControlPlane{
			ApiServer: ApiServer{
				URL: "http://localhost:5681",
			},
		},
		Dataplane: Dataplane{
			Mesh:      "default",
			Name:      "",                                                      // Dataplane name must be set explicitly
			AdminPort: config_types.MustPortRange(30001, config_types.MaxPort), // by default, automatically choose a free port for Envoy Admin interface
			DrainTime: 30 * time.Second,
		},
		DataplaneRuntime: DataplaneRuntime{
			BinaryPath: "envoy",
			ConfigDir:  "", // if left empty, a temporary directory will be generated automatically
		},
		SDS: SDS{
			Type: SdsCp,
		},
	}
}

// Config defines configuration of the Kuma Dataplane Manager.
type Config struct {
	// ControlPlane defines coordinates of the Kuma Control Plane.
	ControlPlane ControlPlane `yaml:"controlPlane,omitempty"`
	// Dataplane defines bootstrap configuration of the dataplane (Envoy).
	Dataplane Dataplane `yaml:"dataplane,omitempty"`
	// DataplaneRuntime defines the context in which dataplane (Envoy) runs.
	DataplaneRuntime DataplaneRuntime `yaml:"dataplaneRuntime,omitempty"`
	// SDS defines configuration for local Secret Discovery Service
	SDS SDS `yaml:"sds"`
}

func (c *Config) Sanitize() {
	c.ControlPlane.Sanitize()
	c.Dataplane.Sanitize()
	c.DataplaneRuntime.Sanitize()
}

// ControlPlane defines coordinates of the Control Plane.
type ControlPlane struct {
	// ApiServer defines coordinates of the Control Plane API Server
	ApiServer ApiServer `yaml:"apiServer,omitempty"`
}

type ApiServer struct {
	// Address defines the address of Control Plane API server.
	URL string `yaml:"url,omitempty" envconfig:"kuma_control_plane_api_server_url"`
}

// Dataplane defines bootstrap configuration of the dataplane (Envoy).
type Dataplane struct {
	// Mesh name.
	Mesh string `yaml:"mesh,omitempty" envconfig:"kuma_dataplane_mesh"`
	// Dataplane name.
	Name string `yaml:"name,omitempty" envconfig:"kuma_dataplane_name"`
	// Port (or range of ports to choose from) for Envoy Admin API to listen on.
	// Empty value indicates that Envoy Admin API should not be exposed over TCP.
	// Format: "9901 | 9901-9999 | 9901- | -9901".
	AdminPort config_types.PortRange `yaml:"adminPort,omitempty" envconfig:"kuma_dataplane_admin_port"`
	// Drain time for listeners.
	DrainTime time.Duration `yaml:"drainTime,omitempty" envconfig:"kuma_dataplane_drain_time"`
}

// DataplaneRuntime defines the context in which dataplane (Envoy) runs.
type DataplaneRuntime struct {
	// Path to Envoy binary.
	BinaryPath string `yaml:"binaryPath,omitempty" envconfig:"kuma_dataplane_runtime_binary_path"`
	// Dir to store auto-generated Envoy bootstrap config in.
	ConfigDir string `yaml:"configDir,omitempty" envconfig:"kuma_dataplane_runtime_config_dir"`
	// Path to a file with dataplane token (use 'kumactl generate dataplane-token' to get one)
	TokenPath string `yaml:"dataplaneTokenPath,omitempty" envconfig:"kuma_dataplane_runtime_token_path"`
}

// SDS defines configuration of Secret Discovery Service that DP will use
type SDS struct {
	// Type of SDS connection that DP will use
	Type string `yaml:"type" envconfig:"kuma_sds_type"`
	// Address of the SDS server. Only Unix socket is supported for now ex. "unix:///tmp/server.sock"
	Address string `yaml:"address" envconfig:"kuma_sds_address"`
	// Vault SDS configuration
	Vault Vault `yaml:"vault"`
}

// SdsCP defines that DP will use SDS from CP therefore it needs Dataplane Token Server
const SdsCp string = "cp"

// SdsDpVault defines that DP will use SDS embedded in Kuma DP that connects to Vault
const SdsDpVault string = "dpVault"

func (s SDS) Sanitize() {
	s.Vault.Sanitize()
}

func (s SDS) Validate() error {
	switch s.Type {
	case SdsCp:
	case SdsDpVault:
		if err := s.Vault.Validate(); err != nil {
			return errors.Wrapf(err, ".TLS is not valid")
		}
	default:
		return errors.Errorf(".Type has to be either %s or %s", SdsCp, SdsDpVault)
	}
	if s.Address == "" && strings.HasPrefix(s.Address, "unix:///") {
		return errors.Errorf(".Address should be valit unix socket address")
	}
	return nil
}

var _ config.Config = SDS{}

// Vault defines configuration for Vault connections from DP embedded in Kuma DP
type Vault struct {
	// Address of Vault
	Address string `yaml:"address" envconfig:"kuma_sds_vault_address"`
	// Agent address of Vault
	AgentAddress string `yaml:"agentAddress" envconfig:"kuma_sds_vault_agent_address"`
	// Token used for authentication with Vault
	Token string `yaml:"token" envconfig:"kuma_sds_vault_token"`
	// Vault namespace
	Namespace string `yaml:"namespace" envconfig:"kuma_sds_vault_namespace"`
	// TLS connection configuration
	TLS VaultTLSConfig `yaml:"tls"`
}

func (v Vault) Sanitize() {
	v.TLS.Sanitize()
}

func (v Vault) Validate() (errs error) {
	if v.Address != "" && v.AgentAddress != "" {
		errs = multierr.Append(errs, errors.Errorf("either .Address or .AgentAddress must be non-empty"))
	}
	if v.Token != "" {
		errs = multierr.Append(errs, errors.Errorf(".Token must be non-empty"))
	}
	if err := v.TLS.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".TLS is not valid"))
	}
	return
}

var _ config.Config = &Vault{}

var _ config.Config = &VaultTLSConfig{}

// VaultTLSConfig defines configuration for secured TLS connection to Vault
type VaultTLSConfig struct {
	// Path to TLS certificate that will be used to connect to Vault
	CaCertPath string `yaml:"caCertPath" envconfig:"kuma_sds_vault_tls_ca_cert_path"`
	// Path to directory of TLS certificates that will be used to connect to Vault
	CaCertDir string `yaml:"caCertDir" envconfig:"kuma_sds_vault_tls_ca_cert_dir"`
	// Path to client TLS certificate that will be used to connect to Vault
	ClientCertPath string `yaml:"clientCertPath" envconfig:"kuma_sds_vault_tls_client_cert_path"`
	// Path to client TLS key that will be used to connect to Vault
	ClientKeyPath string `yaml:"clientKeyPath" envconfig:"kuma_sds_vault_tls_client_key_path"`
	// Disables TLS verification
	SkipVerify bool `yaml:"skipVerify" envconfig:"kuma_sds_vault_tls_skip_verify"`
	// If set, it is used to set the SNI host when connecting via TLS
	ServerName string `yaml:"serverName" envconfig:"kuma_sds_vault_tls_server_name"`
}

func (v VaultTLSConfig) Sanitize() {
}

func (v VaultTLSConfig) Validate() (errs error) {
	if v.SkipVerify {
		return nil
	}
	if v.ClientKeyPath != "" && v.ClientCertPath == "" {
		errs = multierr.Append(errs, errors.New(".ClientCertPath cannot be empty when .ClientKeyPath is set"))
	}
	if v.ClientKeyPath == "" && v.ClientCertPath != "" {
		errs = multierr.Append(errs, errors.New(".ClientKeyPath cannot be empty when .ClientCertPath is set"))
	}
	return
}

var _ config.Config = &Config{}

func (c *Config) Validate() (errs error) {
	if err := c.ControlPlane.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ControlPlane is not valid"))
	}
	if err := c.Dataplane.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Dataplane is not valid"))
	}
	if err := c.DataplaneRuntime.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".DataplaneRuntime is not valid"))
	}
	return
}

var _ config.Config = &ControlPlane{}

func (c *ControlPlane) Sanitize() {
	c.ApiServer.Sanitize()
}

func (c *ControlPlane) Validate() (errs error) {
	if err := c.ApiServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ApiServer is not valid"))
	}
	return
}

var _ config.Config = &Dataplane{}

func (d *Dataplane) Sanitize() {
}

func (d *Dataplane) Validate() (errs error) {
	if d.Mesh == "" {
		errs = multierr.Append(errs, errors.Errorf(".Mesh must be non-empty"))
	}
	if d.Name == "" {
		errs = multierr.Append(errs, errors.Errorf(".Name must be non-empty"))
	}
	// Notice that d.AdminPort is always valid by design of PortRange
	if d.DrainTime <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".DrainTime must be positive"))
	}
	return
}

var _ config.Config = &DataplaneRuntime{}

func (d *DataplaneRuntime) Sanitize() {
}

func (d *DataplaneRuntime) Validate() (errs error) {
	if d.BinaryPath == "" {
		errs = multierr.Append(errs, errors.Errorf(".BinaryPath must be non-empty"))
	}
	return
}

var _ config.Config = &ApiServer{}

func (d *ApiServer) Sanitize() {
}

func (d *ApiServer) Validate() (errs error) {
	if d.URL == "" {
		errs = multierr.Append(errs, errors.Errorf(".URL must be non-empty"))
	}
	if url, err := url.Parse(d.URL); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".URL must be a valid absolute URI"))
	} else if !url.IsAbs() {
		errs = multierr.Append(errs, errors.Errorf(".URL must be a valid absolute URI"))
	}
	return
}
