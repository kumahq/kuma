package k8s_api_bypass

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func K8sApiBypass() {
	meshName := "k8s-api-bypass"
	namespace := "k8s-api-bypass"

	meshDefaultMtlsOn := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: k8s-api-bypass
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

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "true"))).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(meshName, namespace)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to communicate with API Server", func() {
		apiServer := "https://kubernetes.default.svc"
		serviceAccount := "/var/run/secrets/kubernetes.io/serviceaccount"
		token := fmt.Sprintf("$(cat %s/token)", serviceAccount)
		caCert := fmt.Sprintf("%s/ca.crt", serviceAccount)
		url := fmt.Sprintf("%s/api", apiServer)

		// given Mesh with passthrough enabled then communication with API Server works
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectRawResponse(
				kubernetes.Cluster, "demo-client", url,
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithCACert(caCert),
				client.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Equal("200"))
		}).Should(Succeed())

		// when passthrough is disabled on the Mesh
		err := kubernetes.Cluster.Install(YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "false")))
		Expect(err).ToNot(HaveOccurred())

		// then communication with API Server still works
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectRawResponse(
				kubernetes.Cluster, "demo-client", url,
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithCACert(caCert),
				client.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Equal("200"))
		}).Should(Succeed())
	})
}
