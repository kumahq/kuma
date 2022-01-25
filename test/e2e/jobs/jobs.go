package jobs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func Jobs() {
	var kubernetes Cluster

	E2EBeforeSuite(func() {
		k8sClusters, err := NewK8sClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		kubernetes = k8sClusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(config_core.Standalone)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(testserver.Install()).
			Setup(kubernetes)
		Expect(err).ToNot(HaveOccurred())
		err = kubernetes.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterSuite(func() {
		Expect(kubernetes.DeleteKuma()).To(Succeed())
		Expect(kubernetes.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(kubernetes.DismissCluster()).To(Succeed())
	})

	It("should terminate jobs without mTLS", func() {
		// when
		err := DemoClientJobK8s(TestNamespace, model.DefaultMesh, "test-server_kuma-test_svc_80.mesh")(kubernetes)

		// then CP terminates the job by sending /quitquitquit to Envoy Admin and verifies connection using self-signed certs
		Expect(err).ToNot(HaveOccurred())
	})

	It("should terminate jobs with mTLS", func() {
		// given mTLS in the Mesh
		meshDefaulMtlsOn := `
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
`
		err := YamlK8s(meshDefaulMtlsOn)(kubernetes)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = DemoClientJobK8s(TestNamespace, model.DefaultMesh, "test-server_kuma-test_svc_80.mesh")(kubernetes)

		// then CP terminates the job by sending /quitquitquit to Envoy Admin and verifies connection using mTLS certs
		Expect(err).ToNot(HaveOccurred())
	})
}
