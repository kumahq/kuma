package certs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"
	"regexp"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	util_tls "github.com/kumahq/kuma/v3/pkg/tls"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
)

func Rotation() {
	const mesh = "cert-rotation"
	// the certificate the control plane autogenerates on the first run lands
	// in the working directory, replacing it there simulates what cert-manager
	// does with a mounted secret
	const certFile = "/kuma/kuma-cp.crt"
	const keyFile = "/kuma/kuma-cp.key"

	var cluster *UniversalCluster

	// servedCert opens a TLS connection to the DP server, which is the port
	// every data plane proxy connects to, and returns the certificate the
	// control plane presented.
	servedCert := func(g Gomega) *x509.Certificate {
		port := cluster.GetApp(AppModeCP).GetPublicPort("5678")
		dialer := &tls.Dialer{Config: &tls.Config{
			InsecureSkipVerify: true, // #nosec G402 -- the test asserts on the served certificate itself
			MinVersion:         tls.VersionTLS12,
		}}
		conn, err := dialer.DialContext(context.Background(), "tcp", net.JoinHostPort("localhost", port))
		g.Expect(err).ToNot(HaveOccurred())
		defer func() {
			g.Expect(conn.Close()).To(Succeed())
		}()
		return conn.(*tls.Conn).ConnectionState().PeerCertificates[0]
	}

	// xdsConnections is how many times the proxy has connected to the DP server,
	// which does not change while its stream stays up.
	xdsConnectionsRE := regexp.MustCompile(`cluster\.ads_cluster\.upstream_cx_total: (\d+)`)
	xdsConnections := func(g Gomega, app string) int {
		stats, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput(
			"inspect", "dataplane", app, "--type", "stats", "--mesh", mesh)
		g.Expect(err).ToNot(HaveOccurred())
		matched := xdsConnectionsRE.FindStringSubmatch(stats)
		g.Expect(matched).To(HaveLen(2), "no ads_cluster connection stat in %s", stats)
		connections, err := strconv.Atoi(matched[1])
		g.Expect(err).ToNot(HaveOccurred())
		return connections
	}

	trafficWorks := func(app string) {
		GinkgoHelper()
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(cluster, app, "test-server.svc.mesh.local")
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	}

	online := func(app string) {
		GinkgoHelper()
		Eventually(func(g Gomega) {
			out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(MatchRegexp(app + `.*Online`))
		}, "60s", "1s").Should(Succeed())
	}

	writeFile := func(path string, content []byte) {
		encoded := base64.StdEncoding.EncodeToString(content)
		_, _, err := cluster.Exec("", "", AppModeCP, fmt.Sprintf("echo %s | base64 -d > %s", encoded, path))
		Expect(err).ToNot(HaveOccurred())
	}

	BeforeAll(func() {
		cluster = NewUniversalCluster(NewTestingT(), "kuma-cert-rotation", Silent)
		Expect(NewClusterSetup().
			Install(Kuma(core.Zone,
				WithEnv("KUMA_GENERAL_WORK_DIR", "/kuma"),
			)).
			Install(MeshUniversal(mesh)).
			Install(TestServerUniversal("test-server", mesh,
				WithArgs([]string{"echo", "--instance", "test-server"}),
			)).
			Setup(cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(cluster, mesh)
	})

	AfterAll(func() {
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should serve a rotated certificate and keep proxies connected", func() {
		// given a data plane proxy connected to the control plane
		Expect(DemoClientUniversal("dp-before-rotation", mesh, WithTransparentProxy(true))(cluster)).To(Succeed())
		online("dp-before-rotation")
		trafficWorks("dp-before-rotation")

		var oldCert *x509.Certificate
		var connectionsBefore int
		Eventually(func(g Gomega) {
			oldCert = servedCert(g)
			connectionsBefore = xdsConnections(g, "dp-before-rotation")
		}, "30s", "1s").Should(Succeed())

		// when the certificate is replaced on disk, keeping the hosts it was
		// issued for so that data plane proxies keep reaching the same address
		hosts := oldCert.DNSNames
		for _, ip := range oldCert.IPAddresses {
			hosts = append(hosts, ip.String())
		}
		newKeyPair, err := util_tls.NewSelfSignedCert(util_tls.ServerCertType, util_tls.DefaultKeyType, hosts...)
		Expect(err).ToNot(HaveOccurred())
		writeFile(keyFile, newKeyPair.KeyPEM)
		writeFile(certFile, newKeyPair.CertPEM)

		// then the control plane serves it without being restarted
		Eventually(func(g Gomega) {
			g.Expect(servedCert(g).SerialNumber).ToNot(Equal(oldCert.SerialNumber))
		}, "60s", "1s").Should(Succeed())

		// and the proxy that connected before the rotation is unaffected, its
		// stream was established with the old certificate and stays up
		online("dp-before-rotation")
		trafficWorks("dp-before-rotation")
		Consistently(func(g Gomega) {
			g.Expect(xdsConnections(g, "dp-before-rotation")).To(Equal(connectionsBefore))
		}, "20s", "5s").Should(Succeed())

		// and a data plane proxy started afterwards still connects
		Expect(DemoClientUniversal("dp-after-rotation", mesh, WithTransparentProxy(true))(cluster)).To(Succeed())
		online("dp-after-rotation")
		trafficWorks("dp-after-rotation")
	})
}
