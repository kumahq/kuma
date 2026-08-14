package framework

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("KnownNack", func() {
	// The signature the helm upgrade suite exempts while the 2.14 fix waits
	// on a release.
	const metrics = `
# HELP kds_nack_total Total KDS NACKs sent by zone and resource type.
# TYPE kds_nack_total counter
kds_nack_total{resource_type="Dataplane",zone="Global",zone_name="kuma-2"} 2
# HELP kds_delta_requests_received Number of confirmations requests from a client
# TYPE kds_delta_requests_received counter
kds_delta_requests_received{confirmation="NACK",error_type="other",type_url="Mesh",zone="Global"} 1
`

	exempt := func(opts ...CPAssertionOpt) map[string]string {
		options := cpAssertionOpts{}
		for _, opt := range opts {
			opt(&options)
		}
		return options.knownNacks
	}

	It("should exempt only the named family", func() {
		nacks, err := findNacks(metrics, exempt(KnownNack("kds_nack_total", "fixed, waiting on release")))
		Expect(err).ToNot(HaveOccurred())
		Expect(nacks).To(ConsistOf(`kds_delta_requests_received{confirmation="NACK",error_type="other",type_url="Mesh",zone="Global"} = 1 (tolerated 0)`))
	})

	It("should report the family when it is not exempt", func() {
		nacks, err := findNacks(metrics, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(nacks).To(ContainElement(`kds_nack_total{resource_type="Dataplane",zone="Global",zone_name="kuma-2"} = 2 (tolerated 0)`))
	})
})

var _ = Describe("findNacks", func() {
	DescribeTable("should report NACK counters over their tolerance",
		func(metrics string, expected []string) {
			nacks, err := findNacks(metrics, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(nacks).To(ConsistOf(expected))
		},
		Entry("no NACK series at all", `
# HELP kds_delta_requests_received Number of confirmations requests from a client
# TYPE kds_delta_requests_received counter
kds_delta_requests_received{confirmation="ACK",error_type="no_error",type_url="Dataplane",zone="Global"} 12
# HELP xds_requests_received Number of confirmations requests from a client
# TYPE xds_requests_received counter
xds_requests_received{confirmation="ACK",error_type="no_error",type_url="type.googleapis.com/envoy.config.cluster.v3.Cluster"} 4
`, nil),
		// A zone that keeps sending a resource this CP rejects, as in the
		// helm upgrade suite where a 2.14 zone leaks VIP outbounds onto a
		// Dataplane. kds_nack_total has no confirmation label.
		Entry("KDS NACK sent by this CP", `
# HELP kds_nack_total Total KDS NACKs sent by zone and resource type.
# TYPE kds_nack_total counter
kds_nack_total{resource_type="Dataplane",zone="Global",zone_name="kuma-2"} 1
`, []string{`kds_nack_total{resource_type="Dataplane",zone="Global",zone_name="kuma-2"} = 1 (tolerated 0)`}),
		Entry("KDS NACK received from the peer CP", `
# HELP kds_delta_requests_received Number of confirmations requests from a client
# TYPE kds_delta_requests_received counter
kds_delta_requests_received{confirmation="ACK",error_type="no_error",type_url="Mesh",zone="Global"} 3
kds_delta_requests_received{confirmation="NACK",error_type="other",type_url="Mesh",zone="Global"} 1
`, []string{`kds_delta_requests_received{confirmation="NACK",error_type="other",type_url="Mesh",zone="Global"} = 1 (tolerated 0)`}),
		// A Secret the user created on both global and the zone: the zone
		// declines to overwrite its own copy and NACKs on purpose, which
		// test/e2e_env/multizone/sync asserts.
		Entry("user-caused KDS NACK is not a defect", `
# HELP kds_delta_requests_received Number of confirmations requests from a client
# TYPE kds_delta_requests_received counter
kds_delta_requests_received{confirmation="NACK",error_type="user",type_url="Secret",zone="Global"} 1
`, nil),
		Entry("user-caused NACK does not mask a real one on the same family", `
# HELP kds_delta_requests_received Number of confirmations requests from a client
# TYPE kds_delta_requests_received counter
kds_delta_requests_received{confirmation="NACK",error_type="user",type_url="Secret",zone="Global"} 1
kds_delta_requests_received{confirmation="NACK",error_type="other",type_url="Mesh",zone="Global"} 1
`, []string{`kds_delta_requests_received{confirmation="NACK",error_type="other",type_url="Mesh",zone="Global"} = 1 (tolerated 0)`}),
		// Proxy-facing xDS keeps its existing tolerance: an Envoy NACK can
		// resolve itself on the next config push.
		Entry("proxy xDS NACKs within tolerance", `
# HELP xds_requests_received Number of confirmations requests from a client
# TYPE xds_requests_received counter
xds_requests_received{confirmation="NACK",error_type="other",type_url="type.googleapis.com/envoy.config.listener.v3.Listener"} 2
`, nil),
		Entry("proxy xDS NACKs over tolerance", `
# HELP xds_requests_received Number of confirmations requests from a client
# TYPE xds_requests_received counter
xds_requests_received{confirmation="NACK",error_type="other",type_url="type.googleapis.com/envoy.config.listener.v3.Listener"} 3
`, []string{`xds_requests_received{confirmation="NACK",error_type="other",type_url="type.googleapis.com/envoy.config.listener.v3.Listener"} = 3 (tolerated 2)`}),
	)
})
