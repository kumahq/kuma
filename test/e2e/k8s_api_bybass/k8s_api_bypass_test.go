package k8s_api_bybass_test

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/kumahq/kuma/test/framework"
)


var _ = Describe("Test Kubernetes API Bypass", func() {
	meshDefaultMtlsOn := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
  networking:
    outbound:
      passthrough: %s
`

	externalService := `
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: default
metadata:
  name: external-service-%s
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
    id: "%s"
  networking:
    address: %s:%d
    tls:
      enabled: %s
`
	es1 := "1"

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

	var cluster *K8sCluster
	var clientPod *v1.Pod
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		c, err := NewK8SCluster(NewTestingT(), Kuma1, Silent)
		Expect(err).ToNot(HaveOccurred())

		cluster = c.(*K8sCluster)

		// Global
		deployOptsFuncs = []DeployOptionsFunc{
			WithEnv("KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED", "true"),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s("default")).
			Install(externalservice.Install(externalservice.HttpServer, []string{})).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(fmt.Sprintf(externalService,
			es1, es1,
			"externalservice-http-server.externalservice-namespace.svc.cluster.local", 10080,
			"false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		pods, err := k8s.ListPodsE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod = &pods[0]
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}

		err := cluster.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DeleteKuma(deployOptsFuncs...)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should be able to communicate with API Server", func() {
		apiServer := "https://kubernetes.default.svc"
		serviceAccount := "/var/run/secrets/kubernetes.io/serviceaccount"
		token := fmt.Sprintf("$(cat %s/token)", serviceAccount)
		caCert := fmt.Sprintf("%s/ca.crt", serviceAccount)
		header := fmt.Sprintf("Authorization: Bearer %s", token)
		url := fmt.Sprintf("%s/api", apiServer)

		cmd := fmt.Sprintf("curl --cacert %s --header %q -X GET -v -m 3 --fail %s", caCert, header, url)

		// given Mesh with passthrough enabled
		err := YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "true"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication with API Server works
		_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"bash", "-c", cmd)
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/2 200"))

		// the same as overall communication outside the mesh
		_, stderr, err = cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// when passthrough is disabled on the Mesh
		err = YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication with API Server still works
		_, stderr, err = cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"bash", "-c", cmd)
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/2 200"))

		// but accessing the external service is no longer possible
		_, err = retry.DoWithRetryE(cluster.GetTesting(), "passthrough access to service", 5, DefaultTimeout, func() (string, error) {
			_, _, err := cluster.Exec("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
			if err != nil {
				return "", err
			}

			return "", nil
		})
		Expect(err).To(HaveOccurred())

		// when apply external service
		err = YamlK8s(fmt.Sprintf(externalService,
			es1, es1,
			"externalservice-http-server.externalservice-namespace.svc.cluster.local", 10080, // .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
			"false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then you can access external service again
		stdout, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://externalservice-http-server.externalservice-namespace:10080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("externalservice-https-server"))
	})
})
