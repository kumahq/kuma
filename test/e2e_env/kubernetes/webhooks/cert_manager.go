package webhooks

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/certmanager"
	"github.com/kumahq/kuma/v2/test/framework/envs/kubernetes"
)

func CertManagerCAInjection() {
	const certManagerNamespace = "cert-manager"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(certmanager.Install(
				certmanager.WithNamespace(certManagerNamespace),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(certManagerNamespace)).To(Succeed())
	})

	It("should inject CA bundle into Kuma webhook configurations", func() {
		// Verify cert-manager pods are running
		pods := k8s.ListPods(kubernetes.Cluster.GetTesting(),
			kubernetes.Cluster.GetKubectlOptions(certManagerNamespace),
			metav1.ListOptions{},
		)
		Expect(len(pods)).To(BeNumerically(">=", 3)) // cert-manager, cainjector, webhook

		const kumaNamespace = "kuma-system"

		// Deploy Kuma control plane with cert-manager enabled
		err := NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithCtlOpts(map[string]string{
					"--set": "controlPlane.tls.general.certManager.enabled=true",
				}),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		defer func() {
			// Cleanup: uninstall Kuma
			err := kubernetes.Cluster.DeleteKuma()
			Expect(err).ToNot(HaveOccurred())
		}()

		// Verify Certificate resource exists
		Eventually(func() error {
			_, err := k8s.RunKubectlAndGetOutputE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(kumaNamespace),
				"get", "certificate", "kuma-tls-cert",
			)
			return err
		}, "60s", "1s").Should(Succeed(), "Certificate resource should exist")

		// Verify Issuer resource exists
		Eventually(func() error {
			_, err := k8s.RunKubectlAndGetOutputE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(kumaNamespace),
				"get", "issuer", "kuma-selfsigned-issuer",
			)
			return err
		}, "60s", "1s").Should(Succeed(), "Issuer resource should exist")

		// Verify cert-manager annotation on validating webhook configuration
		Eventually(func() string {
			output, err := k8s.RunKubectlAndGetOutputE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(),
				"get", "validatingwebhookconfiguration", "kuma-validating-webhook-configuration",
				"-o", "jsonpath={.metadata.annotations.cert-manager\\.io/inject-ca-from}",
			)
			if err != nil {
				return ""
			}
			return output
		}, "60s", "1s").Should(Equal(fmt.Sprintf("%s/kuma-tls-cert", kumaNamespace)), "Validating webhook should have cert-manager annotation")

		// Verify cert-manager annotation on mutating webhook configuration
		Eventually(func() string {
			output, err := k8s.RunKubectlAndGetOutputE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(),
				"get", "mutatingwebhookconfiguration", "kuma-admission-mutating-webhook-configuration",
				"-o", "jsonpath={.metadata.annotations.cert-manager\\.io/inject-ca-from}",
			)
			if err != nil {
				return ""
			}
			return output
		}, "60s", "1s").Should(Equal(fmt.Sprintf("%s/kuma-tls-cert", kumaNamespace)), "Mutating webhook should have cert-manager annotation")

		// Verify CA bundle is injected into validating webhook configuration
		Eventually(func() bool {
			output, err := k8s.RunKubectlAndGetOutputE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(),
				"get", "validatingwebhookconfiguration", "kuma-validating-webhook-configuration",
				"-o", "jsonpath={.webhooks[0].clientConfig.caBundle}",
			)
			if err != nil {
				return false
			}
			// CA bundle should be non-empty base64-encoded certificate
			return len(output) > 100
		}, "60s", "1s").Should(BeTrue(), "CA bundle should be injected into validating webhook by cert-manager")

		// Verify CA bundle is injected into mutating webhook configuration
		Eventually(func() bool {
			output, err := k8s.RunKubectlAndGetOutputE(
				kubernetes.Cluster.GetTesting(),
				kubernetes.Cluster.GetKubectlOptions(),
				"get", "mutatingwebhookconfiguration", "kuma-admission-mutating-webhook-configuration",
				"-o", "jsonpath={.webhooks[0].clientConfig.caBundle}",
			)
			if err != nil {
				return false
			}
			// CA bundle should be non-empty base64-encoded certificate
			return len(output) > 100
		}, "60s", "1s").Should(BeTrue(), "CA bundle should be injected into mutating webhook by cert-manager")
	})
}
