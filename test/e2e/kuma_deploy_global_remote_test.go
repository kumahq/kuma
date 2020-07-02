package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kong/kuma/pkg/config/mode"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"

	"github.com/Kong/kuma/pkg/clusters/poller"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/test/framework"
)

var _ = Describe("Test Remote and Global", func() {
	namespaceWithSidecarInjection := func(namespace string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    kuma.io/sidecar-injection: "enabled"
`, namespace)
	}

	var clusters Clusters
	var c1, c2 Cluster
	var global, remote ControlPlane

	BeforeEach(func() {
		var err error
		clusters, err = NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		c1 = clusters.GetCluster(Kuma1)
		c2 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(Kuma(mode.Global)).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())

		global = c1.GetKuma()
		Expect(global).ToNot(BeNil())

		err = NewClusterSetup().
			Install(Kuma(mode.Remote)).
			Install(KumaDNS()).
			Install(Yaml(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s()).
			Install(EchoServerK8s()).
			Setup(c2)
		Expect(err).ToNot(HaveOccurred())

		remote = c2.GetKuma()
		Expect(remote).ToNot(BeNil())

		// when
		err = c1.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c2.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		err = global.AddCluster(remote.GetName(),
			global.GetKDSServerAddress(), remote.GetKDSServerAddress(), remote.GetIngressAddress())
		Expect(err).ToNot(HaveOccurred())

		err = c1.RestartKuma()
		Expect(err).ToNot(HaveOccurred())

		// then
		logs1, err := global.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs1).To(ContainSubstring("\"mode\":\"global\""))

		// and
		logs2, err := remote.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs2).To(ContainSubstring("\"mode\":\"remote\""))

	})

	AfterEach(func() {
		err := c2.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()
	})

	It("Should deploy Remote and Global on 2 clusters", func() {
		// when
		status, response := http_helper.HttpGet(c1.GetTesting(), global.GetGlobaStatusAPI(), nil)
		// then
		Expect(status).To(Equal(http.StatusOK))

		// when
		clustersStatus := poller.Clusters{}
		_ = json.Unmarshal([]byte(response), &clustersStatus)

		// then
		found := false
		for _, cluster := range clustersStatus {
			if cluster.URL == remote.GetKDSServerAddress() {
				Expect(cluster.Active).To(BeTrue())
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())

	})

	It("should deploy Remote and Global on 2 clusters and sync dataplanes", func() {
		// given
		namespace := func(namespace string) string {
			return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, namespace)
		}
		dp := func(cluster, namespace, name string) string {
			return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Dataplane
mesh: default
metadata:
  name: %s
  namespace: %s
spec:
  networking:
    address: 192.168.0.1
    inbound:
      - port: 12343
        tags:
          service: backend
          zone: %s
    outbound:
      - port: 1212
        tags:
          service: web
`, name, namespace, cluster)
		}

		// when
		err := Yaml(namespace("custom-ns"))(c2)
		Expect(err).ToNot(HaveOccurred())
		err = Yaml(dp("kuma-2-remote", "custom-ns", "dp-1"))(c2)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() string {
			output, err := k8s.RunKubectlAndGetOutputE(c1.GetTesting(), c1.GetKubectlOptions("default"), "get", "dataplanes")
			Expect(err).ToNot(HaveOccurred())
			return output
		}, "5s", "500ms").Should(ContainSubstring("kuma-2-remote.dp-1.custom-ns"))
	})
})
