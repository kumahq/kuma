package externalservices

import (
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

func ExternalServicesOnMultizoneUniversal() {
	meshDefaulMtlsOn := `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
networking:
  outbound:
    passthrough: true
routing:
  localityAwareLoadBalancing: %s
`

	externalService := `
type: ExternalService
mesh: default
name: external-service-%s
tags:
  kuma.io/service: external-service-%s
  kuma.io/protocol: http
networking:
  address: %s
  tls:
    enabled: %s
    caCert:
      inline: "%s"
`
	es1 := "1"
	es2 := "2"

	const defaultMesh = "default"

	var global, zone1, zone2, external Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3, Kuma4, Kuma5, Kuma6},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// External Service non-Kuma Cluster
		external = clusters.GetCluster(Kuma3)

		// todo(lobkovilya): use test-server as an external service
		err = NewClusterSetup().
			Install(externalservice.Install(externalservice.HttpServer, externalservice.UniversalAppEchoServer)).
			Install(externalservice.Install(externalservice.HttpsServer, externalservice.UniversalAppHttpsEchoServer)).
			Install(externalservice.Install("es-for-kuma-4", externalservice.ExternalServiceCommand(80, "{\\\"instance\\\":\\\"kuma-4\\\"}"))).
			Install(externalservice.Install("es-for-kuma-5", externalservice.ExternalServiceCommand(80, "{\\\"instance\\\":\\\"kuma-5\\\"}"))).
			Setup(external)
		Expect(err).ToNot(HaveOccurred())

		externalServiceAddress := externalservice.From(external, externalservice.HttpServer).GetExternalAppAddress()
		Expect(externalServiceAddress).ToNot(BeEmpty())

		// Global
		global = clusters.GetCluster(Kuma6)
		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Install(YamlUniversal(fmt.Sprintf(meshDefaulMtlsOn, "false"))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		demoClientToken, err := globalCP.GenerateDpToken(defaultMesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())

		// Cluster 1
		zone1 = clusters.GetCluster(Kuma4)

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithHDS(false),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, demoClientToken, WithTransparentProxy(true))).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())

		// Cluster 2
		zone2 = clusters.GetCluster(Kuma5)

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithHDS(false),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, demoClientToken, WithTransparentProxy(true))).
			Setup(zone2)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(external.DismissCluster()).To(Succeed())

		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())

		Expect(zone2.DeleteKuma()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())

		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("should route to external-service", func() {
		err := YamlUniversal(fmt.Sprintf(externalService,
			es1, es1,
			"kuma-3_externalservice-http-server:80",
			"false", ""))(global)
		Expect(err).ToNot(HaveOccurred())

		stdout, _, err := zone1.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "external-service-1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("HTTPS"))

		stdout, _, err = zone1.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "kuma-3_externalservice-http-server:80")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("HTTPS"))

		stdout, _, err = zone2.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "external-service-1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("HTTPS"))

		stdout, _, err = zone2.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "kuma-3_externalservice-http-server:80")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).ToNot(ContainSubstring("HTTPS"))
	})

	It("should route to external-service over tls", func() {
		// when set invalid certificate
		otherCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURMRENDQWhTZ0F3SUJBZ0lRSGRQaHhPZlhnV3VOeG9GbFYvRXdxVEFOQmdrcWhraUc5dzBCQVFzRkFEQVAKTVEwd0N3WURWUVFERXdScmRXMWhNQjRYRFRJd01Ea3hOakV5TWpnME5Gb1hEVE13TURreE5ERXlNamcwTkZvdwpEekVOTUFzR0ExVUVBeE1FYTNWdFlUQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCCkFPWkdiV2hTbFFTUnhGTnQ1cC8yV0NLRnlIWjNDdXdOZ3lMRVA3blM0Wlh5a3hzRmJZU3VWM2JJZ0Y3YlQvdXEKYTVRaXJlK0M2MGd1aEZicExjUGgyWjZVZmdJZDY5R2xRekhNVlljbUxHalZRdXlBdDRGTU1rVGZWRWw1STRPYQorMml0M0J2aWhWa0toVXo4eTVSUjVLYnFKZkdwNFoyMEZoNmZ0dG9DRmJlT0RtdkJzWUpGbVVRUytpZm95TVkvClAzUjAzU3U3ZzVpSXZuejd0bWt5ZG9OQzhuR1JEemRENUM4Zkp2clZJMVVYNkpSR3lMS3Q0NW9RWHQxbXhLMTAKNUthTjJ6TlYyV3RIc2FKcDlid3JQSCtKaVpHZVp5dnVoNVV3ckxkSENtcUs3c205VG9kR3p0VVpZMFZ6QWM0cQprWVZpWFk4Z1VqZk5tK2NRclBPMWtOOENBd0VBQWFPQmd6Q0JnREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEClZSMGxCQll3RkFZSUt3WUJCUVVIQXdFR0NDc0dBUVVGQndNQk1BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWUQKVlIwT0JCWUVGR01EQlBQaUJGSjNtdjJvQTlDVHFqZW1GVFYyTUI4R0ExVWRFUVFZTUJhQ0NXeHZZMkZzYUc5egpkSUlKYkc5allXeG9iM04wTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFDLzE3UXdlT3BHZGIxTUVCSjhYUEc3CjNzSy91dG9XTFgxdGpmOFN1MURnYTZDRFQvZVRXSFpyV1JmODFLT1ZZMDdkbGU1U1JJREsxUWhmYkdHdEZQK1QKdlprcm9vdXNJOVVTMmFDV2xrZUNaV0dUbnF2TG1Eb091anFhZ0RvS1JSdWs0bVFkdE5Ob254aUwvd1p0VEZLaQorMWlOalVWYkxXaURYZEJMeG9SSVZkTE96cWIvTU54d0VsVXlhVERBa29wUXlPV2FURGtZUHJHbWFXamNzZlBHCmFPS293MHplK3pIVkZxVEhiam5DcUVWM2huc1V5UlV3c0JsbjkrakRKWGd3Wk0vdE1sVkpyWkNoMFNsZTlZNVoKTU9CMGZDZjZzVE1OUlRHZzVMcGw2dUlZTS81SU5wbUhWTW8zbjdNQlNucEVEQVVTMmJmL3VvNWdJaXE2WENkcAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
		err := YamlUniversal(fmt.Sprintf(externalService,
			es2, es2,
			"kuma-3_externalservice-https-server:443",
			"true",
			otherCert))(global)
		Expect(err).ToNot(HaveOccurred())

		// then accessing the secured external service fails
		_, _, err = zone1.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://kuma-3_externalservice-https-server:443")
		Expect(err).To(HaveOccurred())

		// when set proper certificate
		externalServiceCaCert := externalservice.From(external, externalservice.HttpsServer).GetCert()
		Expect(externalServiceCaCert).ToNot(BeEmpty())

		err = YamlUniversal(fmt.Sprintf(externalService,
			es2, es2,
			"kuma-3_externalservice-https-server:443",
			"true",
			base64.StdEncoding.EncodeToString([]byte(externalServiceCaCert))))(global)
		Expect(err).ToNot(HaveOccurred())

		// then accessing the secured external service succeeds
		stdout, _, err := zone1.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://kuma-3_externalservice-https-server:443")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring("HTTPS"))

		stdout, _, err = zone2.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://kuma-3_externalservice-https-server:443")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring("HTTPS"))
	})

	It("should respect external-service's zone tag in locality-aware lb mode", func() {
		externalServiceWithZone := func(zone, address string) string {
			return fmt.Sprintf(`
type: ExternalService
mesh: default
name: es-for-%s
tags:
  kuma.io/service: es-for-zones
  kuma.io/protocol: http
  kuma.io/zone: %s
networking:
  address: %s
`, zone, zone, address)
		}

		// given 2 external services with different zone tag
		Expect(YamlUniversal(externalServiceWithZone("kuma-4", "kuma-3_externalservice-es-for-kuma-4:80"))(global)).To(Succeed())
		Expect(YamlUniversal(externalServiceWithZone("kuma-5", "kuma-3_externalservice-es-for-kuma-5:80"))(global)).To(Succeed())
		// then
		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(zone1, "demo-client", "es-for-zones.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKey(Equal("kuma-4")),
				HaveKey(Equal("kuma-5")),
			),
		)

		// when locality-aware lb is enabled
		Expect(YamlUniversal(fmt.Sprintf(meshDefaulMtlsOn, "true"))(global)).To(Succeed())
		// then
		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(zone1, "demo-client", "es-for-zones.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(1),
				HaveKey(Equal("kuma-4")),
			),
		)
	})
}
