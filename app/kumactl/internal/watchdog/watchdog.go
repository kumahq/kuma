// Package watchdog spawns a goroutine at package init time that exits
// the process if it is still running after a hard deadline. Imported as
// a side-effect from main as early as possible so the watchdog goroutine
// is scheduled before any heavy package init runs (notably
// controller-runtime/pkg/cache/internal, which has been observed
// hanging the kuma-init container indefinitely on calico clusters).
//
// The watchdog is only armed when kumactl is invoked as the kuma-init
// container entrypoint (i.e. `kumactl install transparent-proxy`). All
// other invocations are unaffected so that long-running interactive
// commands are not killed.
//
// Disabled when KUMACTL_WATCHDOG_DISABLED=1.
package watchdog

import (
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

const defaultTimeout = 120 * time.Second

var done atomic.Bool

func init() {
	if !shouldArm() {
		return
	}
	timeout := defaultTimeout
	if v := os.Getenv("KUMACTL_WATCHDOG_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = time.Duration(n) * time.Second
		}
	}
	go func() {
		time.Sleep(timeout)
		if done.Load() {
			return
		}
		_, _ = os.Stderr.WriteString("kumactl watchdog: process exceeded deadline, exiting to allow restart\n")
		os.Exit(2)
	}()
}

func shouldArm() bool {
	if os.Getenv("KUMACTL_WATCHDOG_DISABLED") == "1" {
		return false
	}
	// Only arm for `kumactl install transparent-proxy` (the kuma-init
	// container entrypoint). Anything else is left alone.
	args := os.Args
	if len(args) < 3 {
		return false
	}
	return args[1] == "install" && args[2] == "transparent-proxy"
}

// Disarm stops the watchdog from killing the process. Call this from main
// once kumactl has finished its work.
func Disarm() {
	done.Store(true)
}
