package multizone

import (
	"strings"
	"testing"
	"time"

	config_types "github.com/kumahq/kuma/v3/pkg/config/types"
)

func TestEventBasedWatchdogConfigValidateRejectsNonPositiveIntervals(t *testing.T) {
	cfg := EventBasedWatchdogConfig{
		FlushInterval:      config_types.Duration{Duration: 0},
		FullResyncInterval: config_types.Duration{Duration: -1 * time.Second},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), ".FlushInterval must be positive") {
		t.Fatalf("expected flush interval validation error, got %v", err)
	}
	if !strings.Contains(err.Error(), ".FullResyncInterval must be positive") {
		t.Fatalf("expected full resync validation error, got %v", err)
	}
}
