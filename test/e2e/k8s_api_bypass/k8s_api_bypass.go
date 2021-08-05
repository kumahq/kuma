package k8s_api_bypass

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func K8sApiBypass() {
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
	var deployOptsFuncs = KumaK8sDeployOpts

	BeforeEach(func() {
		c, err := NewK8SCluster(NewTestingT(), Kuma1, Silent)
		Expect(err).ToNot(HaveOccurred())

		cluster = c.(*K8sCluster)

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s("default")).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "true"))(cluster)
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

		// given Mesh with passthrough enabled then communication with API Server works
		_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"bash", "-c", cmd)
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/2 200"))

		// when passthrough is disabled on the Mesh
		err = YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication with API Server still works
		_, stderr, err = cluster.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"bash", "-c", cmd)
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/2 200"))
	})
}
