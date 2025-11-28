package universal

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/v2/pkg/config"
	config_types "github.com/kumahq/kuma/v2/pkg/config/types"
)

func DefaultUniversalRuntimeConfig() *UniversalRuntimeConfig {
	return &UniversalRuntimeConfig{
		DataplaneCleanupAge:    config_types.Duration{Duration: 3 * 24 * time.Hour},
		ZoneResourceCleanupAge: config_types.Duration{Duration: 3 * 24 * time.Hour},
		VIPRefreshInterval:     config_types.Duration{Duration: 500 * time.Millisecond},
		Spire: Spire{
			SocketPath: "/tmp/spire-agent/public/api.sock",
		},
	}
}

type Spire struct {
	// SocketPath is the location of the socket file one the host
	SocketPath string `json:"socketPath" envconfig:"kuma_runtime_universal_spire_socket_path"`
}

var _ config.Config = &UniversalRuntimeConfig{}

// UniversalRuntimeConfig defines Universal-specific configuration
type UniversalRuntimeConfig struct {
	config.BaseConfig

	// DataplaneCleanupAge defines how long Dataplane should be offline to be cleaned up by GC
	DataplaneCleanupAge config_types.Duration `json:"dataplaneCleanupAge" envconfig:"kuma_runtime_universal_dataplane_cleanup_age"`
	// ZoneResourceCleanupAge defines how long ZoneIngress and ZoneEgress should be offline to be cleaned up by GC
	ZoneResourceCleanupAge config_types.Duration `json:"zoneResourceCleanupAge" envconfig:"kuma_runtime_universal_zone_resource_cleanup_age"`
	// VIPRefreshInterval defines how often all meshes' VIPs should be recomputed
	VIPRefreshInterval config_types.Duration `json:"vipRefreshInterval" envconfig:"kuma_runtime_universal_vip_refresh_interval"`
	// Spire defines default configuration of the spire properties
	Spire Spire `json:"spire"`
}

func (u *UniversalRuntimeConfig) Validate() error {
	var errs error
	if u.DataplaneCleanupAge.Duration <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".DataplaneCleanupAge must be positive"))
	}
	if u.ZoneResourceCleanupAge.Duration <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".ZoneResourceCleanupAge must be positive"))
	}
	if u.VIPRefreshInterval.Duration <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".VIPRefreshInterval must be positive"))
	}
	return errs
}
