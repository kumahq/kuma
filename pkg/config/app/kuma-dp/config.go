package kumadp

import (
	"net/url"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

func DefaultConfig() Config {
	return Config{
		ControlPlane: ControlPlane{
			ApiServer: ApiServer{
				URL: "http://localhost:5681",
				Retry: ApiServerRetry{
					Backoff:     3 * time.Second,
					MaxDuration: 5 * time.Minute, // this value can be fairy long since what will happen when there there is a connection error is that the Dataplane will be restarted (by process manager like systemd/K8S etc.) and will try to connect again.
				},
			},
			BootstrapServer: BootstrapServer{
				Retry: BootstrapServerRetry{
					Backoff:     1 * time.Second,
					MaxDuration: 30 * time.Second, // CP should be able to create Dataplen definition in this time, if it hasn't there is probably an error with Pod -> Dataplane conversion.
				},
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
	// BootstrapServer defines settings of the Control Plane Bootstrap Server
	BootstrapServer BootstrapServer `yaml:"bootstrapServer,omitempty"`
	// CaCert defines Certificate Authority that will be used to verify connection to the Control Plane. It takes precedence over CaCertFile.
	CaCert string `yaml:"caCert" envconfig:"kuma_control_plane_ca_cert"`
	// CaCertFile defines a file for Certificate Authority that will be used to verifiy connection to the Control Plane.
	CaCertFile string `yaml:"caCertFile" envconfig:"kuma_control_plane_ca_cert_file"`
}

type ApiServer struct {
	// Address defines the address of Control Plane API server.
	URL string `yaml:"url,omitempty" envconfig:"kuma_control_plane_api_server_url"`
	// Retry settings for API Server
	Retry ApiServerRetry `yaml:"retry,omitempty"`
}

type ApiServerRetry struct {
	// Duration to wait between retries
	Backoff time.Duration `yaml:"backoff,omitempty" envconfig:"kuma_control_plane_api_server_retry_backoff"`
	// Max duration for retries (this is not exact time for execution, the check is done between retries)
	MaxDuration time.Duration `yaml:"maxDuration,omitempty" envconfig:"kuma_control_plane_api_server_retry_max_duration"`
}

func (a *ApiServerRetry) Sanitize() {
}

func (a *ApiServerRetry) Validate() error {
	if a.Backoff <= 0 {
		return errors.New(".Backoff must be a positive duration")
	}
	if a.MaxDuration <= 0 {
		return errors.New(".MaxDuration must be a positive duration")
	}
	return nil
}

var _ config.Config = &ApiServerRetry{}

type BootstrapServer struct {
	Retry BootstrapServerRetry `yaml:"retry,omitempty"`
}

func (b *BootstrapServer) Sanitize() {
}

func (b *BootstrapServer) Validate() error {
	if err := b.Retry.Validate(); err != nil {
		return errors.Wrap(err, ".Retry is not valid")
	}
	return nil
}

var _ config.Config = &BootstrapServer{}

type BootstrapServerRetry struct {
	// Duration to wait between retries
	Backoff time.Duration `yaml:"backoff,omitempty" envconfig:"kuma_control_plane_bootstrap_server_retry_backoff"`
	// Max duration for retries (this is not exact time for execution, the check is done between retries)
	MaxDuration time.Duration `yaml:"maxDuration,omitempty" envconfig:"kuma_control_plane_bootstrap_server_retry_max_duration"`
}

func (b *BootstrapServerRetry) Sanitize() {
}

func (b *BootstrapServerRetry) Validate() error {
	if b.Backoff <= 0 {
		return errors.New(".Backoff must be a positive duration")
	}
	if b.MaxDuration <= 0 {
		return errors.New(".MaxDuration must be a positive duration")
	}
	return nil
}

var _ config.Config = &BootstrapServerRetry{}

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
	// Token is dataplane token's value provided directly, will be stored to a temporary file before applying
	Token string `yaml:"dataplaneToken,omitempty" envconfig:"kuma_dataplane_runtime_token"`
	// Resource is a Dataplane resource that will be applied on Kuma CP
	Resource string `yaml:"resource,omitempty" envconfig:"kuma_dataplane_runtime_resource"`
	// ResourcePath is a path to Dataplane resource that will be applied on Kuma CP
	ResourcePath string `yaml:"resourcePath,omitempty" envconfig:"kuma_dataplane_runtime_resource_path"`
	// ResourceVars are the StringToString values that can fill the Resource template
	ResourceVars map[string]string `yaml:"resourceVars,omitempty"`
}

var _ config.Config = &Config{}

func (c *Config) Validate() (errs error) {
	if err := c.ControlPlane.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ControlPlane is not valid"))
	}
	if c.DataplaneRuntime.Resource != "" || c.DataplaneRuntime.ResourcePath != "" {
		if err := c.Dataplane.ValidateForTemplate(); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, ".Dataplane is not valid"))
		}
	} else {
		if err := c.Dataplane.Validate(); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, ".Dataplane is not valid"))
		}
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
	if err := c.BootstrapServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".BootstrapServer is not valid"))
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

func (d *Dataplane) ValidateForTemplate() (errs error) {
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
	if err := d.Retry.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrap(err, ".Retry is not valid"))
	}
	return
}
