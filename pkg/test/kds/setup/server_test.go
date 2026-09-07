package setup

import (
	"testing"
	"time"

	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

func TestKdsServerBuilderUsesFastEventBasedWatchdogIntervals(t *testing.T) {
	builder := NewKdsServerBuilder(memory.NewStore())

	if got := builder.rt.cfg.Multizone.Global.KDS.EventBasedWatchdog.FlushInterval.Duration; got != 100*time.Millisecond {
		t.Fatalf("expected global flush interval to be 100ms, got %s", got)
	}
	if got := builder.rt.cfg.Multizone.Global.KDS.EventBasedWatchdog.FullResyncInterval.Duration; got != 100*time.Millisecond {
		t.Fatalf("expected global full resync interval to be 100ms, got %s", got)
	}

	builder.AsZone("zone-1")

	if got := builder.rt.cfg.Multizone.Zone.KDS.EventBasedWatchdog.FlushInterval.Duration; got != 100*time.Millisecond {
		t.Fatalf("expected zone flush interval to be 100ms, got %s", got)
	}
	if got := builder.rt.cfg.Multizone.Zone.KDS.EventBasedWatchdog.FullResyncInterval.Duration; got != 100*time.Millisecond {
		t.Fatalf("expected zone full resync interval to be 100ms, got %s", got)
	}
}
