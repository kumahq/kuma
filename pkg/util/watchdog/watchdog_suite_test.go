package watchdog_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestWatchdog(t *testing.T) {
	test.RunSpecs(t, "Watchdog Suite")
}
