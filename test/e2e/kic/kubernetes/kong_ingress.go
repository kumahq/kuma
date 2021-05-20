package kubernetes

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func KICKubernetes() {
	namespaceWithSidecarInjection := func(namespace string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  annotations:
    kuma.io/sidecar-injection: "enabled"
`, namespace)
	}
	var ingressNamespace string
	var defaultIngressNamespace = "kuma-gateway"
	var altIngressNamespace = "kuma-yawetag"
	var ingressApp = "ingress-kong"
	var kubernetes Cluster
	var kubernetesOps []DeployOptionsFunc
	E2EBeforeSuite(func() {
		k8sClusters, err := NewK8sClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())
		// Global
		kubernetes = k8sClusters.GetCluster(Kuma1)
		kubernetesOps = []DeployOptionsFunc{
			WithEnv("KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED", "true"),
		}
		err = NewClusterSetup().
			Install(Kuma(config_core.Standalone, kubernetesOps...)).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(EchoServerK8s("default")).
			Setup(kubernetes)
		Expect(err).ToNot(HaveOccurred())
		err = kubernetes.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

	})
	E2EAfterEach(func() {
		Expect(kubernetes.DeleteNamespace(ingressNamespace)).To(Succeed())
	})
	E2EAfterSuite(func() {
		Expect(kubernetes.DeleteKuma(kubernetesOps...)).To(Succeed())
		Expect(kubernetes.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(kubernetes.DismissCluster()).To(Succeed())
	})
	It("should install kong ingress into default namespace", func() {
		ingressNamespace = defaultIngressNamespace
		// given kong ingress
		output, err := kubernetes.GetKumactlOptions().RunKumactlAndGetOutputV(Verbose, "install", "gateway", "kong")
		Expect(err).ToNot(HaveOccurred())
		err = NewClusterSetup().Install(YamlK8s(output)).Setup(kubernetes)
		Expect(err).ToNot(HaveOccurred())

		// Wait for ingress
		err = NewClusterSetup().Install(Combine(
			//WaitService(ingressNamespace, "kong-proxy"),
			WaitNumPodsNamespace(ingressNamespace, 1, ingressApp),
			WaitPodsAvailable(ingressNamespace, ingressApp))).Setup(kubernetes)
		Expect(err).ToNot(HaveOccurred())

		// Test connection to echo server through ingress "proxy" pod
		pods, err := k8s.ListPodsE(
			kubernetes.GetTesting(),
			kubernetes.GetKubectlOptions(ingressNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", ingressApp),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))
		clientPod := &pods[0]

		_, _, err = kubernetes.ExecWithRetries(ingressNamespace, clientPod.GetName(), "proxy",
			"wget", "-O-", "echo-server_kuma-test_svc_80.mesh")
		Expect(err).ToNot(HaveOccurred())
	})
	It("should install kong ingress into non-default namespace", func() {
		ingressNamespace = altIngressNamespace
		// given kong ingress
		output, err := kubernetes.GetKumactlOptions().RunKumactlAndGetOutputV(Verbose, "install", "gateway", "kong", "--namespace", ingressNamespace)
		Expect(err).ToNot(HaveOccurred())
		err = NewClusterSetup().Install(YamlK8s(output)).Setup(kubernetes)
		Expect(err).ToNot(HaveOccurred())

		// Wait for ingress
		err = NewClusterSetup().Install(Combine(
			//WaitService(ingressNamespace, "kong-proxy"),
			WaitNumPodsNamespace(ingressNamespace, 1, ingressApp),
			WaitPodsAvailable(ingressNamespace, ingressApp))).Setup(kubernetes)
		Expect(err).ToNot(HaveOccurred())

		// Test connection to echo server through ingress "proxy" pod
		pods, err := k8s.ListPodsE(
			kubernetes.GetTesting(),
			kubernetes.GetKubectlOptions(ingressNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", ingressApp),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))
		clientPod := &pods[0]

		_, _, err = kubernetes.ExecWithRetries(ingressNamespace, clientPod.GetName(), "proxy",
			"wget", "-O-", "echo-server_kuma-test_svc_80.mesh")
		Expect(err).ToNot(HaveOccurred())
	})

}
