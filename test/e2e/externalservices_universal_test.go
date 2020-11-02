package e2e_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

var _ = Describe("Test ExternalServices on Universal", func() {

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
    ca_cert:
      inline: "%s"
`
	externalServiceCaCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM5akNDQWQ0Q0NRRElFLzJhN1N1alhqQU5CZ2txaGtpRzl3MEJBUXNGQURBOU1Rc3dDUVlEVlFRR0V3SlYKVXpFV01CUUdBMVVFQ0F3TlUyRnVJRVp5WVc1amFYTmpiekVXTUJRR0ExVUVCd3dOVTJGdUlFWnlZVzVqYVhOagpiekFlRncweU1ERXhNREl4TlRFeU16bGFGdzB6TURFd016RXhOVEV5TXpsYU1EMHhDekFKQmdOVkJBWVRBbFZUCk1SWXdGQVlEVlFRSURBMVRZVzRnUm5KaGJtTnBjMk52TVJZd0ZBWURWUVFIREExVFlXNGdSbkpoYm1OcGMyTnYKTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUF0dzlPZUwvVlltSk5zeXZYNUcvNwplaFpOZkxFRWJyUXJVQjB4NmtEM1Y1aGFUV1NUOW9mbjN1NEljN1VIcG1lV0pla2NLMjFGKyt1QVVVeXZEd3grClAySGpZWm9ZTFQ1Z1ZTaHpFZlFoSjNnNWFjZDU4ZUt3LzRQL0JncDRySnFGT2hzeU1TV1JvRFFwVFdYTkwrUWoKUVljLzdJT2VMMkxBcmhpL3VmdEFuMEJYRERlQmhyNTl3RWkwY2UvNVpEM3dGMGlDeW5sajRudlVFNlg5MzVLZgpoRWc1K3piTHp5RXpJQ29qajJoMDNlYURUM29yM2ZUQmhmdDFRalYyTUxCLytBbWM5eUtQcUNGMVJ0NTNpN29SCjhFY3FLTXRmSXhQZ0IydkhjbnkvU3NDeVhSU3NGcStJN0tidTdKOEpqWm43ZitJSGRhNGduUW9hbkpUbnpiZC8Ka3dJREFRQUJNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUNjK2V1alVvZlhzeGpxZ3FXaFpsL25lVUp5bVBXYQpRM0VlTVlsZHQzT3huYXd6SDZEKzJQemM0d0RacVl6dWlNTk51emp1ZEpFOW9kcUkrSUwwVmdadmJPQ00weExmClM2blNWcVRNc1lVL1VDcHdPZXk1MWRzSTd3Y21nYlJzVnl6TzM5ZDJIRUhLZ0VUbVZ1eXlQWTRMTEw5aW1aUjQKdVIrK2dlVTd6bjVzbG9MWDZFU3VIMEIxSEJRNVJmcXdOMWxGL2NvcUE0QjhhU1pJYnA0QjhDajBRTUcxdXMzZApFTnE1b0VaRmtDdzhnMnBkakdSSytlbDlmaGN6K0xZWjBSaEwzK2tWTzlFcks2RGd5UVJ4UGErM29TOWkzazJvCkhoOU9jbUJRWnZXNlNuY2RvMHhVM0plRDNkR0p0dkhJWjk2bmE4cXMvbFRGejVkZWdiZk5kRUloCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
	es1 := "1"
	es2 := "2"

	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma1)
		deployOptsFuncs = []DeployOptionsFunc{}

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
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
			"false",
			externalServiceCaCert))(cluster)
		Expect(err).ToNot(HaveOccurred())

		externalServiceAddress = externalservice.From(cluster, externalservice.HttpsServer).GetExternalAppAddress()
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(fmt.Sprintf(externalService,
			es2, es2,
			externalServiceAddress+":443",
			"true",
			externalServiceCaCert))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := cluster.DeleteKuma(deployOptsFuncs...)
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
