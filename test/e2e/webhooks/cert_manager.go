package webhooks

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/config/core"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/certmanager"
)

func CertManagerCAInjection() {
	var cluster Cluster

	BeforeAll(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		const certManagerNamespace = "cert-manager"

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		err := NewClusterSetup().
			Install(certmanager.Install(
				certmanager.WithNamespace(certManagerNamespace),
			)).
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithHelmOpt("controlPlane.tls.general.certManager.enabled", "true"),
			)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(cluster, Config.KumaNamespace)
	})

	E2EAfterAll(func() {
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should inject CA bundle into Kuma webhook configurations", func() {
		kumaNamespace := Config.KumaNamespace

		// Verify Certificate resource exists and is Ready
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(kumaNamespace),
				"get", "certificate", "kuma-tls-cert",
				"-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).To(Equal("True"))
		}, "60s", "1s").Should(Succeed(), "Certificate resource should exist and be Ready")

		// Verify Issuer resource exists and is Ready
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(kumaNamespace),
				"get", "issuer", "kuma-selfsigned-issuer",
				"-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).To(Equal("True"))
		}, "60s", "1s").Should(Succeed(), "Issuer resource should exist and be Ready")

		// Verify cert-manager annotation on validating webhook configuration
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(),
				"get", "validatingwebhookconfiguration", "kuma-validating-webhook-configuration",
				"-o", "jsonpath={.metadata.annotations.cert-manager\\.io/inject-ca-from}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).To(Equal(fmt.Sprintf("%s/kuma-tls-cert", kumaNamespace)))
		}, "60s", "1s").Should(Succeed(), "Validating webhook should have cert-manager annotation")

		// Verify cert-manager annotation on mutating webhook configuration
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(),
				"get", "mutatingwebhookconfiguration", "kuma-admission-mutating-webhook-configuration",
				"-o", "jsonpath={.metadata.annotations.cert-manager\\.io/inject-ca-from}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).To(Equal(fmt.Sprintf("%s/kuma-tls-cert", kumaNamespace)))
		}, "60s", "1s").Should(Succeed(), "Mutating webhook should have cert-manager annotation")

		// Verify CA bundle is injected into validating webhook configuration
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(),
				"get", "validatingwebhookconfiguration", "kuma-validating-webhook-configuration",
				"-o", "jsonpath={.webhooks[0].clientConfig.caBundle}",
			)
			g.Expect(err).ToNot(HaveOccurred())

			// Verify it's valid base64
			decoded, err := base64.StdEncoding.DecodeString(output)
			g.Expect(err).ToNot(HaveOccurred())

			// Verify it contains PEM certificate
			g.Expect(string(decoded)).To(ContainSubstring("BEGIN CERTIFICATE"))
			g.Expect(len(decoded)).To(BeNumerically(">", 100))
		}, "60s", "1s").Should(Succeed(), "CA bundle should be injected into validating webhook by cert-manager")

		// Verify CA bundle is injected into mutating webhook configuration
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(),
				"get", "mutatingwebhookconfiguration", "kuma-admission-mutating-webhook-configuration",
				"-o", "jsonpath={.webhooks[0].clientConfig.caBundle}",
			)
			g.Expect(err).ToNot(HaveOccurred())

			// Verify it's valid base64
			decoded, err := base64.StdEncoding.DecodeString(output)
			g.Expect(err).ToNot(HaveOccurred())

			// Verify it contains PEM certificate
			g.Expect(string(decoded)).To(ContainSubstring("BEGIN CERTIFICATE"))
			g.Expect(len(decoded)).To(BeNumerically(">", 100))
		}, "60s", "1s").Should(Succeed(), "CA bundle should be injected into mutating webhook by cert-manager")
	})

	It("should successfully validate resources using the webhook", func() {
		// Test that webhooks actually work by creating a Mesh resource
		meshYaml := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: test-mesh
spec:
  mtls:
    enabledBackend: ca-1
    backends:
    - name: ca-1
      type: builtin
`
		err := k8s.KubectlApplyFromStringE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(Config.KumaNamespace),
			meshYaml,
		)
		Expect(err).ToNot(HaveOccurred(), "Webhooks should successfully validate Mesh resource")

		// Clean up the test mesh
		err = k8s.KubectlDeleteFromStringE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(Config.KumaNamespace),
			meshYaml,
		)
		Expect(err).ToNot(HaveOccurred())
	})
}
