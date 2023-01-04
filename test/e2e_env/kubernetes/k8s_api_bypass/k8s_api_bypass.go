package k8s_api_bypass

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
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
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(meshName, namespace)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(env.Cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to communicate with API Server", func() {
		apiServer := "https://kubernetes.default.svc"
		serviceAccount := "/var/run/secrets/kubernetes.io/serviceaccount"
		token := fmt.Sprintf("$(cat %s/token)", serviceAccount)
		caCert := fmt.Sprintf("%s/ca.crt", serviceAccount)
		header := fmt.Sprintf("Authorization: Bearer %s", token)
		url := fmt.Sprintf("%s/api", apiServer)
		cmd := fmt.Sprintf("curl -s --cacert %s --header %q -o /dev/null -w '%%{http_code}\\n' -m 3 --fail %s", caCert, header, url)

		// given Mesh with passthrough enabled then communication with API Server works
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec(namespace, clientPodName, "demo-client", "bash", "-c", cmd)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Equal("200"))
		}).Should(Succeed())

		// when passthrough is disabled on the Mesh
		err := env.Cluster.Install(YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "false")))
		Expect(err).ToNot(HaveOccurred())

		// then communication with API Server still works
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec(namespace, clientPodName, "demo-client", "bash", "-c", cmd)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Equal("200"))
		}).Should(Succeed())
	})
}
