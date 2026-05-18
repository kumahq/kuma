package framework

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	Describe("leaseHeldByCurrentPod", func() {
		It("matches holder by current pod prefix", func() {
			holder := "kuma-control-plane-1_57b1f32d"

			Expect(leaseHeldByCurrentPod(&holder, []string{"kuma-control-plane-1"})).To(BeTrue())
		})

		It("rejects stale holder", func() {
			holder := "kuma-control-plane-old_57b1f32d"

			Expect(leaseHeldByCurrentPod(&holder, []string{"kuma-control-plane-new"})).To(BeFalse())
		})
	})

	Describe("controlPlaneLeaseExpired", func() {
		It("does not expire a live lease", func() {
			now := time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)
			leaseDurationSeconds := int32(15)
			renewTime := metav1.NewMicroTime(now.Add(-10 * time.Second))

			lease := &coordinationv1.Lease{
				Spec: coordinationv1.LeaseSpec{
					LeaseDurationSeconds: &leaseDurationSeconds,
					RenewTime:            &renewTime,
				},
			}

			Expect(controlPlaneLeaseExpired(lease, now)).To(BeFalse())
		})

		It("expires a stale lease", func() {
			now := time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)
			leaseDurationSeconds := int32(15)
			renewTime := metav1.NewMicroTime(now.Add(-16 * time.Second))

			lease := &coordinationv1.Lease{
				Spec: coordinationv1.LeaseSpec{
					LeaseDurationSeconds: &leaseDurationSeconds,
					RenewTime:            &renewTime,
				},
			}

			Expect(controlPlaneLeaseExpired(lease, now)).To(BeTrue())
		})
	})
})
