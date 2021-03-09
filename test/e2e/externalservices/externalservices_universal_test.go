package externalservices_test

import (
	"encoding/base64"
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
  split:
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
    caCert:
      inline: "%s"
`
	es1 := "1"
	es2 := "2"

	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma3)
		deployOptsFuncs = []DeployOptionsFunc{}

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(externalservice.Install(externalservice.HttpServer, externalservice.UniversalAppEchoServer)).
			Install(externalservice.Install(externalservice.HttpsServer, externalservice.UniversalAppHttpsEchoServer)).
			Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(meshDefaulMtlsOn)(cluster)
		Expect(err).ToNot(HaveOccurred())

		externalServiceAddress := externalservice.From(cluster, externalservice.HttpServer).GetExternalAppAddress()
		Expect(externalServiceAddress).ToNot(BeEmpty())

		err = YamlUniversal(fmt.Sprintf(externalService,
			es1, es1,
			externalServiceAddress+":80",
			"false", ""))(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
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
		// set the route to the secured external service
		err := YamlUniversal(fmt.Sprintf(trafficRoute, es2))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// when set invalid certificate
		externalServiceAddress := externalservice.From(cluster, externalservice.HttpsServer).GetExternalAppAddress()
		Expect(externalServiceAddress).ToNot(BeEmpty())

		otherCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURMRENDQWhTZ0F3SUJBZ0lRSGRQaHhPZlhnV3VOeG9GbFYvRXdxVEFOQmdrcWhraUc5dzBCQVFzRkFEQVAKTVEwd0N3WURWUVFERXdScmRXMWhNQjRYRFRJd01Ea3hOakV5TWpnME5Gb1hEVE13TURreE5ERXlNamcwTkZvdwpEekVOTUFzR0ExVUVBeE1FYTNWdFlUQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCCkFPWkdiV2hTbFFTUnhGTnQ1cC8yV0NLRnlIWjNDdXdOZ3lMRVA3blM0Wlh5a3hzRmJZU3VWM2JJZ0Y3YlQvdXEKYTVRaXJlK0M2MGd1aEZicExjUGgyWjZVZmdJZDY5R2xRekhNVlljbUxHalZRdXlBdDRGTU1rVGZWRWw1STRPYQorMml0M0J2aWhWa0toVXo4eTVSUjVLYnFKZkdwNFoyMEZoNmZ0dG9DRmJlT0RtdkJzWUpGbVVRUytpZm95TVkvClAzUjAzU3U3ZzVpSXZuejd0bWt5ZG9OQzhuR1JEemRENUM4Zkp2clZJMVVYNkpSR3lMS3Q0NW9RWHQxbXhLMTAKNUthTjJ6TlYyV3RIc2FKcDlid3JQSCtKaVpHZVp5dnVoNVV3ckxkSENtcUs3c205VG9kR3p0VVpZMFZ6QWM0cQprWVZpWFk4Z1VqZk5tK2NRclBPMWtOOENBd0VBQWFPQmd6Q0JnREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEClZSMGxCQll3RkFZSUt3WUJCUVVIQXdFR0NDc0dBUVVGQndNQk1BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWUQKVlIwT0JCWUVGR01EQlBQaUJGSjNtdjJvQTlDVHFqZW1GVFYyTUI4R0ExVWRFUVFZTUJhQ0NXeHZZMkZzYUc5egpkSUlKYkc5allXeG9iM04wTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFDLzE3UXdlT3BHZGIxTUVCSjhYUEc3CjNzSy91dG9XTFgxdGpmOFN1MURnYTZDRFQvZVRXSFpyV1JmODFLT1ZZMDdkbGU1U1JJREsxUWhmYkdHdEZQK1QKdlprcm9vdXNJOVVTMmFDV2xrZUNaV0dUbnF2TG1Eb091anFhZ0RvS1JSdWs0bVFkdE5Ob254aUwvd1p0VEZLaQorMWlOalVWYkxXaURYZEJMeG9SSVZkTE96cWIvTU54d0VsVXlhVERBa29wUXlPV2FURGtZUHJHbWFXamNzZlBHCmFPS293MHplK3pIVkZxVEhiam5DcUVWM2huc1V5UlV3c0JsbjkrakRKWGd3Wk0vdE1sVkpyWkNoMFNsZTlZNVoKTU9CMGZDZjZzVE1OUlRHZzVMcGw2dUlZTS81SU5wbUhWTW8zbjdNQlNucEVEQVVTMmJmL3VvNWdJaXE2WENkcAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
		err = YamlUniversal(fmt.Sprintf(externalService,
			es2, es2,
			externalServiceAddress+":443",
			"true",
			otherCert))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then accessing the secured external service fails
		_, _, err = cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://localhost:5000")
		Expect(err).To(HaveOccurred())

		// when set proper certificate
		externalServiceCaCert := externalservice.From(cluster, externalservice.HttpsServer).GetCert()
		Expect(externalServiceCaCert).ToNot(BeEmpty())

		err = YamlUniversal(fmt.Sprintf(externalService,
			es2, es2,
			externalServiceAddress+":443",
			"true",
			base64.StdEncoding.EncodeToString([]byte(externalServiceCaCert))))(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then accessing the secured external service succeeds
		stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://localhost:5000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring("HTTPS"))
	})

})
