package k8s_api_bypass

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

	var cluster *K8sCluster
	var clientPodName string

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default", TestNamespace)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "true"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		clientPodName, err = PodNameOfApp(cluster, "demo-client", TestNamespace)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		err := cluster.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DeleteKuma()
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

		cmd := fmt.Sprintf("curl -s --cacert %s --header %q -o /dev/null -w '%%{http_code}\\n' -m 3 --fail %s", caCert, header, url)

		// given Mesh with passthrough enabled then communication with API Server works
		stdout, _, err := cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
			"bash", "-c", cmd)
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(Equal("200"))

		// when passthrough is disabled on the Mesh
		err = YamlK8s(fmt.Sprintf(meshDefaultMtlsOn, "false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then communication with API Server still works
		stdout, _, err = cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
			"bash", "-c", cmd)
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(Equal("200"))
	})
}
