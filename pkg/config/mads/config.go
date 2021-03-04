package mads

import (
	"github.com/kumahq/kuma/pkg/mads"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
)

func DefaultMonitoringAssignmentServerConfig() *MonitoringAssignmentServerConfig {
	return &MonitoringAssignmentServerConfig{
		GrpcPort:                  5676,
		GrpcEnabled:               true,
		HttpPort:                  5677,
		HttpEnabled:               true,
		HttpTimeout:               30 * time.Second,
		ApiVersions:               []mads.ApiVersion{mads.MADS_V1, mads.MADS_V1_ALPHA1},
		AssignmentRefreshInterval: 1 * time.Second,
	}
}

// Monitoring Assignment Discovery Service (MADS) server configuration.
type MonitoringAssignmentServerConfig struct {
	// Port of a gRPC server that serves Monitoring Assignment Discovery Service (MADS).
	GrpcPort uint32 `yaml:"grpcPort" envconfig:"kuma_monitoring_assignment_server_grpc_port"`
	// Whether to run a gRPC server
	GrpcEnabled bool `yaml:"grpcEnabled" envconfig:"kuma_monitoring_assignment_server_grpc_enabled"`
	// Port of a HTTP server that serves Monitoring Assignment Discovery Service (MADS)
	HttpPort uint32 `yaml:"httpPort" envconfig:"kuma_monitoring_assignment_server_http_port"`
	// Whether to run a HTTP discovery server. Only available for v1.
	HttpEnabled bool `yaml:"httpEnabled" envconfig:"kuma_monitoring_assignment_server_http_enabled"`
	// The timeout for a single HTTP discovery request.
	HttpTimeout time.Duration `yaml:"httpTimeout" envconfig:"kuma_monitoring_assignment_server_http_timeout"`
	// Which observability apiVersions to serve
	ApiVersions []string `yaml:"apiVersions" envconfig:"kuma_monitoring_assignment_server_api_versions"`
	// Interval for re-generating monitoring assignments for clients connected to the Control Plane.
	AssignmentRefreshInterval time.Duration `yaml:"assignmentRefreshInterval" envconfig:"kuma_monitoring_assignment_server_assignment_refresh_interval"`
}

var _ config.Config = &MonitoringAssignmentServerConfig{}

func (c *MonitoringAssignmentServerConfig) Sanitize() {
}

func (c *MonitoringAssignmentServerConfig) Validate() (errs error) {
	if 65535 < c.GrpcPort {
		errs = multierr.Append(errs, errors.Errorf(".GrpcPort must be in the range [0, 65535]"))
	}
	if 65535 < c.HttpPort {
		errs = multierr.Append(errs, errors.Errorf(".HttpPort must be in the range [0, 65535]"))
	}

	if c.GrpcPort == c.HttpPort {
		errs = multierr.Append(errs, errors.Errorf(".HttpPort and .GrpcPort must be different"))
	}

	if len(c.ApiVersions) == 0 {
		errs = multierr.Append(errs, errors.Errorf(".ApiVersions must contain at least one version"))
	}

	for _, apiVersion := range c.ApiVersions {
		if apiVersion != mads.MADS_V1 && apiVersion != mads.MADS_V1_ALPHA1 {
			errs = multierr.Append(errs, errors.Errorf(".ApiVersions contains invalid version %s", apiVersion))
		}
	}

	if c.AssignmentRefreshInterval <= 0 {
		return errors.New(".AssignmentRefreshInterval must be positive")
	}
	return
}

// VersionIsEnabled checks whether a MADS version has been enabled and should be served.
func (c *MonitoringAssignmentServerConfig) VersionIsEnabled(apiVersion mads.ApiVersion) bool {
	for _, version := range c.ApiVersions {
		if apiVersion == version {
			return true
		}
	}
	return false
}
