package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"

	"github.com/Kong/kuma/pkg/clusters/poller"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config/core"
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

	var c1, c2 Cluster
	var global, remote ControlPlane
	var nodeIP string

	BeforeEach(func() {
		clusters, err := NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		c1 = clusters.GetCluster(Kuma1)
		c2 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())

		global = c1.GetKuma()
		Expect(global).ToNot(BeNil())

		err = NewClusterSetup().
			Install(Kuma(core.Remote)).
			Install(KumaDNS()).
			Install(Yaml(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClient()).
			Install(EchoServer()).
			Setup(c2)
		Expect(err).ToNot(HaveOccurred())

		remote = c2.GetKuma()
		Expect(remote).ToNot(BeNil())

		clientset, err := k8s.GetKubernetesClientFromOptionsE(c1.GetTesting(), c1.GetKubectlOptions())
		Expect(err).ToNot(HaveOccurred())

		nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		Expect(err).ToNot(HaveOccurred())
		Expect(nodeList.Items).To(HaveLen(1)) // our strategy of getting IP based on the fact that cluster has single node

		for _, addr := range nodeList.Items[0].Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				nodeIP = addr.Address
			}
		}
		Expect(nodeIP).ToNot(BeEmpty())
	})

	AfterEach(func() {
		_ = c1.DeleteKuma()
		_ = c2.DeleteKuma()
		_ = k8s.KubectlDeleteFromStringE(c2.GetTesting(), c2.GetKubectlOptions(), namespaceWithSidecarInjection(TestNamespace))
	})

	It("Should deploy Local and Global on 2 clusters", func() {
		err := global.AddCluster(remote.GetName(), remote.GetHostAPI(), fmt.Sprintf("http://%s:%d", nodeIP, LocalCPSyncNodePort))
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
			if cluster.URL == remote.GetHostAPI() {
				Expect(cluster.Active).To(BeTrue())
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())

	})

	It("should deploy Remote and Global on 2 clusters and sync dataplanes", func() {
		err := global.AddCluster(remote.GetName(), remote.GetHostAPI(), fmt.Sprintf("http://%s:%d", nodeIP, LocalCPSyncNodePort))
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
          cluster: %s
    outbound:
      - port: 1212
        tags:
          service: web
`, name, namespace, cluster)
		}

		err = Yaml(namespace("custom-ns"))(c2)
		Expect(err).ToNot(HaveOccurred())
		err = Yaml(dp("kuma-2-remote", "custom-ns", "dp-1"))(c2)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() string {
			output, err := k8s.RunKubectlAndGetOutputE(c1.GetTesting(), c1.GetKubectlOptions("default"), "get", "dataplanes")
			Expect(err).ToNot(HaveOccurred())
			return output
		}, "5s", "500ms").Should(ContainSubstring("kuma-2-remote.dp-1.custom-ns"))
	})
})
