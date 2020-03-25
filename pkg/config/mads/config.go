package mads

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/Kong/kuma/pkg/config"
)

func DefaultMonitoringAssignmentServerConfig() *MonitoringAssignmentServerConfig {
	return &MonitoringAssignmentServerConfig{
		GrpcPort:                  5676,
		AssignmentRefreshInterval: 1 * time.Second,
	}
}

// Monitoring Assignment Discovery Service (MADS) server configuration.
type MonitoringAssignmentServerConfig struct {
	// Port of a gRPC server that serves Monitoring Assignment Discovery Service (MADS).
	GrpcPort uint32 `yaml:"grpcPort" envconfig:"kuma_monitoring_assignment_server_grpc_port"`

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
	if c.AssignmentRefreshInterval <= 0 {
		return errors.New(".AssignmentRefreshInterval must be positive")
	}
	return
}
