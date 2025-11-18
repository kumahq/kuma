package webhooks

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/kumahq/kuma/v2/pkg/config/core"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/certmanager"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	FIt("should inject CA bundle into Kuma webhook configurations", func() {
		kumaNamespace := Config.KumaNamespace

		// Verify Certificate resource exists
		Eventually(func(g Gomega) {
			_, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(kumaNamespace),
				"get", "certificate", "kuma-tls-cert",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "60s", "1s").Should(Succeed(), "Certificate resource should exist")

		// Verify Issuer resource exists
		Eventually(func(g Gomega) {
			_, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(kumaNamespace),
				"get", "issuer", "kuma-selfsigned-issuer",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "60s", "1s").Should(Succeed(), "Issuer resource should exist")

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
			// CA bundle should be non-empty base64-encoded certificate
			g.Expect(len(output)).To(BeNumerically(">", 100))
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
			// CA bundle should be non-empty base64-encoded certificate
			g.Expect(len(output)).To(BeNumerically(">", 100))
		}, "60s", "1s").Should(Succeed(), "CA bundle should be injected into mutating webhook by cert-manager")
	})
}

