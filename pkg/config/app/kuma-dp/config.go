package kumadp

import (
	"net/url"
	"time"

	"github.com/Kong/kuma/pkg/config"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func DefaultConfig() Config {
	return Config{
		ControlPlane: ControlPlane{
			BootstrapServer: BootstrapServer{
				URL: "", // Address of the Bootstrap Server must be set explicitly
			},
		},
		Dataplane: Dataplane{
			Mesh:      "default",
			Name:      "", // Dataplane name must be set explicitly
			AdminPort: 0,  // by default, turn off Admin interface of Envoy
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

// ControlPlane defines coordinates of the Control Plane.
type ControlPlane struct {
	// BootstrapServer defines coordinates of the Control Plane Bootstrap Server.
	BootstrapServer BootstrapServer `yaml:"bootstrapServer,omitempty"`
}

type BootstrapServer struct {
	// Address defines the address of Bootstrap configuration server.
	URL string `yaml:"url,omitempty" envconfig:"kuma_control_plane_bootstrap_server_url"`
}

// Dataplane defines bootstrap configuration of the dataplane (Envoy).
type Dataplane struct {
	// Mesh name.
	Mesh string `yaml:"mesh,omitempty" envconfig:"kuma_dataplane_mesh"`
	// Dataplane name.
	Name string `yaml:"name,omitempty" envconfig:"kuma_dataplane_name"`
	// Envoy Admin port.
	AdminPort uint32 `yaml:"adminPort,omitempty" envconfig:"kuma_dataplane_admin_port"`
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

func (c *ControlPlane) Validate() (errs error) {
	if err := c.BootstrapServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".BootstrapServer is not valid"))
	}
	return
}

var _ config.Config = &Dataplane{}

func (d *Dataplane) Validate() (errs error) {
	if d.Mesh == "" {
		errs = multierr.Append(errs, errors.Errorf(".Mesh must be non-empty"))
	}
	if d.Name == "" {
		errs = multierr.Append(errs, errors.Errorf(".Name must be non-empty"))
	}
	if 65535 < d.AdminPort {
		errs = multierr.Append(errs, errors.Errorf(".AdminPort must be in the range [0, 65535]"))
	}
	if d.DrainTime <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".DrainTime must be positive"))
	}
	return
}

var _ config.Config = &DataplaneRuntime{}

func (d *DataplaneRuntime) Validate() (errs error) {
	if d.BinaryPath == "" {
		errs = multierr.Append(errs, errors.Errorf(".BinaryPath must be non-empty"))
	}
	return
}

func (d *BootstrapServer) Validate() (errs error) {
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
