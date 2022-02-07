package inspect

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func Universal() {
	var cluster *UniversalCluster

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)

		Expect(Kuma(config_core.Standalone, WithVerbose())(cluster)).To(Succeed())
		Expect(cluster.VerifyKuma()).To(Succeed())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken)(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	// Before each test, verify the cluster is up and stable.
	JustBeforeEach(func() {
		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := cluster.GetKumactlOptions().KumactlList("dataplanes", "default")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(dataplanes).Should(ContainElements("demo-client"))
		}, "60s", "1s").Should(Succeed())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).ToNot(HaveOccurred())
	})

	It("should return envoy config_dump", func() {
		cmd := []string{"curl", "-v", "-m", "3", "--fail", "localhost:5681/meshes/default/dataplanes/demo-client/xds"}
		stdout, _, err := cluster.ExecWithRetries("", "", "kuma-cp", cmd...)
		Expect(err).ToNot(HaveOccurred())

		Expect(stdout).To(ContainSubstring(`"name": "kuma:envoy:admin"`))
		Expect(stdout).To(ContainSubstring(`"name": "outbound:127.0.0.1:4000"`))
		Expect(stdout).To(ContainSubstring(`"name": "outbound:127.0.0.1:4001"`))
		Expect(stdout).To(ContainSubstring(`"name": "outbound:127.0.0.1:5000"`))
		Expect(stdout).To(ContainSubstring(`"dataplane.resource": "{\"type\":\"Dataplane\",\"mesh\":\"default\",\"name\":\"demo-client\",\"creationTime\":\"0001-01-01T00:00:00Z\",\"modificationTime\":\"0001-01-01T00:00:00Z\",\"networking\":{\"address\":\"172.18.0.3\",\"inbound\":[{\"port\":13000,\"servicePort\":3000,\"tags\":{\"kuma.io/service\":\"demo-client\",\"team\":\"client-owners\"}}],\"outbound\":[{\"port\":4000,\"tags\":{\"kuma.io/service\":\"echo-server_kuma-test_svc_80\"}},{\"port\":4001,\"tags\":{\"kuma.io/service\":\"echo-server_kuma-test_svc_8080\"}},{\"port\":5000,\"tags\":{\"kuma.io/service\":\"external-service\"}}]}}",`))
	})
}
