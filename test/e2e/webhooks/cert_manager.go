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
	var releaseName string

	BeforeAll(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		const certManagerNamespace = "cert-manager"

		releaseName = fmt.Sprintf(
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
		Expect(cluster.DeleteDeployment(certmanager.DeploymentName)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should have certificate with all required DNS SANs", func() {
		kumaNamespace := Config.KumaNamespace
		serviceName := releaseName + "-control-plane"

		// Verify Certificate has all required DNS names
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(kumaNamespace),
				"get", "certificate", "kuma-tls-cert",
				"-o", "jsonpath={.spec.dnsNames}",
			)
			g.Expect(err).ToNot(HaveOccurred())

			// Expected SANs:
			// 1. Short hostname: <service>.<namespace>
			// 2. With .svc: <service>.<namespace>.svc
			// 3. FQDN: <service>.<namespace>.svc.cluster.local
			expectedShortHostname := fmt.Sprintf("%s.%s", serviceName, kumaNamespace)
			expectedSvcHostname := fmt.Sprintf("%s.%s.svc", serviceName, kumaNamespace)
			expectedFQDN := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, kumaNamespace)

			g.Expect(output).To(ContainSubstring(expectedShortHostname))
			g.Expect(output).To(ContainSubstring(expectedSvcHostname))
			g.Expect(output).To(ContainSubstring(expectedFQDN))
		}, "30s", "1s").Should(Succeed(), "Certificate should have all required DNS SANs including short hostname")
	})

	It("should allow dataplane to connect to control plane", func() {
		const namespace = "cert-manager-dp-test"
		const mesh = "default"

		// Create namespace with sidecar injection enabled
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		// Deploy a simple workload that will get sidecar injected
		deployment := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  namespace: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
        kuma.io/mesh: %s
    spec:
      containers:
      - name: test-app
        image: busybox:1.36
        command: ["sleep", "infinity"]
`
		err = k8s.KubectlApplyFromStringE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(namespace),
			fmt.Sprintf(deployment, namespace, mesh),
		)
		Expect(err).ToNot(HaveOccurred())

		// Wait for pod to be ready (sidecar injected and running)
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(namespace),
				"get", "pods", "-l", "app=test-app",
				"-o", "jsonpath={.items[0].status.containerStatuses[*].ready}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			// Both containers (app + sidecar) should be ready
			g.Expect(output).To(ContainSubstring("true"))
		}, "120s", "1s").Should(Succeed(), "Pod with sidecar should become ready")

		// Verify dataplane is online (connected to control plane)
		Eventually(func(g Gomega) {
			online, found, err := IsDataplaneOnline(cluster, mesh, "test-app")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(found).To(BeTrue(), "Dataplane should be found")
			g.Expect(online).To(BeTrue(), "Dataplane should be online")
		}, "60s", "1s").Should(Succeed(), "Dataplane should connect to control plane using cert-manager certificate")

		// Cleanup
		err = k8s.KubectlDeleteFromStringE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(namespace),
			fmt.Sprintf(deployment, namespace, mesh),
		)
		Expect(err).ToNot(HaveOccurred())

		_, _ = k8s.RunKubectlAndGetOutputE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(),
			"delete", "namespace", namespace,
		)
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

		// Verify CA bundle is injected into all validating webhooks
		verifyWebhookCABundle(cluster, "validatingwebhookconfiguration", "kuma-validating-webhook-configuration")

		// Verify CA bundle is injected into all mutating webhooks
		verifyWebhookCABundle(cluster, "mutatingwebhookconfiguration", "kuma-admission-mutating-webhook-configuration")
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

	It("should continue working after certificate rotation", func() {
		kumaNamespace := Config.KumaNamespace

		// Get the current certificate's serial number
		var originalSerial string
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(kumaNamespace),
				"get", "secret", "kuma-tls-cert",
				"-o", "jsonpath={.data.tls\\.crt}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			originalSerial = output
		}, "10s", "1s").Should(Succeed())

		// Force certificate rotation by deleting the secret
		// cert-manager will automatically recreate it
		_, err := k8s.RunKubectlAndGetOutputE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(kumaNamespace),
			"delete", "secret", "kuma-tls-cert",
		)
		Expect(err).ToNot(HaveOccurred())

		// Wait for cert-manager to recreate the certificate
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(kumaNamespace),
				"get", "certificate", "kuma-tls-cert",
				"-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).To(Equal("True"))
		}, "60s", "1s").Should(Succeed(), "Certificate should be recreated and ready after rotation")

		// Verify the certificate has changed (new serial number)
		Eventually(func(g Gomega) {
			output, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(kumaNamespace),
				"get", "secret", "kuma-tls-cert",
				"-o", "jsonpath={.data.tls\\.crt}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).ToNot(Equal(originalSerial), "Certificate should have been rotated")
		}, "60s", "1s").Should(Succeed())

		// Verify CA bundle was re-injected into all webhooks after rotation
		verifyWebhookCABundle(cluster, "validatingwebhookconfiguration", "kuma-validating-webhook-configuration")
		verifyWebhookCABundle(cluster, "mutatingwebhookconfiguration", "kuma-admission-mutating-webhook-configuration")

		// Verify webhooks still work after rotation by creating a test resource
		meshYaml := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: test-mesh-rotation
spec:
  mtls:
    enabledBackend: ca-1
    backends:
    - name: ca-1
      type: builtin
`
		err = k8s.KubectlApplyFromStringE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(Config.KumaNamespace),
			meshYaml,
		)
		Expect(err).ToNot(HaveOccurred(), "Webhooks should work after certificate rotation")

		// Clean up
		err = k8s.KubectlDeleteFromStringE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(Config.KumaNamespace),
			meshYaml,
		)
		Expect(err).ToNot(HaveOccurred())
	})
}

// verifyWebhookCABundle verifies that CA bundle is properly injected into all webhooks
func verifyWebhookCABundle(cluster Cluster, webhookType, webhookName string) {
	Eventually(func(g Gomega) {
		// Get the number of webhooks
		output, err := k8s.RunKubectlAndGetOutputE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(),
			"get", webhookType, webhookName,
			"-o", "jsonpath={.webhooks[*].name}",
		)
		g.Expect(err).ToNot(HaveOccurred())
		webhooks := strings.Fields(output)
		g.Expect(webhooks).ToNot(BeEmpty(), "Should have at least one webhook")

		// Verify CA bundle for each webhook
		for i := range webhooks {
			caBundleOutput, err := k8s.RunKubectlAndGetOutputE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(),
				"get", webhookType, webhookName,
				"-o", fmt.Sprintf("jsonpath={.webhooks[%d].clientConfig.caBundle}", i),
			)
			g.Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to get CA bundle for webhook %d", i))

			// Verify it's valid base64
			decoded, err := base64.StdEncoding.DecodeString(caBundleOutput)
			g.Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("CA bundle for webhook %d should be valid base64", i))

			// Verify it contains PEM certificate
			g.Expect(string(decoded)).To(ContainSubstring("BEGIN CERTIFICATE"), fmt.Sprintf("Webhook %d should have PEM certificate", i))
			g.Expect(len(decoded)).To(BeNumerically(">", 100), fmt.Sprintf("Webhook %d CA bundle should be substantial", i))
		}
	}, "60s", "1s").Should(Succeed(), fmt.Sprintf("CA bundle should be injected into all webhooks in %s/%s", webhookType, webhookName))
}
