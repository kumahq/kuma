package konvoydp

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func DefaultConfig() Config {
	return Config{
		ControlPlane: ControlPlane{
			XdsServer: XdsServer{
				Address: "", // Address of the xDS Server must be set explicitly
				Port:    5678,
			},
		},
		Dataplane: Dataplane{
			Id:        "", // Envoy Id must must be set explicitly
			Service:   "", // Envoy Cluster must must be set explicitly
			AdminPort: 9901,
		},
		DataplaneRuntime: DataplaneRuntime{
			BinaryPath: "envoy",
			ConfigDir:  "/tmp/getkonvoy.io/envoy",
		},
	}
}

// Config defines configuration of the Konvoy Dataplane Manager.
type Config struct {
	// ControlPlane defines coordinates of the Konvoy Control Plane.
	ControlPlane ControlPlane `yaml:"controlPlane,omitempty"`
	// Dataplane defines bootstrap configuration of the dataplane (Envoy).
	Dataplane Dataplane `yaml:"dataplane,omitempty"`
	// DataplaneRuntime defines the context in which dataplane (Envoy) runs.
	DataplaneRuntime DataplaneRuntime `yaml:"dataplaneRuntime,omitempty"`
}

// ControlPlane defines coordinates of the Control Plane.
type ControlPlane struct {
	// XdsServer defines coordinates of the Control Plane xDS Server.
	XdsServer XdsServer `yaml:"xdsServer,omitempty"`
}

// XdsServer defines coordinates of the Control Plane xDS Server.
type XdsServer struct {
	// Address defines the address of xDS server.
	Address string `yaml:"address,omitempty" envconfig:"konvoy_control_plane_xds_server_address"`
	// Port defines the port of xDS server.
	Port uint32 `yaml:"port,omitempty" envconfig:"konvoy_control_plane_xds_server_port"`
}

// Dataplane defines bootstrap configuration of the dataplane (Envoy).
type Dataplane struct {
	// Envoy node Id.
	Id string `yaml:"id,omitempty" envconfig:"konvoy_dataplane_id"`
	// Envoy node cluster.
	Service string `yaml:"service,omitempty" envconfig:"konvoy_dataplane_service"`
	// Envoy Admin port.
	AdminPort uint32 `yaml:"adminPort,omitempty" envconfig:"konvoy_dataplane_admin_port"`
}

// DataplaneRuntime defines the context in which dataplane (Envoy) runs.
type DataplaneRuntime struct {
	// Path to Envoy binary.
	BinaryPath string `yaml:"binaryPath,omitempty" envconfig:"konvoy_dataplane_runtime_binary_path"`
	// Dir to store auto-generated Envoy bootstrap config in.
	ConfigDir string `yaml:"configDir,omitempty" envconfig:"konvoy_dataplane_runtime_config_dir"`
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
	if err := c.XdsServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".XdsServer is not valid"))
	}
	return
}

var _ config.Config = &XdsServer{}

func (s *XdsServer) Validate() (errs error) {
	if s.Address == "" {
		errs = multierr.Append(errs, errors.Errorf(".Address must be non-empty"))
	}
	if 65535 < s.Port {
		errs = multierr.Append(errs, errors.Errorf(".Port must be in the range [0, 65535]"))
	}
	return
}

var _ config.Config = &Dataplane{}

func (d *Dataplane) Validate() (errs error) {
	if d.Id == "" {
		errs = multierr.Append(errs, errors.Errorf(".Id must be non-empty"))
	}
	if d.Service == "" {
		errs = multierr.Append(errs, errors.Errorf(".Service must be non-empty"))
	}
	if 65535 < d.AdminPort {
		errs = multierr.Append(errs, errors.Errorf(".AdminPort must be in the range [0, 65535]"))
	}
	return
}

var _ config.Config = &DataplaneRuntime{}

func (d *DataplaneRuntime) Validate() (errs error) {
	if d.BinaryPath == "" {
		errs = multierr.Append(errs, errors.Errorf(".BinaryPath must be non-empty"))
	}
	if d.ConfigDir == "" {
		errs = multierr.Append(errs, errors.Errorf(".ConfigDir must be non-empty"))
	}
	return
}
