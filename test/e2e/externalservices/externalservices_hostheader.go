package externalservices

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ExternalServiceHostHeader() {
	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	externalService := `
type: ExternalService
mesh: default
name: httpbin
tags:
  kuma.io/service: httpbin
  kuma.io/protocol: http
networking:
  address: httpbin.org:80
  tls:
    enabled: false
`

	BeforeEach(func() {
		cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		deployOptsFuncs = KumaUniversalDeployOpts

		err := NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Install(YamlUniversal(externalService)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "dp-demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = DemoClientUniversal("dp-demo-client", "default", demoClientToken, WithTransparentProxy(true))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should auto rewrite host header", func() {
		Eventually(func() bool {
			stdout, _, _ := cluster.Exec("", "", "dp-demo-client",
				"curl", "httpbin.mesh/get")
			return strings.Contains(stdout, `"Host": "httpbin.org"`)
		}, "30s", "500ms").Should(BeTrue())
	})
}
