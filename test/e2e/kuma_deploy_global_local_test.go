package e2e_test

import (
	"encoding/json"
	"net/http"

	"github.com/Kong/kuma/pkg/clusters/poller"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config/core"
	. "github.com/Kong/kuma/test/framework"
)

var _ = Describe("Test Local and Global", func() {
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

	It("Should deploy Local and Global on 2 clusters", func() {
		// given
		c1 := clusters.GetCluster(Kuma1)
		c2 := clusters.GetCluster(Kuma2)

		global, err := c1.DeployKuma(core.Global)
		Expect(err).ToNot(HaveOccurred())

		local, err := c2.DeployKuma(core.Local)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c1.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c2.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		err = global.AddCluster(local.GetName(), local.GetHostAPI())
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
		Expect(logs2).To(ContainSubstring("\"mode\":\"local\""))

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
			if cluster.URL == local.GetHostAPI() {
				Expect(cluster.Active).To(BeTrue())
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())

	})
})
