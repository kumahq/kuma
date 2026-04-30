package framework

import (
	"os"
	"sync"
	"time"
)

// E2E suites on shared CI runners (4 vCPUs) routinely create several
// workload pods near-simultaneously: a single test calls Deploy() against
// kuma-1 and kuma-2 in parallel via Parallel(), each Deploy spawns a
// kuma-init + kuma-dp + envoy + payload container, and the control plane
// reacts with KDS pushes. The combined startup wave saturates the host's
// CPU and individual pods miss their readiness budget, producing the
// long-tail "Pod Pending" timeouts that have been the dominant failure
// mode for the multizone matrix. The new sidecars compete on CFS shares
// without quota limits, so the symptom presents as multi-minute apparent
// freezes rather than throttle counters going up.
//
// We damp the wave at the source by serializing K8sCluster.Deploy across
// every cluster on the runner. The semaphore is process-global by design:
// a per-cluster lock would still let kuma-1 and kuma-2 deploys race each
// other on the shared host, which is exactly the case that hurts.
//
// KUMA_DEPLOY_THROTTLE_COOLDOWN overrides the post-deploy settle period.
// The gate is held for this duration after deployment.Deploy returns so
// the next Deploy waits while the just-spawned sidecars finish their
// iptables / envoy bootstrap / KDS handshake. Default is 2s, chosen
// from the observed ~2s burst of high-CPU sidecar startup activity on
// the freezing CI runs. Set to 0 to keep serialization without any
// post-deploy wait.
//
// KUMA_DEPLOY_THROTTLE_DISABLED skips the gate entirely - no
// serialization, no cooldown. Useful for local development where the
// runner has many cores and the throttle would just slow tests down.

const defaultDeployCooldown = 2 * time.Second

var (
	deployGate         = make(chan struct{}, 1)
	deployCooldownOnce sync.Once
	deployCooldown     time.Duration
	deployDisabled     bool
)

func loadDeployThrottleEnv() {
	deployCooldownOnce.Do(func() {
		if v := os.Getenv("KUMA_DEPLOY_THROTTLE_DISABLED"); v == "1" || v == "true" {
			deployDisabled = true
			return
		}
		deployCooldown = defaultDeployCooldown
		if v := os.Getenv("KUMA_DEPLOY_THROTTLE_COOLDOWN"); v != "" {
			if d, err := time.ParseDuration(v); err == nil {
				deployCooldown = d
			}
		}
	})
}

// withDeployGate runs fn under the global deploy gate. The gate
// serializes Deploy() across all clusters on this runner. After fn
// returns, the gate is held for deployCooldown so the just-spawned pods
// have wall-clock time to finish their startup work before the next
// Deploy adds more load.
func withDeployGate(fn func() error) error {
	loadDeployThrottleEnv()
	if deployDisabled {
		return fn()
	}
	deployGate <- struct{}{}
	defer func() {
		if deployCooldown > 0 {
			time.Sleep(deployCooldown)
		}
		<-deployGate
	}()
	return fn()
}
