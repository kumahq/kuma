package externalservices

import (
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "external-service-base"
	meshNameNoDefaults := "external-service-no-default-policy"
	meshDefaulMtlsOn := func(meshName string) InstallFunc {
		return YamlUniversal(fmt.Sprintf(`
type: Mesh
name: "%s"
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
networking:
  outbound:
    passthrough: false
`, meshName))
	}

	externalService := func(service, address, meshName string, tls bool, caCert []byte) *core_mesh.ExternalServiceResource {
		res := &core_mesh.ExternalServiceResource{
			Meta: &test_model.ResourceMeta{
				Mesh: meshName,
				Name: fmt.Sprintf("%s-%s", service, meshName),
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

	esHttpName := "es-http"
	esHttpsName := "es-https"
	esHttp2Name := "es-http-2"

	var esHttpHostPort string
	var esHttp2HostPort string
	var esHttpsHostPort string

	var esHttpContainerName string
	var esHttpsContainerName string
	var esHttp2ContainerName string

	BeforeAll(func() {
		esHttpContainerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, esHttpName)
		esHttpsContainerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, esHttpsName)
		esHttp2ContainerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, esHttp2Name)

		esHttpHostPort = fmt.Sprintf("%s:%d", esHttpContainerName, 80)
		esHttpsHostPort = fmt.Sprintf("%s:%d", esHttpsContainerName, 443)
		esHttp2HostPort = fmt.Sprintf("%s:%d", esHttp2ContainerName, 81)

		err := NewClusterSetup().
			Install(meshDefaulMtlsOn(meshName)).
			Install(TrafficPermissionUniversal(meshName)).
			Install(TrafficRouteUniversal(meshName)).
			Install(meshDefaulMtlsOn(meshNameNoDefaults)).
			Install(TestServerExternalServiceUniversal(esHttpName, 80, false, WithDockerContainerName(esHttpContainerName))).
			Install(TestServerExternalServiceUniversal(esHttpsName, 443, true, WithDockerContainerName(esHttpsContainerName))).
			Install(TestServerExternalServiceUniversal(esHttp2Name, 81, false, WithDockerContainerName(esHttp2ContainerName))).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-no-defaults", meshNameNoDefaults, WithTransparentProxy(true))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(esHttpName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(esHttpsName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(esHttp2Name)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	checkSuccessfulRequest := func(url, clientName string, matcher types.GomegaMatcher) {
		Eventually(func(g Gomega) {
			stdout, stderr, err := client.CollectResponse(
				universal.Cluster, clientName, url,
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(matcher)
		}, "30s", "500ms").WithOffset(1).Should(Succeed())
	}

	contextFor := func(name, meshName, clientName string) {
		Context(name, func() {
			It("should route to external-service", func() {
				err := universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-1", esHttpHostPort, meshName, false, nil)))
				Expect(err).ToNot(HaveOccurred())

				checkSuccessfulRequest("ext-srv-1.mesh", clientName, And(
					Not(ContainSubstring("HTTPS")),
					// Should rewrite host header
					ContainSubstring(fmt.Sprintf(`"Host":["%s"]`, esHttpContainerName)),
				))
				checkSuccessfulRequest(esHttpHostPort, clientName, Not(ContainSubstring("HTTPS")))
			})

			It("should route to external-service with same hostname but different ports", func() {
				err := universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-1", esHttpHostPort, meshName, false, nil)))
				Expect(err).ToNot(HaveOccurred())

				err = universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-2", esHttp2HostPort, meshName, false, nil)))
				Expect(err).ToNot(HaveOccurred())

				// when access the first external service with .mesh
				checkSuccessfulRequest("ext-srv-1.mesh", clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("es-http")))
				checkSuccessfulRequest(esHttpHostPort, clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("es-http")))

				checkSuccessfulRequest("ext-srv-2.mesh", clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("es-http-2")))
				checkSuccessfulRequest(esHttp2HostPort, clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("es-http-2")))
			})

			It("should route to external-service over tls", func() {
				// when set invalid certificate
				otherCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURMRENDQWhTZ0F3SUJBZ0lRSGRQaHhPZlhnV3VOeG9GbFYvRXdxVEFOQmdrcWhraUc5dzBCQVFzRkFEQVAKTVEwd0N3WURWUVFERXdScmRXMWhNQjRYRFRJd01Ea3hOakV5TWpnME5Gb1hEVE13TURreE5ERXlNamcwTkZvdwpEekVOTUFzR0ExVUVBeE1FYTNWdFlUQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCCkFPWkdiV2hTbFFTUnhGTnQ1cC8yV0NLRnlIWjNDdXdOZ3lMRVA3blM0Wlh5a3hzRmJZU3VWM2JJZ0Y3YlQvdXEKYTVRaXJlK0M2MGd1aEZicExjUGgyWjZVZmdJZDY5R2xRekhNVlljbUxHalZRdXlBdDRGTU1rVGZWRWw1STRPYQorMml0M0J2aWhWa0toVXo4eTVSUjVLYnFKZkdwNFoyMEZoNmZ0dG9DRmJlT0RtdkJzWUpGbVVRUytpZm95TVkvClAzUjAzU3U3ZzVpSXZuejd0bWt5ZG9OQzhuR1JEemRENUM4Zkp2clZJMVVYNkpSR3lMS3Q0NW9RWHQxbXhLMTAKNUthTjJ6TlYyV3RIc2FKcDlid3JQSCtKaVpHZVp5dnVoNVV3ckxkSENtcUs3c205VG9kR3p0VVpZMFZ6QWM0cQprWVZpWFk4Z1VqZk5tK2NRclBPMWtOOENBd0VBQWFPQmd6Q0JnREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEClZSMGxCQll3RkFZSUt3WUJCUVVIQXdFR0NDc0dBUVVGQndNQk1BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWUQKVlIwT0JCWUVGR01EQlBQaUJGSjNtdjJvQTlDVHFqZW1GVFYyTUI4R0ExVWRFUVFZTUJhQ0NXeHZZMkZzYUc5egpkSUlKYkc5allXeG9iM04wTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFDLzE3UXdlT3BHZGIxTUVCSjhYUEc3CjNzSy91dG9XTFgxdGpmOFN1MURnYTZDRFQvZVRXSFpyV1JmODFLT1ZZMDdkbGU1U1JJREsxUWhmYkdHdEZQK1QKdlprcm9vdXNJOVVTMmFDV2xrZUNaV0dUbnF2TG1Eb091anFhZ0RvS1JSdWs0bVFkdE5Ob254aUwvd1p0VEZLaQorMWlOalVWYkxXaURYZEJMeG9SSVZkTE96cWIvTU54d0VsVXlhVERBa29wUXlPV2FURGtZUHJHbWFXamNzZlBHCmFPS293MHplK3pIVkZxVEhiam5DcUVWM2huc1V5UlV3c0JsbjkrakRKWGd3Wk0vdE1sVkpyWkNoMFNsZTlZNVoKTU9CMGZDZjZzVE1OUlRHZzVMcGw2dUlZTS81SU5wbUhWTW8zbjdNQlNucEVEQVVTMmJmL3VvNWdJaXE2WENkcAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
				caCert, _ := base64.StdEncoding.DecodeString(otherCert)
				err := universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-tls", esHttpsHostPort, meshName, true, caCert)))
				Expect(err).ToNot(HaveOccurred())

				// then accessing the secured external service fails
				Eventually(func(g Gomega) {
					response, err := client.CollectFailure(universal.Cluster, clientName, "http://"+esHttpsHostPort)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.ResponseCode).To(Equal(503))
				}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

				// when set proper certificate
				cert, _, err := universal.Cluster.Exec("", "", "es-https", "cat /certs/cert.pem")
				Expect(err).ToNot(HaveOccurred())
				Expect(cert).ToNot(BeEmpty())

				err = universal.Cluster.Install(ResourceUniversal(externalService("ext-srv-tls", esHttpsHostPort, meshName, true, []byte(cert))))
				Expect(err).ToNot(HaveOccurred())

				// then accessing the secured external service succeeds
				checkSuccessfulRequest("http://"+esHttpsHostPort, clientName, Not(ContainSubstring("HTTPS")))
			})
		})
	}

	contextFor("without default policies", meshNameNoDefaults, "demo-client-no-defaults")
	contextFor("with default policies", meshName, "demo-client")
}
