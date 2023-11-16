package universal

import (
	"fmt"
	"time"

	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

func DefaultUniversalRuntimeConfig() *UniversalRuntimeConfig {
	return &UniversalRuntimeConfig{
		DataplaneCleanupAge: config_types.Duration{Duration: 3 * 24 * time.Hour},
	}
}

var _ config.Config = &UniversalRuntimeConfig{}

// UniversalRuntimeConfig defines Universal-specific configuration
type UniversalRuntimeConfig struct {
	config.BaseConfig

	// DataplaneCleanupAge defines how long Dataplane should be offline to be cleaned up by GC
	DataplaneCleanupAge config_types.Duration `json:"dataplaneCleanupAge" envconfig:"kuma_runtime_universal_dataplane_cleanup_age"`
}

func (u *UniversalRuntimeConfig) Validate() error {
	var errs error
	if u.DataplaneCleanupAge.Duration <= 0 {
		errs = multierr.Append(errs, fmt.Errorf(".DataplaneCleanupAge must be positive"))
	}
	return errs
}
