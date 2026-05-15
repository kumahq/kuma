package framework

import "testing"

func TestUsesUniversalLeaderElection(t *testing.T) {
	cluster := &K8sCluster{}
	if cluster.usesUniversalLeaderElection() {
		t.Fatal("expected Kubernetes leader election by default")
	}

	cluster.opts.helmOpts = map[string]string{
		"controlPlane.environment": "universal",
	}
	if !cluster.usesUniversalLeaderElection() {
		t.Fatal("expected universal leader election for universal control-plane")
	}
}

func TestHasLeaderMetric(t *testing.T) {
	if !hasLeaderMetric("leader{zone=\"zone-1\"} 1\n") {
		t.Fatal("expected leader metric")
	}
	if hasLeaderMetric("leader{zone=\"zone-1\"} 0\n") {
		t.Fatal("expected non-leader metric to be rejected")
	}
	if hasLeaderMetric("not_leader{zone=\"zone-1\"} 1\n") {
		t.Fatal("expected different metric to be rejected")
	}
}
