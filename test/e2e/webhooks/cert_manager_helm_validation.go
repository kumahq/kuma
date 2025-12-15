package webhooks

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/helm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/config/core"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/certmanager"
)

func CertManagerHelmValidation() {
	var cluster Cluster

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
	})

	AfterEach(func() {
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should reject cert-manager enabled with secretName set", func() {
		const certManagerNamespace = "cert-manager"
		releaseName := fmt.Sprintf("kuma-test-%d", GinkgoRandomSeed())

		err := NewClusterSetup().
			Install(certmanager.Install(
				certmanager.WithNamespace(certManagerNamespace),
			)).
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithHelmOpt("controlPlane.tls.general.certManager.enabled", "true"),
				WithHelmOpt("controlPlane.tls.general.secretName", "my-custom-secret"),
			)).
			Setup(cluster)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("secretName and caBundle must be empty"))
	})

	It("should reject cert-manager enabled with caBundle set", func() {
		const certManagerNamespace = "cert-manager"
		releaseName := fmt.Sprintf("kuma-test-%d", GinkgoRandomSeed())

		err := NewClusterSetup().
			Install(certmanager.Install(
				certmanager.WithNamespace(certManagerNamespace),
			)).
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithHelmOpt("controlPlane.tls.general.certManager.enabled", "true"),
				WithHelmOpt("controlPlane.tls.general.caBundle", "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t"),
			)).
			Setup(cluster)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("secretName and caBundle must be empty"))
	})

	It("should reject cert-manager enabled with both secretName and caBundle set", func() {
		const certManagerNamespace = "cert-manager"
		releaseName := fmt.Sprintf("kuma-test-%d", GinkgoRandomSeed())

		err := NewClusterSetup().
			Install(certmanager.Install(
				certmanager.WithNamespace(certManagerNamespace),
			)).
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithHelmOpt("controlPlane.tls.general.certManager.enabled", "true"),
				WithHelmOpt("controlPlane.tls.general.secretName", "my-custom-secret"),
				WithHelmOpt("controlPlane.tls.general.caBundle", "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t"),
			)).
			Setup(cluster)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("secretName and caBundle must be empty"))
	})

	It("should reject cert-manager enabled without cert-manager CRDs installed", func() {
		// Don't install cert-manager, just try to install Kuma with cert-manager enabled
		releaseName := fmt.Sprintf("kuma-test-%d", GinkgoRandomSeed())

		err := NewClusterSetup().
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithHelmOpt("controlPlane.tls.general.certManager.enabled", "true"),
			)).
			Setup(cluster)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("cert-manager CRDs not found"))
	})

	It("should accept secretName without caBundle", func() {
		releaseName := fmt.Sprintf("kuma-test-%d", GinkgoRandomSeed())

		// First, create the secret that we'll reference
		opts := &helm.Options{
			KubectlOptions: cluster.GetKubectlOptions(Config.KumaNamespace),
		}
		_, err := helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(),
			opts,
			"install", releaseName,
			"--create-namespace",
			"--namespace", Config.KumaNamespace,
			"--set", "controlPlane.mode=zone",
			"--set", "controlPlane.tls.general.secretName=my-custom-secret",
			"--dry-run",
			"deployments/charts/kuma")

		// This should not produce an error during template rendering
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reject only caBundle without secretName", func() {
		releaseName := fmt.Sprintf("kuma-test-%d", GinkgoRandomSeed())

		opts := &helm.Options{
			KubectlOptions: cluster.GetKubectlOptions(Config.KumaNamespace),
		}
		_, err := helm.RunHelmCommandAndGetStdOutE(cluster.GetTesting(),
			opts,
			"install", releaseName,
			"--create-namespace",
			"--namespace", Config.KumaNamespace,
			"--set", "controlPlane.mode=zone",
			"--set", "controlPlane.tls.general.caBundle=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t",
			"--dry-run",
			"deployments/charts/kuma")

		Expect(err).To(HaveOccurred())
		errOutput := strings.ToLower(err.Error())
		Expect(errOutput).To(Or(
			ContainSubstring("both or neither"),
			ContainSubstring("secretname"),
		))
	})
}
