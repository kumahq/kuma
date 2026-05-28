package readiness

import "time"

var NewReporterWithDeadline = newReporterWithDeadline

const (
	StateReady          = stateReady
	StateNotReady       = stateNotReady
	StateRedirectFailed = stateRedirectFailed
)

// SetTproxyProbe overrides the transparent-proxy self-test function used by
// the readiness reporter. Tests stub this to simulate iptables-redirect-broken
// state without needing a real netns iptables setup.
func (r *Reporter) SetTproxyProbe(p func(useIPv6 bool, timeout time.Duration) error) {
	r.tproxyProbe = p
}

// ResetTproxyCacheForTest clears the cached self-test result so the next
// readiness probe will rerun the dial. Tests use this to observe a
// state change without waiting tproxyCheckInterval.
func (r *Reporter) ResetTproxyCacheForTest() {
	r.tproxyLastChecked.Store(0)
}
