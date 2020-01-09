package kuma_prometheus_sd

import (
	"net/url"

	"github.com/Kong/kuma/pkg/config"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
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
	}
}

// Config defines configuration of the Kuma Prometheus SD
// (Prometheus service discovery adapter).
type Config struct {
	// ControlPlane defines coordinates of the Kuma Control Plane.
	ControlPlane ControlPlaneConfig `yaml:"controlPlane,omitempty"`
	// MonitoringAssignment defines configuration related to Monitoring Assignment in Kuma.
	MonitoringAssignment MonitoringAssignmentConfig `yaml:"monitoringAssignment,omitempty"`
}

var _ config.Config = &Config{}

func (c *Config) Sanitize() {
	c.ControlPlane.Sanitize()
	c.MonitoringAssignment.Sanitize()
}

func (c *Config) Validate() (errs error) {
	if err := c.ControlPlane.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".ControlPlane is not valid"))
	}
	if err := c.MonitoringAssignment.Validate(); err != nil {
		errs = multierr.Append(errs, errors.Wrapf(err, ".MonitoringAssignment is not valid"))
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
