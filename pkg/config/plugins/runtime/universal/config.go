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
		Workload: WorkloadConfig{
			GenerationInterval:  config_types.Duration{Duration: 2 * time.Second},
			DeletionGracePeriod: config_types.Duration{Duration: 1 * time.Hour},
		},
	}
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
	// Workload holds configuration for Workload generation features
	Workload WorkloadConfig `json:"workload"`
}

// WorkloadConfig holds configuration for Workload generation from Dataplanes.
type WorkloadConfig struct {
	// GenerationInterval is how often we check whether Workloads need to be
	// generated from Dataplanes
	GenerationInterval config_types.Duration `json:"generationInterval" envconfig:"KUMA_RUNTIME_UNIVERSAL_WORKLOAD_GENERATION_INTERVAL"`
	// DeletionGracePeriod is how long we wait before deleting a Workload if
	// all Dataplanes are gone
	DeletionGracePeriod config_types.Duration `json:"deletionGracePeriod" envconfig:"KUMA_RUNTIME_UNIVERSAL_WORKLOAD_DELETION_GRACE_PERIOD"`
}

func (i WorkloadConfig) Validate() error {
	return nil
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
