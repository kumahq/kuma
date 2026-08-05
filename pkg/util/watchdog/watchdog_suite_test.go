package watchdog_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestWatchdog(t *testing.T) {
	test.RunSpecs(t, "Watchdog Suite")
}
