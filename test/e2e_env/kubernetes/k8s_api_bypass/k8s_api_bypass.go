package k8s_api_bypass

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
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
	var clientPodName string

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "true"))).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(kubernetes.Cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to communicate with API Server", func() {
		serviceAccount := "/var/run/secrets/kubernetes.io/serviceaccount"
		caCert := fmt.Sprintf("%s/ca.crt", serviceAccount)

		// read service account token
		var token string
		Eventually(func(g Gomega) {
			stdout, _, err := kubernetes.Cluster.Exec(
				namespace, clientPodName, "demo-client",
				"cat", fmt.Sprintf("%s/token", serviceAccount),
			)
			g.Expect(err).ToNot(HaveOccurred())
			token = stdout
		}).Should(Succeed())

		// given Mesh with passthrough enabled then communication with API Server works
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "demo-client", "https://kubernetes.default.svc/api",
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithCACert(caCert),
				client.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)),
			)
			g.Expect(err).ToNot(HaveOccurred())
			// we expect k8s resource 'meta/v1, Kind=APIVersions'
			g.Expect(stdout).To(ContainSubstring(`"kind": "APIVersions"`))
		}).Should(Succeed())

		// when passthrough is disabled on the Mesh
		err := kubernetes.Cluster.Install(YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "false")))
		Expect(err).ToNot(HaveOccurred())

		// then communication with API Server still works
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "demo-client", "https://kubernetes.default.svc/api",
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithCACert(caCert),
				client.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token)),
			)
			g.Expect(err).ToNot(HaveOccurred())
			// we expect k8s resource 'meta/v1, Kind=APIVersions'
			g.Expect(stdout).To(ContainSubstring(`"kind": "APIVersions"`))
		}).Should(Succeed())
	})
}
