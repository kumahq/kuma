package kdsmtls

import (
	"crypto"
	"crypto/tls"
	"crypto/x509/pkix"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	util_tls "github.com/kumahq/kuma/v3/pkg/tls"
	. "github.com/kumahq/kuma/v3/test/framework"
)

const (
	certsDir     = "/kuma/kds-mtls"
	clientCAFile = certsDir + "/ca.crt"
	clientCert   = certsDir + "/tls.crt"
	clientKey    = certsDir + "/tls.key"
)

func KDSMutualTLS() {
	var global, trustedZone, anonymousZone Cluster
	var tmpDir string

	writeFile := func(name string, content []byte) string {
		path := filepath.Join(tmpDir, name)
		Expect(os.WriteFile(path, content, 0o600)).To(Succeed())
		return path
	}

	issueClientCert := func(ca *util_tls.KeyPair, name string, hosts ...string) []KumaDeploymentOption {
		caPair, err := tls.X509KeyPair(ca.CertPEM, ca.KeyPEM)
		Expect(err).ToNot(HaveOccurred())
		caKey, ok := caPair.PrivateKey.(crypto.Signer)
		Expect(ok).To(BeTrue())
		pair, err := util_tls.NewCert(*caPair.Leaf, caKey, util_tls.ClientCertType, util_tls.DefaultKeyType, hosts...)
		Expect(err).ToNot(HaveOccurred())
		return []KumaDeploymentOption{
			WithCPDockerVolumes(
				writeFile(name+".crt", pair.CertPEM)+":"+clientCert,
				writeFile(name+".key", pair.KeyPEM)+":"+clientKey,
			),
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_TLS_CERT_FILE", clientCert),
			WithEnv("KUMA_MULTIZONE_ZONE_KDS_TLS_KEY_FILE", clientKey),
		}
	}

	BeforeAll(func() {
		tmpDir = GinkgoT().TempDir()
		ca, err := util_tls.GenerateCA(util_tls.DefaultKeyType, pkix.Name{CommonName: "kds-client-ca"})
		Expect(err).ToNot(HaveOccurred())

		global = NewUniversalCluster(NewTestingT(), "kds-mtls-global", Silent)
		Expect(NewClusterSetup().
			Install(Kuma(core.Global,
				WithCPDockerVolumes(writeFile("ca.crt", ca.CertPEM)+":"+clientCAFile),
				WithEnv("KUMA_MULTIZONE_GLOBAL_KDS_TLS_CLIENT_CA_FILE", clientCAFile),
				WithEnv("KUMA_MULTIZONE_GLOBAL_KDS_REQUIRE_CLIENT_CERT", "true"),
			)).
			Setup(global)).To(Succeed())

		trustedZone = NewUniversalCluster(NewTestingT(), "kds-mtls-trusted", Silent)
		anonymousZone = NewUniversalCluster(NewTestingT(), "kds-mtls-anonymous", Silent)

		zoneOpts := func(extra ...KumaDeploymentOption) []KumaDeploymentOption {
			return append([]KumaDeploymentOption{WithGlobalAddress(global.GetKuma().GetKDSServerAddress())}, extra...)
		}
		Expect(NewClusterSetup().
			Install(Kuma(core.Zone, zoneOpts(issueClientCert(ca, "trusted", trustedZone.ZoneName())...)...)).
			Setup(trustedZone)).To(Succeed())
		Expect(NewClusterSetup().
			Install(Kuma(core.Zone, zoneOpts()...)).
			Setup(anonymousZone)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(global, "default")
		DebugUniversal(trustedZone, "default")
		DebugUniversal(anonymousZone, "default")
	})

	E2EAfterAll(func() {
		for _, c := range []Cluster{trustedZone, anonymousZone, global} {
			ControlPlaneAssertions(c)
			Expect(c.DeleteKuma()).To(Succeed())
			Expect(c.DismissCluster()).To(Succeed())
		}
	})

	It("should connect a zone presenting a client certificate", func() {
		Expect(WaitForZoneOnline(global, trustedZone.ZoneName())).To(Succeed())
	})

	It("should sync resources to the zone connected over mTLS", func() {
		Expect(global.Install(MeshUniversal("kds-mtls"))).To(Succeed())

		Eventually(func(g Gomega) {
			out, err := trustedZone.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring("kds-mtls"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should reject a zone that doesn't present a certificate", func() {
		// the trusted zone connecting to the same Global CP rules out any other reason for the zone to be missing
		Expect(WaitForZoneOnline(global, trustedZone.ZoneName())).To(Succeed())
		Consistently(func(g Gomega) {
			out, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(ContainSubstring(anonymousZone.ZoneName()))
		}, "60s", "1s").Should(Succeed())
	})
}
