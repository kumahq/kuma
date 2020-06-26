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
	var clusters Clusters

	BeforeEach(func() {
		var err error
		clusters, err = NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.CreateNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.LabelNamespaceForSidecarInjection(TestNamespace)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := clusters.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()
	})

	// todo (lobkoviya): implement Cluster polling based on KDS
	XIt("Should deploy Remote and Global on 2 clusters", func() {
		// given
		c1 := clusters.GetCluster(Kuma1)
		c2 := clusters.GetCluster(Kuma2)

		global, err := c1.DeployKuma(core.Global)
		Expect(err).ToNot(HaveOccurred())

		remote, err := c2.DeployKuma(core.Remote)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c1.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c2.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		err = global.AddCluster(remote.GetName(), remote.GetHostAPI(), "") // todo (lobkovilya): pass Node IP as lbAddress
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
		// given
		c1 := clusters.GetCluster(Kuma1)
		c2 := clusters.GetCluster(Kuma2)

		global, err := c1.DeployKuma(core.Global)
		Expect(err).ToNot(HaveOccurred())

		local, err := c2.DeployKuma(core.Remote)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c1.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c2.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		clientset, err := k8s.GetKubernetesClientFromOptionsE(c1.GetTesting(), c1.GetKubectlOptions())
		Expect(err).ToNot(HaveOccurred())

		nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		Expect(err).ToNot(HaveOccurred())
		Expect(nodeList.Items).To(HaveLen(1)) // our strategy of getting IP based on the fact that cluster has single node
		var nodeIP string
		for _, addr := range nodeList.Items[0].Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				nodeIP = addr.Address
			}
		}
		Expect(nodeIP).ToNot(BeEmpty())
		err = global.AddCluster(local.GetName(), local.GetHostAPI(), fmt.Sprintf("http://%s:%d", nodeIP, LocalCPSyncNodePort))
		Expect(err).ToNot(HaveOccurred())

		err = c1.RestartKuma()
		Expect(err).ToNot(HaveOccurred())

		// then
		logs1, err := global.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs1).To(ContainSubstring("\"mode\":\"global\""))

		// and
		logs2, err := local.GetKumaCPLogs()
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
