package framework

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("K8sCluster", func() {
	Describe("usesUniversalLeaderElection", func() {
		It("uses Kubernetes leader election by default", func() {
			cluster := &K8sCluster{}
			Expect(cluster.usesUniversalLeaderElection()).To(BeFalse())
		})

		It("detects universal control-plane environment", func() {
			cluster := &K8sCluster{}
			cluster.opts.helmOpts = map[string]string{
				"controlPlane.environment": "universal",
			}
			Expect(cluster.usesUniversalLeaderElection()).To(BeTrue())
		})
	})

	Describe("hasLeaderMetric", func() {
		It("matches leader metric with value 1", func() {
			Expect(hasLeaderMetric("leader{zone=\"zone-1\"} 1\n")).To(BeTrue())
		})

		It("rejects non-leader metric values", func() {
			Expect(hasLeaderMetric("leader{zone=\"zone-1\"} 0\n")).To(BeFalse())
		})

		It("rejects other metric names", func() {
			Expect(hasLeaderMetric("not_leader{zone=\"zone-1\"} 1\n")).To(BeFalse())
		})
	})
})
