package certs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	util_tls "github.com/kumahq/kuma/v3/pkg/tls"
	. "github.com/kumahq/kuma/v3/test/framework"
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
			Setup(cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(cluster, mesh)
	})

	AfterAll(func() {
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should serve a rotated certificate and keep accepting new data plane proxies", func() {
		// given a data plane proxy connected to the control plane
		Expect(DemoClientUniversal("dp-before-rotation", mesh, WithTransparentProxy(true))(cluster)).To(Succeed())
		Eventually(func() (string, error) {
			return cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", mesh)
		}, "30s", "1s").Should(ContainSubstring("Online"))

		var oldCert *x509.Certificate
		Eventually(func(g Gomega) {
			oldCert = servedCert(g)
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

		// and a data plane proxy started afterwards still connects
		Expect(DemoClientUniversal("dp-after-rotation", mesh, WithTransparentProxy(true))(cluster)).To(Succeed())
		Eventually(func(g Gomega) {
			out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(MatchRegexp(`dp-after-rotation.*Online`))
		}, "60s", "1s").Should(Succeed())
	})
}
