package universal

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	config_types "github.com/kumahq/kuma/pkg/config/types"
)

func DefaultUniversalRuntimeConfig() *UniversalRuntimeConfig {
	return &UniversalRuntimeConfig{
		DataplaneCleanupAge: config_types.Duration{Duration: 3 * 24 * time.Hour},
	}
}

// Universal-specific configuration
type UniversalRuntimeConfig struct {
	// DataplaneCleanupAge defines how long Dataplane should be offline to be cleaned up by GC
	DataplaneCleanupAge config_types.Duration `json:"dataplaneCleanupAge" envconfig:"kuma_runtime_universal_dataplane_cleanup_age"`
}

func (u *UniversalRuntimeConfig) Sanitize() {
}

func (u *UniversalRuntimeConfig) Validate() error {
	var errs error
	if u.DataplaneCleanupAge.Duration <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".DataplaneCleanupAge must be positive"))
	}
	return errs
}
