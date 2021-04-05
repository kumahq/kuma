package mads

import (
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/mads"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
)

var log = core.Log.WithName("mads-config")

func DefaultMonitoringAssignmentServerConfig() *MonitoringAssignmentServerConfig {
	return &MonitoringAssignmentServerConfig{
		Port:                      5676,
		GrpcEnabled:               true,
		HttpEnabled:               true,
		FetchTimeout:              30 * time.Second,
		ApiVersions:               []mads.ApiVersion{mads.API_V1_ALPHA1, mads.API_V1},
		AssignmentRefreshInterval: 1 * time.Second,
	}
}

// Monitoring Assignment Discovery Service (MADS) server configuration.
type MonitoringAssignmentServerConfig struct {
	// GrpcPort is the port of the gRPC server that serves Monitoring Assignment Discovery Service (MADS).
	//
	// Deprecated: GrpcPort has been replaced with Port to multiplex both HTTP and gRPC
	GrpcPort uint32 `yaml:"grpcPort" envconfig:"kuma_monitoring_assignment_server_grpc_port"`
	// Port of the server that serves Monitoring Assignment Discovery Service (MADS)
	// over both grpc and http.
	Port uint32 `yaml:"port" envconfig:"kuma_monitoring_assignment_server_port"`
	// The timeout for a single fetch-based discovery request.
	FetchTimeout time.Duration `yaml:"fetchTimeout" envconfig:"kuma_monitoring_assignment_server_fetch_timeout"`
	// Which observability apiVersions to serve
	ApiVersions []string `yaml:"apiVersions" envconfig:"kuma_monitoring_assignment_server_api_versions"`
	// Interval for re-generating monitoring assignments for clients connected to the Control Plane.
	AssignmentRefreshInterval time.Duration `yaml:"assignmentRefreshInterval" envconfig:"kuma_monitoring_assignment_server_assignment_refresh_interval"`
	// Whether to run a HTTP discovery server. Only available for MADS v1.
	HttpEnabled bool `yaml:"httpEnabled" envconfig:"kuma_monitoring_assignment_server_http_enabled"`
	// Whether to run a gRPC server. Required for v1alpha1.
	GrpcEnabled bool `yaml:"grpcEnabled" envconfig:"kuma_monitoring_assignment_server_grpc_enabled"`
}

var _ config.Config = &MonitoringAssignmentServerConfig{}

func (c *MonitoringAssignmentServerConfig) Sanitize() {
}

func (c *MonitoringAssignmentServerConfig) Validate() (errs error) {
	if c.GrpcPort != 0 {
		log.V(1).Info(".GrpcPort is deprecated. Please use .Port instead")
		if c.Port == 0 {
			c.Port = c.GrpcPort
		}
	}

	if 65535 < c.Port {
		errs = multierr.Append(errs, errors.Errorf(".Port must be in the range [0, 65535]"))
	}

	if len(c.ApiVersions) == 0 {
		errs = multierr.Append(errs, errors.Errorf(".ApiVersions must contain at least one version"))
	}

	for _, apiVersion := range c.ApiVersions {
		if apiVersion != mads.API_V1 && apiVersion != mads.API_V1_ALPHA1 {
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
