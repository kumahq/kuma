package e2e_test

import (
	"encoding/json"
	"net/http"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config/core"
	"github.com/Kong/kuma/pkg/globalcp"
	"github.com/Kong/kuma/test/framework"
)

var _ = Describe("Test Local and Global", func() {
	var clusters framework.Clusters

	BeforeEach(func() {
		var err error
		clusters, err = framework.NewK8sClusters(
			[]string{framework.Kuma1, framework.Kuma2},
			framework.Verbose)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.CreateNamespace("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		err = clusters.LabelNamespaceForSidecarInjection("kuma-test")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := clusters.DeleteNamespace("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()
	})

	It("Should deploy Local and Global on 2 clusters", func() {
		// given
		c1 := clusters.GetCluster(framework.Kuma1)
		c2 := clusters.GetCluster(framework.Kuma2)

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

		err = global.AddLocalCP(local.GetName(), local.GetHostAPI())
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
		localCPMap := globalcp.LocalCPMap{}
		_ = json.Unmarshal([]byte(response), &localCPMap)
		// then
		Expect(localCPMap).To(HaveKey(local.GetName()))
		// and
		Expect(localCPMap[local.GetName()].Active).To(BeTrue())


	})
})
