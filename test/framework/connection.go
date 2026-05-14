package framework

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/stats"
)

// HoldConnection opens a long-lived HTTP/1.1 keep-alive TCP connection from
// srcApp to dstApp:dstPort inside a UniversalCluster, then blocks until Envoy
// on dstApp reports the connection on the listener bound to dstPort
// (downstream_cx_active > 0). It returns a cancel function that terminates the
// keep-alive by killing ncat on srcApp.
//
// Use this primitive when a test needs to "occupy a connection slot" before
// asserting listener-level connection-limit behavior. It removes the race
// between a fire-and-forget goroutine that runs ncat and the assertion that
// follows it — if ncat fails to connect, this fails fast at the call site
// instead of producing a misleading downstream failure 60s later.
//
// The test fails if the connection is not observed within 30s.
func HoldConnection(cluster *UniversalCluster, srcApp, dstApp string, dstPort uint32) func() {
	GinkgoHelper()

	dst := cluster.GetApp(dstApp)
	Expect(dst).ToNot(BeNil(), "destination app %q not found", dstApp)
	dstIP := dst.GetIP()
	addr := net.JoinHostPort(dstIP, fmt.Sprintf("%d", dstPort))

	// Send HTTP/1.1 keep-alive requests in a loop to prevent
	// request_headers_timeout from closing the TCP connection.
	cmd := fmt.Sprintf(
		`bash -c 'while true; do printf "GET / HTTP/1.1\r\nHost: %s\r\nConnection: keep-alive\r\n\r\n"; sleep 0.3; done | ncat %s %d'`,
		addr, dstIP, dstPort,
	)

	go func() {
		defer GinkgoRecover()
		_, _, _ = cluster.Exec("", "", srcApp, cmd)
	}()

	// Envoy names its listener stats by bound address: listener.<ip>_<port>.*
	listenerStat := fmt.Sprintf(`listener\..*_%d\.downstream_cx_active`, dstPort)
	dstAdmin := dst.GetEnvoyAdminTunnel()
	Eventually(func(g Gomega) {
		s, err := dstAdmin.GetStats(listenerStat)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(s).To(stats.BeGreaterThanZero())
	}, "30s", "1s").Should(Succeed(),
		"ncat from %q did not establish a connection to %s:%d", srcApp, dstIP, dstPort)

	return func() {
		Expect(cluster.Kill(srcApp, "ncat")).To(Succeed())
	}
}
