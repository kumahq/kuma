package meshexternalservice

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshaccesslog_api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshcircuitbreaker_api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshhealthcheck_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshloadbalancingstrategy_api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func MeshExternalService() {
	var tcpSinkDockerName string
	meshNameNoDefaults := "mesh-external-service-no-default-policy"
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

	meshExternalService := func(service, host, meshName string, port int, tls bool, caCert []byte) *meshexternalservice_api.MeshExternalServiceResource {
		mes := &meshexternalservice_api.MeshExternalServiceResource{
			Meta: &test_model.ResourceMeta{
				Mesh: meshName,
				Name: service,
				Labels: map[string]string{
					"kuma.io/origin": "zone",
				},
			},
			Spec: &meshexternalservice_api.MeshExternalService{
				Match: meshexternalservice_api.Match{
					Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
					Port:     80,
					Protocol: core_mesh.ProtocolHTTP,
				},
				Endpoints: []meshexternalservice_api.Endpoint{{
					Address: host,
					Port:    pointer.To(meshexternalservice_api.Port(port)),
				}},
			},
			Status: &meshexternalservice_api.MeshExternalServiceStatus{},
		}

		if tls {
			mes.Spec.Tls = &meshexternalservice_api.Tls{
				Enabled: true,
				Verification: &meshexternalservice_api.Verification{
					CaCert: &common_api.DataSource{Inline: &caCert},
				},
			}
		}

		return mes
	}

	var esHttpContainerName string
	var esHttpsContainerName string
	var esHttp2ContainerName string

	BeforeAll(func() {
		esHttpName := "mes-http"
		esHttpsName := "mes-https"
		esHttp2Name := "mes-http-2"

		tcpSinkDockerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshNameNoDefaults, "mes-tcp-sink")

		esHttpContainerName = fmt.Sprintf("%s_%s", universal.Cluster.Name(), esHttpName)
		esHttpsContainerName = fmt.Sprintf("%s_%s", universal.Cluster.Name(), esHttpsName)
		esHttp2ContainerName = fmt.Sprintf("%s_%s", universal.Cluster.Name(), esHttp2Name)

		err := NewClusterSetup().
			Install(meshDefaulMtlsOn(meshNameNoDefaults)).
			Install(TcpSinkUniversal("mes-tcp-sink", WithDockerContainerName(tcpSinkDockerName))).
			Install(TestServerExternalServiceUniversal(esHttpName, 80, false, WithDockerContainerName(esHttpContainerName))).
			Install(TestServerExternalServiceUniversal(esHttpsName, 443, true, WithDockerContainerName(esHttpsContainerName))).
			Install(TestServerExternalServiceUniversal(esHttp2Name, 81, false, WithDockerContainerName(esHttp2ContainerName))).
			Install(DemoClientUniversal("mes-demo-client-no-defaults", meshNameNoDefaults, WithTransparentProxy(true))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshNameNoDefaults)
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(universal.Cluster, meshNameNoDefaults,
			meshretry_api.MeshRetryResourceTypeDescriptor,
			meshtimeout_api.MeshTimeoutResourceTypeDescriptor,
			meshcircuitbreaker_api.MeshCircuitBreakerResourceTypeDescriptor,
		)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshNameNoDefaults)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshNameNoDefaults)).To(Succeed())
	})

	checkSuccessfulRequest := func(url, clientName string, matcher types.GomegaMatcher) {
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				universal.Cluster, clientName, url,
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			g.Expect(stdout).To(matcher)
		}, "30s", "500ms").WithOffset(1).Should(Succeed())
	}

	contextFor := func(name, meshName, clientName string) {
		Context(name, func() {
			It("should route to mesh-external-service", func() {
				err := universal.Cluster.Install(ResourceUniversal(meshExternalService("ext-srv-1", esHttpContainerName, meshName, 80, false, nil)))
				Expect(err).ToNot(HaveOccurred())

				checkSuccessfulRequest("ext-srv-1.extsvc.mesh.local", clientName, And(
					Not(ContainSubstring("HTTPS")),
					// Should rewrite host header
					ContainSubstring(fmt.Sprintf(`"Host":["%s"]`, esHttpContainerName)),
				))
			})

			It("should route to mesh-external-service with same hostname but different ports", func() {
				err := universal.Cluster.Install(ResourceUniversal(meshExternalService("ext-srv-1", esHttpContainerName, meshName, 80, false, nil)))
				Expect(err).ToNot(HaveOccurred())

				err = universal.Cluster.Install(ResourceUniversal(meshExternalService("ext-srv-2", esHttp2ContainerName, meshName, 81, false, nil)))
				Expect(err).ToNot(HaveOccurred())

				// when access the first external service with .mesh
				checkSuccessfulRequest("ext-srv-1.extsvc.mesh.local", clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("mes-http")))

				checkSuccessfulRequest("ext-srv-2.extsvc.mesh.local", clientName,
					And(Not(ContainSubstring("HTTPS")), ContainSubstring("mes-http-2")))
			})

			It("should route to mesh-external-service over tls", func() {
				// when set invalid certificate
				otherCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURMRENDQWhTZ0F3SUJBZ0lRSGRQaHhPZlhnV3VOeG9GbFYvRXdxVEFOQmdrcWhraUc5dzBCQVFzRkFEQVAKTVEwd0N3WURWUVFERXdScmRXMWhNQjRYRFRJd01Ea3hOakV5TWpnME5Gb1hEVE13TURreE5ERXlNamcwTkZvdwpEekVOTUFzR0ExVUVBeE1FYTNWdFlUQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCCkFPWkdiV2hTbFFTUnhGTnQ1cC8yV0NLRnlIWjNDdXdOZ3lMRVA3blM0Wlh5a3hzRmJZU3VWM2JJZ0Y3YlQvdXEKYTVRaXJlK0M2MGd1aEZicExjUGgyWjZVZmdJZDY5R2xRekhNVlljbUxHalZRdXlBdDRGTU1rVGZWRWw1STRPYQorMml0M0J2aWhWa0toVXo4eTVSUjVLYnFKZkdwNFoyMEZoNmZ0dG9DRmJlT0RtdkJzWUpGbVVRUytpZm95TVkvClAzUjAzU3U3ZzVpSXZuejd0bWt5ZG9OQzhuR1JEemRENUM4Zkp2clZJMVVYNkpSR3lMS3Q0NW9RWHQxbXhLMTAKNUthTjJ6TlYyV3RIc2FKcDlid3JQSCtKaVpHZVp5dnVoNVV3ckxkSENtcUs3c205VG9kR3p0VVpZMFZ6QWM0cQprWVZpWFk4Z1VqZk5tK2NRclBPMWtOOENBd0VBQWFPQmd6Q0JnREFPQmdOVkhROEJBZjhFQkFNQ0FxUXdIUVlEClZSMGxCQll3RkFZSUt3WUJCUVVIQXdFR0NDc0dBUVVGQndNQk1BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWUQKVlIwT0JCWUVGR01EQlBQaUJGSjNtdjJvQTlDVHFqZW1GVFYyTUI4R0ExVWRFUVFZTUJhQ0NXeHZZMkZzYUc5egpkSUlKYkc5allXeG9iM04wTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFDLzE3UXdlT3BHZGIxTUVCSjhYUEc3CjNzSy91dG9XTFgxdGpmOFN1MURnYTZDRFQvZVRXSFpyV1JmODFLT1ZZMDdkbGU1U1JJREsxUWhmYkdHdEZQK1QKdlprcm9vdXNJOVVTMmFDV2xrZUNaV0dUbnF2TG1Eb091anFhZ0RvS1JSdWs0bVFkdE5Ob254aUwvd1p0VEZLaQorMWlOalVWYkxXaURYZEJMeG9SSVZkTE96cWIvTU54d0VsVXlhVERBa29wUXlPV2FURGtZUHJHbWFXamNzZlBHCmFPS293MHplK3pIVkZxVEhiam5DcUVWM2huc1V5UlV3c0JsbjkrakRKWGd3Wk0vdE1sVkpyWkNoMFNsZTlZNVoKTU9CMGZDZjZzVE1OUlRHZzVMcGw2dUlZTS81SU5wbUhWTW8zbjdNQlNucEVEQVVTMmJmL3VvNWdJaXE2WENkcAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
				caCert, _ := base64.StdEncoding.DecodeString(otherCert)
				err := universal.Cluster.Install(ResourceUniversal(meshExternalService("ext-srv-tls", esHttpsContainerName, meshName, 443, true, caCert)))
				Expect(err).ToNot(HaveOccurred())

				// then accessing the secured external service fails
				Eventually(func(g Gomega) {
					response, err := client.CollectFailure(universal.Cluster, clientName, "http://ext-srv-tls.extsvc.mesh.local")
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.ResponseCode).To(Equal(503))
				}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

				// when set proper certificate
				cert, _, err := universal.Cluster.Exec("", "", "mes-https", "cat /certs/cert.pem")
				Expect(err).ToNot(HaveOccurred())
				Expect(cert).ToNot(BeEmpty())

				err = universal.Cluster.Install(ResourceUniversal(meshExternalService("ext-srv-tls", esHttpsContainerName, meshName, 443, true, []byte(cert))))
				Expect(err).ToNot(HaveOccurred())

				// then accessing the secured external service succeeds
				checkSuccessfulRequest("http://ext-srv-tls.extsvc.mesh.local", clientName, Not(ContainSubstring("HTTPS")))
			})
		})
	}

	Context("MeshExternalService with MeshRetry", func() {
		E2EAfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshNameNoDefaults,
				meshretry_api.MeshRetryResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should retry on error", func() {
			meshExternalService := fmt.Sprintf(`
type: MeshExternalService
name: mes-retry
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 80
    - address: %s
      port: 80
`, meshNameNoDefaults, esHttpsContainerName, esHttpContainerName)

			meshRetryPolicy := fmt.Sprintf(`
type: MeshRetry
mesh: %s
name: meshretry-mes-policy
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: mes-retry
      default:
        http:
          numRetries: 5
          retryOn:
            - "5xx"
`, meshNameNoDefaults)
			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService))).To(Succeed())

			// we have 2 endpoints, one http and another https so some requests should fail
			By("Check some errors happen")
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-retry.extsvc.mesh.local",
					client.NoFail(),
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(503))
			}, "30s", "100ms").Should(Succeed())

			By("Apply a MeshRetry policy")
			Expect(universal.Cluster.Install(YamlUniversal(meshRetryPolicy))).To(Succeed())

			By("Eventually all requests succeed consistently")
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-retry.extsvc.mesh.local",
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
		})
	})

	Context("MeshExternalService with MeshTimeout", func() {
		E2EAfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshNameNoDefaults,
				meshtimeout_api.MeshTimeoutResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should target real MeshExternalService resource", func() {
			meshExternalService := fmt.Sprintf(`
type: MeshExternalService
name: mes-timeout
mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 80
`, meshNameNoDefaults, esHttpContainerName)
			timeoutConfig := fmt.Sprintf(`
type: MeshTimeout
name: timeout-for-mes
mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: mes-timeout
      default:
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, meshNameNoDefaults)

			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService))).To(Succeed())

			// given no MeshTimeout
			By("request should pass with the delay")
			Eventually(func(g Gomega) {
				start := time.Now()
				_, err := client.CollectEchoResponse(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-timeout.extsvc.mesh.local",
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
			}, "30s", "1s").Should(Succeed())

			// when timeout applied
			Expect(universal.Cluster.Install(YamlUniversal(timeoutConfig))).To(Succeed())

			// then should timeout after 5 seconds
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-timeout.extsvc.mesh.local",
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(504))
			}, "30s", "1s", MustPassRepeatedly(3)).Should(Succeed())
		})
	})

	Context("MeshExternalService with MeshHTTPRoute", func() {
		E2EAfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshNameNoDefaults,
				meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should route to other endpoint", func() {
			meshExternalService := fmt.Sprintf(`
type: MeshExternalService
name: mes-http-route
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 80
`, meshNameNoDefaults, esHttpContainerName)

			meshExternalService2 := fmt.Sprintf(`
type: MeshExternalService
name: mes-http-2-route
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 81
`, meshNameNoDefaults, esHttp2ContainerName)

			meshHttpRoutePolicy := fmt.Sprintf(`
type: MeshHTTPRoute
mesh: %s
name: mes-http-route-policy
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: mes-http-route
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /
          default:
            backendRefs:
              - kind: MeshExternalService
                name: mes-http-2-route
                weight: 100
`, meshNameNoDefaults)
			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService))).To(Succeed())
			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService2))).To(Succeed())

			By("Check response arrives to mes-http")
			Eventually(func(g Gomega) {
				resp, err := client.CollectEchoResponse(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-http-route.extsvc.mesh.local",
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("mes-http"))
			}, "30s", "1s").Should(Succeed())

			By("Apply a MeshHTTPRoute policy")
			Expect(universal.Cluster.Install(YamlUniversal(meshHttpRoutePolicy))).To(Succeed())

			By("Eventually all arrives to mes-http-2")
			Eventually(func(g Gomega) {
				resp, err := client.CollectEchoResponse(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-http-route.extsvc.mesh.local",
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("mes-http-2"))
			}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
		})
	})

	Context("MeshExternalService with MeshTCPRoute", func() {
		E2EAfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshNameNoDefaults,
				meshtcproute_api.MeshTCPRouteResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should route to other backend", func() {
			meshExternalService := fmt.Sprintf(`
type: MeshExternalService
name: mes-tcp-route
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: tcp
  endpoints:
    - address: %s
      port: 80
`, meshNameNoDefaults, esHttpContainerName)

			meshExternalService2 := fmt.Sprintf(`
type: MeshExternalService
name: mes-tcp-2-route
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: tcp
  endpoints:
    - address: %s
      port: 81
`, meshNameNoDefaults, esHttp2ContainerName)

			meshTcpRoute := fmt.Sprintf(`
type: MeshTCPRoute
name: mes-tcp-route-1
mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: mes-tcp-route
      rules:
        - default:
            backendRefs:
              - kind: MeshExternalService
                name: mes-tcp-2-route
`, meshNameNoDefaults)
			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService))).To(Succeed())
			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService2))).To(Succeed())

			By("Check response arrives to mes-http")
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-tcp-route.extsvc.mesh.local",
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal("mes-http"))
			}, "30s", "1s").Should(Succeed())

			By("Apply a MeshTCPRoute policy")
			Expect(universal.Cluster.Install(YamlUniversal(meshTcpRoute))).To(Succeed())

			By("Eventually all arrives to mes-http-2")
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-tcp-route.extsvc.mesh.local",
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal("mes-http-2"))
			}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
		})
	})

	Context("MeshExternalService with MeshAccessLog", func() {
		expectTrafficLogged := func(makeRequest func(g Gomega)) (string, string) {
			var src, dst string

			Eventually(func(g Gomega) {
				makeRequest(g)

				stdout, _, err := universal.Cluster.Exec("", "", "mes-tcp-sink", "head", "-1", "/nc.out")
				g.Expect(err).ToNot(HaveOccurred())
				parts := strings.Split(stdout, ",")
				g.Expect(parts).To(HaveLen(3))

				startTimeInt, err := strconv.Atoi(strings.TrimSpace(parts[0]))
				Expect(err).ToNot(HaveOccurred())
				startTime := time.Unix(int64(startTimeInt), 0)
				Expect(startTime).To(BeTemporally("~", time.Now(), time.Hour))

				src, dst = parts[1], parts[2]
			}, "30s", "1s").Should(Succeed())

			return strings.TrimSpace(src), strings.TrimSpace(dst)
		}

		E2EAfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshNameNoDefaults,
				meshaccesslog_api.MeshAccessLogResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should log outgoing traffic to MeshExternalService", func() {
			trafficLogFormat := "%START_TIME(%s)%,%KUMA_SOURCE_SERVICE%,%KUMA_DESTINATION_SERVICE%"
			meshExternalService := fmt.Sprintf(`
type: MeshExternalService
name: mes-access-log
mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 80
`, meshNameNoDefaults, esHttpContainerName)
			meshAccessLog := fmt.Sprintf(`
type: MeshAccessLog
name: mes-access-log-policy
mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: mes-access-log
      default:
        backends:
          - type: Tcp
            tcp:
              format:
                type: Plain
                plain: '%s'
              address: "%s:9999"`, meshNameNoDefaults, trafficLogFormat, tcpSinkDockerName)

			makeRequest := func(g Gomega) {
				_, err := client.CollectEchoResponse(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-access-log.extsvc.mesh.local",
				)
				g.Expect(err).ToNot(HaveOccurred())
			}

			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService))).To(Succeed())
			Expect(universal.Cluster.Install(YamlUniversal(meshAccessLog))).To(Succeed())

			// then should timeout after 5 seconds
			src, dst := expectTrafficLogged(makeRequest)
			Expect(src).To(Equal("mes-demo-client-no-defaults"))
			Expect(dst).To(Equal("mes-access-log"))
		})
	})

	Context("MeshExternalService with MeshHealthCheck", func() {
		E2EAfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshNameNoDefaults,
				meshhealthcheck_api.MeshHealthCheckResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should target real MeshExternalService resource", func() {
			meshExternalService := fmt.Sprintf(`
type: MeshExternalService
name: mes-health-check
mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 80
`, meshNameNoDefaults, esHttpContainerName)
			healthCheck := fmt.Sprintf(`
type: MeshHealthCheck
mesh: %s
name: mes-health-check-policy
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: mes-health-check
      default:
        interval: 10s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
        http:
          path: /test
          expectedStatuses:
          - 500`, meshNameNoDefaults)

			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService))).To(Succeed())

			// given no MeshHealthCheck
			By("check if service is healthy")
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-health-check.extsvc.mesh.local/test",
				)

				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal("mes-http"))
			}, "30s", "1s").Should(Succeed())

			// when MeshHealthCheck applied
			Expect(universal.Cluster.Install(YamlUniversal(healthCheck))).To(Succeed())

			// wait cluster mes-health-check to be marked as unhealthy
			Eventually(func(g Gomega) {
				egressClusters, err := universal.Cluster.GetZoneEgressEnvoyTunnel().GetClusters()
				g.Expect(err).ToNot(HaveOccurred())
				cluster := egressClusters.GetCluster(fmt.Sprintf("%s_mes-health-check__kuma-3_extsvc_80", meshNameNoDefaults))
				g.Expect(cluster).ToNot(BeNil())
				g.Expect(cluster.HostStatuses).To(HaveLen(1))
				g.Expect(cluster.HostStatuses[0].HealthStatus.FailedActiveHealthCheck).To(BeTrue())
			}, "30s", "1s").Should(Succeed())

			// check that mes-health-check is unhealthy
			Consistently(func(g Gomega) {
				response, err := client.CollectFailure(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-health-check.extsvc.mesh.local/test",
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(503))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("MeshExternalService with MeshCircuitBreaker", func() {
		E2EAfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshNameNoDefaults,
				meshcircuitbreaker_api.MeshCircuitBreakerResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should target real MeshExternalService resource", func() {
			meshExternalService := fmt.Sprintf(`
type: MeshExternalService
name: mes-circuit-breaker
mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 80
`, meshNameNoDefaults, esHttpContainerName)
			circuitBreaker := fmt.Sprintf(`
type: MeshCircuitBreaker
mesh: %s
name: mes-circuit-breaker-policy
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: mes-circuit-breaker
      default:
        connectionLimits:
          maxConnectionPools: 1
          maxConnections: 1
          maxPendingRequests: 1
          maxRequests: 1
          maxRetries: 1`, meshNameNoDefaults)

			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService))).To(Succeed())

			// given no MeshCircuitBreaker
			By("check if service is healthy")
			Eventually(func() ([]client.FailureResponse, error) {
				return client.CollectResponsesAndFailures(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-circuit-breaker.extsvc.mesh.local",
					client.WithNumberOfRequests(10),
				)
			}, "30s", "1s").Should(And(
				HaveLen(10),
				HaveEach(HaveField("ResponseCode", 200)),
			))

			// when MeshHealthCheck applied
			Expect(universal.Cluster.Install(YamlUniversal(circuitBreaker))).To(Succeed())

			By("should return 503")
			Eventually(func(g Gomega) ([]client.FailureResponse, error) {
				return client.CollectResponsesAndFailures(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-circuit-breaker.extsvc.mesh.local",
					client.WithNumberOfRequests(10),
					// increase processing time of a request to increase a probability of triggering maxPendingRequest limit
					client.WithHeader("x-set-response-delay-ms", "1000"),
					client.WithoutRetries(),
				)
			}, "30s", "1s").Should(And(
				HaveLen(10),
				ContainElement(HaveField("ResponseCode", 503)),
			))
		})
	})

	Context("MeshExternalService with MeshLoadBalancingStrategy", func() {
		E2EAfterEach(func() {
			Expect(DeleteMeshResources(universal.Cluster, meshNameNoDefaults,
				meshloadbalancingstrategy_api.MeshLoadBalancingStrategyResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should target real MeshExternalService resource", func() {
			meshExternalService := fmt.Sprintf(`
type: MeshExternalService
name: mes-load-balancing
mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 80
    - address: %s
      port: 81  
`, meshNameNoDefaults, esHttpContainerName, esHttp2ContainerName)
			loadBalancing := fmt.Sprintf(`
type: MeshLoadBalancingStrategy
mesh: %s
name: mes-load-balancing-policy
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: mes-load-balancing
      default:
        loadBalancer:
          type: RingHash
          ringHash:
            hashPolicies:
              - type: Header
                header:
                  name: x-header`, meshNameNoDefaults)

			Expect(universal.Cluster.Install(YamlUniversal(meshExternalService))).To(Succeed())

			// given no MeshLoadBalancingStrategy
			By("check if responses comes from 2 endpoints")
			Eventually(func(g Gomega) {
				responses, err := client.CollectResponsesByInstance(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-load-balancing.extsvc.mesh.local",
					client.WithHeader("x-header", "value"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(responses).To(HaveLen(2))
			}, "30s", "1s").Should(Succeed())

			// when MeshLoadBalancingStrategy applied
			Expect(universal.Cluster.Install(YamlUniversal(loadBalancing))).To(Succeed())

			By("should return responses only from 1 instance")
			Eventually(func(g Gomega) {
				responses, err := client.CollectResponsesByInstance(
					universal.Cluster, "mes-demo-client-no-defaults", "mes-load-balancing.extsvc.mesh.local",
					client.WithHeader("x-header", "value"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(responses).To(HaveLen(1))
			}, "30s", "500ms").Should(Succeed())
		})
	})

	contextFor("without default policies", meshNameNoDefaults, "mes-demo-client-no-defaults")
}
