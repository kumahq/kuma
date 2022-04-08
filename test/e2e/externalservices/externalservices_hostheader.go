package externalservices

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ExternalServiceHostHeader() {
	var cluster Cluster

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

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(YamlUniversal(externalService)).
			Install(DemoClientUniversal("dp-demo-client", "default", WithTransparentProxy(true))).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteKuma()).To(Succeed())
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
