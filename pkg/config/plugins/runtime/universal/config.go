package universal

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func DefaultUniversalRuntimeConfig() *UniversalRuntimeConfig {
	return &UniversalRuntimeConfig{
		DataplaneCleanupAge: 3 * 24 * time.Hour,
	}
}

// Universal-specific configuration
type UniversalRuntimeConfig struct {
	// DataplaneCleanupAge defines how long Dataplane should be offline to be cleaned up by GC
	DataplaneCleanupAge time.Duration `yaml:"dataplaneCleanupAge" envconfig:"kuma_runtime_universal_dataplane_cleanup_age"`
}

func (u *UniversalRuntimeConfig) Sanitize() {
}

func (u *UniversalRuntimeConfig) Validate() (errs error) {
	if u.DataplaneCleanupAge <= 0 {
		errs = multierr.Append(errs, errors.Errorf(".DataplaneCleanupAge must be positive"))
	}
	return
}
