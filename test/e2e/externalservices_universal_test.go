package e2e_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

var _ = XDescribe("Test ExternalServices on Universal", func() {

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
  - weight: 1
    destination:
      kuma.io/service: external-service
      id: "%s"
`

	externalService := `
type: ExternalService
mesh: default
name: external-service-%s
tags:
  kuma.io/service: external-service
  kuma.io/protocol: http
  id: "%s"
networking:
  address: %s
  tls:
    enabled: %s
`
	es1 := "1"
	es2 := "2"

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
			Install(externalservice.Install(externalservice.HttpServer, externalservice.UniversalAppEchoServer)).
			Install(externalservice.Install(externalservice.HttpsServer, externalservice.UniversalAppHttpsEchoServer)).
			Install(DemoClientUniversal(demoClientToken)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(meshDefaulMtlsOn)(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(trafficPermissionAll)(cluster)
		Expect(err).ToNot(HaveOccurred())

		externalServiceAddress := externalservice.From(cluster, externalservice.HttpServer).GetExternalAppAddress()
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(fmt.Sprintf(externalService,
			es1, es1,
			externalServiceAddress+":80",
			"false"))(cluster)
		Expect(err).ToNot(HaveOccurred())

		externalServiceAddress = externalservice.From(cluster, externalservice.HttpsServer).GetExternalAppAddress()
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(fmt.Sprintf(externalService,
			es2, es2,
			externalServiceAddress+":443",
			"true"))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := cluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should route to external-service", func() {
		err := YamlUniversal(fmt.Sprintf(trafficRoute, es1))(cluster)
		Expect(err).ToNot(HaveOccurred())

		stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "localhost:5000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("HTTPS"))
	})

	It("should route to external-service over tls", func() {
		err := YamlUniversal(fmt.Sprintf(trafficRoute, es2))(cluster)
		Expect(err).ToNot(HaveOccurred())

		stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://localhost:5000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring("HTTPS"))
	})

})
