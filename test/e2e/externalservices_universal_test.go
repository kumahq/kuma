package e2e_test

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/kumahq/kuma/test/framework"
)

var _ = FDescribe("Test Universal deployment", func() {

	meshDefaulMtlsOn := `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`
	trafficPermissionAll := `
type: TrafficPermission
name: traffic-permission
mesh: default
sources:
- match:
   kuma.io/service: "*"
destinations:
- match:
   kuma.io/service: "*"
`

	trafficRoute := `
type: TrafficRoute
name: route-example
mesh: default
sources:
- match:
   kuma.io/service: "*"
destinations:
- match:
   kuma.io/service: external-service
conf:
  - weight: %s
    destination:
      kuma.io/service: external-service
`

	externalService := `
type: ExternalService
mesh: default
name: httpbin
tags:
  kuma.io/service: external-service
  kuma.io/protocol: http
networking:
  address: %s
`
	var cluster Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(ExternalServiceUniversal()).
			Install(DemoClientUniversal(demoClientToken)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(meshDefaulMtlsOn)(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(trafficPermissionAll)(cluster)
		Expect(err).ToNot(HaveOccurred())

		externalServiceAddres, err := cluster.GetExternalAppAddress("", AppModeEchoServer)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(fmt.Sprintf(externalService, externalServiceAddres+":80"))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := cluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("access external-service", func() {
		stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4002")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	})

	It("disable access to external-service", func() {
		err := YamlUniversal(fmt.Sprintf(trafficRoute, "100"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4002")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		err = YamlUniversal(fmt.Sprintf(trafficRoute, "0"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		retry.DoWithRetry(cluster.GetTesting(), "check service is still accessible", 10, DefaultTimeout, func() (string, error) {
			stdout, _, err = cluster.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "localhost:4002")
			if err != nil {
				return "", err
			}

			if strings.Contains(stdout, "HTTP/1.1 200 OK") {
				return "", nil
			}

			return "", fmt.Errorf("External service still accessible [%s]", stdout)
		})
	})

})
