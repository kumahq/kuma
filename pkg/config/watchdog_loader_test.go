package config_test

import (
	"testing"
	"time"

	"github.com/kumahq/kuma/v3/pkg/config"
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
)

func TestLoadStableKDSEventBasedWatchdogFromYAML(t *testing.T) {
	cfg := kuma_cp.DefaultConfig()

	err := config.NewLoader(&cfg).WithValidation().LoadBytes([]byte(`
multizone:
  global:
    kds:
      eventBasedWatchdog:
        flushInterval: 10s
        fullResyncInterval: 15s
        delayFullResync: true
  zone:
    kds:
      eventBasedWatchdog:
        flushInterval: 11s
        fullResyncInterval: 16s
        delayFullResync: true
`))
	if err != nil {
		t.Fatalf("load YAML: %v", err)
	}

	if got := cfg.Multizone.Global.KDS.EventBasedWatchdog.FlushInterval.Duration; got != 10*time.Second {
		t.Fatalf("global flushInterval = %s, want 10s", got)
	}
	if got := cfg.Multizone.Global.KDS.EventBasedWatchdog.FullResyncInterval.Duration; got != 15*time.Second {
		t.Fatalf("global fullResyncInterval = %s, want 15s", got)
	}
	if !cfg.Multizone.Global.KDS.EventBasedWatchdog.DelayFullResync {
		t.Fatal("global delayFullResync = false, want true")
	}
	if got := cfg.Multizone.Zone.KDS.EventBasedWatchdog.FlushInterval.Duration; got != 11*time.Second {
		t.Fatalf("zone flushInterval = %s, want 11s", got)
	}
	if got := cfg.Multizone.Zone.KDS.EventBasedWatchdog.FullResyncInterval.Duration; got != 16*time.Second {
		t.Fatalf("zone fullResyncInterval = %s, want 16s", got)
	}
	if !cfg.Multizone.Zone.KDS.EventBasedWatchdog.DelayFullResync {
		t.Fatal("zone delayFullResync = false, want true")
	}
}

func TestLoadStableKDSEventBasedWatchdogFromEnv(t *testing.T) {
	t.Setenv("KUMA_MULTIZONE_GLOBAL_KDS_EVENT_BASED_WATCHDOG_FLUSH_INTERVAL", "10s")
	t.Setenv("KUMA_MULTIZONE_GLOBAL_KDS_EVENT_BASED_WATCHDOG_FULL_RESYNC_INTERVAL", "15s")
	t.Setenv("KUMA_MULTIZONE_GLOBAL_KDS_EVENT_BASED_WATCHDOG_DELAY_FULL_RESYNC", "true")
	t.Setenv("KUMA_MULTIZONE_ZONE_KDS_EVENT_BASED_WATCHDOG_FLUSH_INTERVAL", "11s")
	t.Setenv("KUMA_MULTIZONE_ZONE_KDS_EVENT_BASED_WATCHDOG_FULL_RESYNC_INTERVAL", "16s")
	t.Setenv("KUMA_MULTIZONE_ZONE_KDS_EVENT_BASED_WATCHDOG_DELAY_FULL_RESYNC", "true")

	cfg := kuma_cp.DefaultConfig()
	err := config.NewLoader(&cfg).WithEnvVarsLoading("").WithValidation().LoadBytes([]byte("{}"))
	if err != nil {
		t.Fatalf("load env: %v", err)
	}

	if got := cfg.Multizone.Global.KDS.EventBasedWatchdog.FlushInterval.Duration; got != 10*time.Second {
		t.Fatalf("global flushInterval = %s, want 10s", got)
	}
	if got := cfg.Multizone.Global.KDS.EventBasedWatchdog.FullResyncInterval.Duration; got != 15*time.Second {
		t.Fatalf("global fullResyncInterval = %s, want 15s", got)
	}
	if !cfg.Multizone.Global.KDS.EventBasedWatchdog.DelayFullResync {
		t.Fatal("global delayFullResync = false, want true")
	}
	if got := cfg.Multizone.Zone.KDS.EventBasedWatchdog.FlushInterval.Duration; got != 11*time.Second {
		t.Fatalf("zone flushInterval = %s, want 11s", got)
	}
	if got := cfg.Multizone.Zone.KDS.EventBasedWatchdog.FullResyncInterval.Duration; got != 16*time.Second {
		t.Fatalf("zone fullResyncInterval = %s, want 16s", got)
	}
	if !cfg.Multizone.Zone.KDS.EventBasedWatchdog.DelayFullResync {
		t.Fatal("zone delayFullResync = false, want true")
	}
}
