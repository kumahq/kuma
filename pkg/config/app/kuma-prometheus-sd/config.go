package kuma_prometheus_sd

import (
	"net/url"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/Kong/kuma/pkg/config"
)

func DefaultConfig() Config {
	return Config{
		ControlPlane: ControlPlaneConfig{
			ApiServer: ApiServerConfig{
				URL: "http://localhost:5681",
			},
		},
		MonitoringAssignment: MonitoringAssignmentConfig{
			Client: MonitoringAssignmentClientConfig{
				Name: "kuma_sd",
			},
		},
		Prometheus: PrometheusConfig{
			OutputFile: "kuma.file_sd.json",
		},
	}
}

// Config defines configuration of the Prometheus service discovery adapter.
type Config struct {
	// ControlPlane defines coordinates of the Kuma Control Plane.
	ControlPlane ControlPlaneConfig `yaml:"controlPlane,omitempty"`
	// MonitoringAssignment defines configuration related to Monitoring Assignment in Kuma.
	MonitoringAssignment MonitoringAssignmentConfig `yaml:"monitoringAssignment,omitempty"`
	// Prometheus defines configuration related to integration with Prometheus.
	Prometheus PrometheusConfig `yaml:"prometheus,omitempty"`
}

var _ config.Config = &Config{}

func (c *Config) Sanitize() {
	c.ControlPlane.Sanitize()
	c.MonitoringAssignment.Sanitize()
	c.Prometheus.Sanitize()
}

func (c *Config) Validate() (errs error) {
	if err := c.ControlPlane.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ControlPlane is not valid"))
	}
	if err := c.MonitoringAssignment.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".MonitoringAssignment is not valid"))
	}
	if err := c.Prometheus.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Prometheus is not valid"))
	}
	return
}

// ControlPlaneConfig defines coordinates of the Control Plane.
type ControlPlaneConfig struct {
	// ApiServer defines coordinates of the Control Plane API Server
	ApiServer ApiServerConfig `yaml:"apiServer,omitempty"`
}

// ApiServerConfig defines coordinates of the Control Plane API Server.
type ApiServerConfig struct {
	// Address defines the address of Control Plane API server.
	URL string `yaml:"url,omitempty" envconfig:"kuma_control_plane_api_server_url"`
}

var _ config.Config = &ControlPlaneConfig{}

func (c *ControlPlaneConfig) Sanitize() {
	c.ApiServer.Sanitize()
}

func (c *ControlPlaneConfig) Validate() (errs error) {
	if err := c.ApiServer.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ApiServer is not valid"))
	}
	return
}

var _ config.Config = &ApiServerConfig{}

func (c *ApiServerConfig) Sanitize() {
}

func (c *ApiServerConfig) Validate() (errs error) {
	if c.URL == "" {
		errs = multierr.Append(errs, errors.Errorf(".URL must be non-empty"))
	}
	if url, err := url.Parse(c.URL); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".URL must be a valid absolute URI"))
	} else if !url.IsAbs() {
		errs = multierr.Append(errs, errors.Errorf(".URL must be a valid absolute URI"))
	}
	return
}

// MonitoringAssignmentConfig defines configuration related to Monitoring Assignment in Kuma.
type MonitoringAssignmentConfig struct {
	// Client defines configuration of a Monitoring Assignment Discovery Service (MADS) client.
	Client MonitoringAssignmentClientConfig `yaml:"client,omitempty"`
}

// MonitoringAssignmentClientConfig defines configuration of a
// Monitoring Assignment Discovery Service (MADS) client.
type MonitoringAssignmentClientConfig struct {
	// Name this adapter should use when connecting to Monitoring Assignment server.
	Name string `yaml:"name,omitempty" envconfig:"kuma_monitoring_assignment_client_name"`
}

var _ config.Config = &MonitoringAssignmentConfig{}

func (c *MonitoringAssignmentConfig) Sanitize() {
}

func (c *MonitoringAssignmentConfig) Validate() (errs error) {
	if err := c.Client.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".Client is not valid"))
	}
	return
}

var _ config.Config = &MonitoringAssignmentClientConfig{}

func (c *MonitoringAssignmentClientConfig) Sanitize() {
}

func (c *MonitoringAssignmentClientConfig) Validate() (errs error) {
	if c.Name == "" {
		errs = multierr.Append(errs, errors.Errorf(".Name must be non-empty"))
	}
	return
}

// PrometheusConfig defines configuration related to integration with Prometheus.
//
// In short, Kuma Prometheus SD adapter integrates with Prometheus via a shared file,
// where the former is a writer and the latter is a reader.
// For further details see https://github.com/prometheus/prometheus/tree/master/documentation/examples/custom-sd
type PrometheusConfig struct {
	// Path to an output file where Kuma Prometheus SD adapter should persists a list of scrape targets.
	// The same file path must be used on Prometheus side in a configuration of `file_sd` discovery mechanism.
	OutputFile string `yaml:"outputFile,omitempty" envconfig:"kuma_prometheus_output_file"`
}

var _ config.Config = &PrometheusConfig{}

func (c *PrometheusConfig) Sanitize() {
}

func (c *PrometheusConfig) Validate() (errs error) {
	if c.OutputFile == "" {
		errs = multierr.Append(errs, errors.Errorf(".OutputFile must be non-empty"))
	}
	return
}
