package externalservices

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func ExternalServiceHostHeader() {
	const meshName = "es-header"

	externalService := `
type: ExternalService
mesh: es-header
name: httpbin
tags:
  kuma.io/service: httpbin
  kuma.io/protocol: http
networking:
  address: httpbin.org:80
  tls:
    enabled: false
`

	BeforeAll(func() {
		E2EDeferCleanup(func() {
			Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("dp-demo-client", meshName, WithTransparentProxy(true))).
			Install(YamlUniversal(externalService)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should auto rewrite host header", func() {
		Eventually(func() bool {
			stdout, _, _ := env.Cluster.Exec("", "", "dp-demo-client",
				"curl", "httpbin.mesh/get")
			return strings.Contains(stdout, `"Host": "httpbin.org"`)
		}, "30s", "500ms").Should(BeTrue())
	})
}
