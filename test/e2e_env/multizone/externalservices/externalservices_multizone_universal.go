package externalservices

import (
	"encoding/base64"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
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

	externalServiceRes := func(service, address string, tls bool, caCert []byte) *core_mesh.ExternalServiceResource {
		res := &core_mesh.ExternalServiceResource{
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: service,
			},
			Spec: &mesh_proto.ExternalService{
				Tags: map[string]string{
					mesh_proto.ServiceTag:  service,
					mesh_proto.ProtocolTag: core_mesh.ProtocolHTTP,
				},
				Networking: &mesh_proto.ExternalService_Networking{
					Address: address,
					Tls: &mesh_proto.ExternalService_Networking_TLS{
						Enabled: tls,
					},
				},
			},
		}
		if tls {
			res.Spec.Networking.Tls.CaCert = &system_proto.DataSource{
				Type: &system_proto.DataSource_Inline{
					Inline: &wrapperspb.BytesValue{
						Value: caCert,
					},
				},
			}
		}
		return res
	}

	es1 := "external-service-1"
	es2 := "external-service-2"

	const clusterName1 = "kuma-es-1"
	const clusterName2 = "kuma-es-2"
	const clusterName3 = "kuma-es-3"
	const clusterName4 = "kuma-es-4"

	const defaultMesh = "default"

	var global, zone1, zone2, external Cluster
	var externalUni *UniversalCluster

	BeforeEach(func() {
		// External Service non-Kuma Cluster
		external = NewUniversalCluster(NewTestingT(), clusterName4, Silent)

		err := NewClusterSetup().
			Install(Parallel(
				TestServerExternalServiceUniversal("es-http", 80, false, WithDockerContainerName("kuma-es-4_es-http")),
				TestServerExternalServiceUniversal("es-https", 443, true, WithDockerContainerName("kuma-es-4_es-https")),
				TestServerExternalServiceUniversal("es-for-kuma-es-2", 80, false, WithDockerContainerName("kuma-es-4_es-for-kuma-es-2")),
				TestServerExternalServiceUniversal("es-for-kuma-es-3", 80, false, WithDockerContainerName("kuma-es-4_es-for-kuma-es-3")),
			)).
			Setup(external)
		Expect(err).ToNot(HaveOccurred())
		externalUni = external.(*UniversalCluster)

		// Global
		global = NewUniversalCluster(NewTestingT(), clusterName1, Silent)
		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Install(YamlUniversal(fmt.Sprintf(meshDefaulMtlsOn, "false"))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		wg := sync.WaitGroup{}
		wg.Add(2)
		// Cluster 1
		zone1 = NewUniversalCluster(NewTestingT(), clusterName2, Silent)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			err = NewClusterSetup().
				Install(Kuma(core.Zone,
					WithGlobalAddress(globalCP.GetKDSServerAddress()),
					WithHDS(false),
				)).
				Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, WithTransparentProxy(true))).
				Setup(zone1)
			Expect(err).ToNot(HaveOccurred())
		}()

		// Cluster 2
		zone2 = NewUniversalCluster(NewTestingT(), clusterName3, Silent)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			err = NewClusterSetup().
				Install(Kuma(core.Zone,
					WithGlobalAddress(globalCP.GetKDSServerAddress()),
					WithHDS(false),
				)).
				Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, WithTransparentProxy(true))).
				Setup(zone2)
			Expect(err).ToNot(HaveOccurred())
		}()
		wg.Wait()
	})

	AfterEachFailure(func() {
		DebugUniversal(global, defaultMesh)
		DebugUniversal(zone1, defaultMesh)
		DebugUniversal(zone2, defaultMesh)
	})

	E2EAfterEach(func() {
		Expect(external.DismissCluster()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("should route to external-service", func() {
		err := ResourceUniversal(externalServiceRes(es1, "kuma-es-4_es-http:80", false, nil))(global)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				zone1, "demo-client", "external-service-1.mesh",
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).ToNot(ContainSubstring("HTTPS"))
		}, "1m", "3s").Should(Succeed())

		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				zone1, "demo-client", "kuma-es-4_es-http:80",
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).ToNot(ContainSubstring("HTTPS"))
		}, "1m", "3s").Should(Succeed())

		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				zone2, "demo-client", "external-service-1.mesh",
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).ToNot(ContainSubstring("HTTPS"))
		}, "1m", "3s").Should(Succeed())

		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				zone2, "demo-client", "kuma-es-4_es-http:80",
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).ToNot(ContainSubstring("HTTPS"))
		}, "1m", "3s").Should(Succeed())
	})

	It("should route to external-service over tls", func() {
		// when set invalid certificate
		otherCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURMRENDQWhTZ0F3SUJBZ0lRSGRQaHhPZlhnV3VOeG9GbFYvRXdxVEFOQmdrcWhraUc5dzBCQVFzRkFEQVAKTVEwd0N3WURWUVFERXdScmRXMWhNQjRYRFRJd01Ea3hOakV5TWpnME5Gb1hEVE13TURreE5ERXlNamcwTkZvdwpEekVOTUFzR0ExVUVBeE1FYTNWdFlUQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCCkFPWkdiV2hTbFFTUnhGTnQ1cC8yV0NLRnlIWjNDdXdOZ3lMRVA3blM0Wlh5a3hzRmJZU3VWM2JJZ0Y3YlQvdXEKYTVRaXJlK0M2MGd1aEZicExjUGgyWjZVZmdJZDY5R2xRekhNVlljbUxHalZRdXlBdDRGTU1rVGZWRWw1STRPYQorMml0M0J2aWhWa0toVXo4eTVSUjVLYnFKZkdwNFoyMEZoNmZ0dG9DRmJlT0RtdkJzWUpGbVVRUytpZm95TVkvClAzUjAzU3U3ZzVpSXZuejd0bWt5ZG9OQzhuR1JEemRENUM4Zkp2clZJMVVYNkpSR3lMS3Q0NW9RWHQxbXhLMTAKNUthTjJ6TlYyV3RIc2FKcDlid3JQSCtKaVpHZVp5dnVoNVV3ckxkSENtcUs3c205VG9kR3p0VVpZMFZ6QWM0cQprWVZpWFk4Z1VqZk5tK2NRclBPMWtOOENBd0VBQWFPQmd6Q0JnREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEClZSMGxCQll3RkFZSUt3WUJCUVVIQXdFR0NDc0dBUVVGQndNQk1BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWUQKVlIwT0JCWUVGR01EQlBQaUJGSjNtdjJvQTlDVHFqZW1GVFYyTUI4R0ExVWRFUVFZTUJhQ0NXeHZZMkZzYUc5egpkSUlKYkc5allXeG9iM04wTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFDLzE3UXdlT3BHZGIxTUVCSjhYUEc3CjNzSy91dG9XTFgxdGpmOFN1MURnYTZDRFQvZVRXSFpyV1JmODFLT1ZZMDdkbGU1U1JJREsxUWhmYkdHdEZQK1QKdlprcm9vdXNJOVVTMmFDV2xrZUNaV0dUbnF2TG1Eb091anFhZ0RvS1JSdWs0bVFkdE5Ob254aUwvd1p0VEZLaQorMWlOalVWYkxXaURYZEJMeG9SSVZkTE96cWIvTU54d0VsVXlhVERBa29wUXlPV2FURGtZUHJHbWFXamNzZlBHCmFPS293MHplK3pIVkZxVEhiam5DcUVWM2huc1V5UlV3c0JsbjkrakRKWGd3Wk0vdE1sVkpyWkNoMFNsZTlZNVoKTU9CMGZDZjZzVE1OUlRHZzVMcGw2dUlZTS81SU5wbUhWTW8zbjdNQlNucEVEQVVTMmJmL3VvNWdJaXE2WENkcAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
		cert, _ := base64.StdEncoding.DecodeString(otherCert)
		err := ResourceUniversal(externalServiceRes(es2, "kuma-es-4_es-https:443", true, cert))(global)
		Expect(err).ToNot(HaveOccurred())

		// then accessing the secured external service fails
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(zone1, "demo-client", "http://kuma-es-4_es-https:443")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(503))
		}, "1m", "1s").Should(Succeed())

		// when correct cert
		correctCert, _, err := externalUni.Exec("", "", "es-https", "cat /certs/cert.pem")
		Expect(err).ToNot(HaveOccurred())
		Expect(correctCert).ToNot(BeEmpty())

		err = ResourceUniversal(externalServiceRes(es2, "kuma-es-4_es-https:443", true, []byte(correctCert)))(global)
		Expect(err).ToNot(HaveOccurred())

		// then accessing the secured external service succeeds
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				zone1, "demo-client", "http://kuma-es-4_es-https:443",
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(ContainSubstring("es-https"))
		}, "1m", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				zone2, "demo-client", "http://kuma-es-4_es-https:443",
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(ContainSubstring("es-https"))
		}, "1m", "3s").Should(Succeed())
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
		Expect(YamlUniversal(externalServiceWithZone("kuma-es-2", "kuma-es-4_es-for-kuma-es-2:80"))(global)).To(Succeed())
		Expect(YamlUniversal(externalServiceWithZone("kuma-es-3", "kuma-es-4_es-for-kuma-es-3:80"))(global)).To(Succeed())
		// then
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(zone1, "demo-client", "es-for-zones.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKey(Equal("es-for-kuma-es-2")),
				HaveKey(Equal("es-for-kuma-es-3")),
			),
		)

		// when locality-aware lb is enabled
		Expect(YamlUniversal(fmt.Sprintf(meshDefaulMtlsOn, "true"))(global)).To(Succeed())
		// then
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(zone1, "demo-client", "es-for-zones.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(1),
				HaveKey(Equal("es-for-kuma-es-2")),
			),
		)
	})
}
