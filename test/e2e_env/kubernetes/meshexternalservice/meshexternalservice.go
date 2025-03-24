package meshexternalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshExternalServices() {
	meshName := "mesh-external-services"
	namespace := "mesh-external-services"
	clientNamespace := "client-mesh-external-services"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(samples.MeshMTLSBuilder().
				WithName(meshName).
				WithEgressRoutingEnabled().KubeYaml())).
			Install(Namespace(namespace)).
			Install(NamespaceWithSidecarInjection(clientNamespace)).
			Install(democlient.Install(democlient.WithNamespace(clientNamespace), democlient.WithMesh(meshName))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace, clientNamespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	Context("http non-TLS", func() {
		meshExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: http-external-service
  namespace: %s
  labels:
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: external-service.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName)

		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithName("external-service"),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		filter := fmt.Sprintf(
			"cluster.%s_%s_%s_default_extsvc_80.upstream_rq_total",
			meshName,
			"http-external-service",
			Config.KumaNamespace,
		)

		It("should route to http external-service", func() {
			// given working communication outside the mesh with passthrough enabled and no traffic permission
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "external-service.mesh-external-services",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// when apply external service
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())

			// and you can also use .mesh on port of the provided host
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "http-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// and flows through Egress
			Eventually(func(g Gomega) {
				stat, err := kubernetes.Cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat).To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("http non-TLS with rbac switch", func() {
		meshExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: mesh-external-service-rbac
  namespace: %s
  labels:
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: external-service-rbac.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName)

		filter := fmt.Sprintf(
			"cluster.%s_%s_%s_default_extsvc_80.upstream_rq_total",
			meshName,
			"mesh-external-service-rbac",
			Config.KumaNamespace,
		)
		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithName("external-service-rbac"),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			Expect(kubernetes.Cluster.Install(YamlK8s(
				samples.MeshMTLSBuilder().
					WithName(meshName).
					WithEgressRoutingEnabled().KubeYaml()),
			)).To(Succeed())
		})

		It("should route to external-service", func() {
			// when apply external service and hostname generator
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())

			// then traffic work
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "mesh-external-service-rbac.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// and flows through Egress
			Eventually(func(g Gomega) {
				stat, err := kubernetes.Cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat).To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())

			// when disable all traffic
			Expect(kubernetes.Cluster.Install(YamlK8s(
				samples.MeshMTLSBuilder().
					WithName(meshName).
					WithoutPassthrough().
					WithMeshExternalServiceTrafficForbidden().
					WithEgressRoutingEnabled().KubeYaml()),
			)).To(Succeed())

			// then traffic doesn't work
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client", "mesh-external-service-rbac.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(403))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("tcp non-TLS", func() {
		meshExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: tcp-external-service
  namespace: %s
  labels:
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: tcp
  endpoints:
    - address: tcp-external-service.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName)
		filter := fmt.Sprintf(
			"cluster.%s_%s_%s_default_extsvc_80.upstream_rq_total",
			meshName,
			"tcp-external-service",
			Config.KumaNamespace,
		)
		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithName("tcp-external-service"),
				testserver.WithServicePortAppProtocol("tcp"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to tcp external-service", func() {
			// given working communication outside the mesh with passthrough enabled and no traffic permission
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tcp-external-service.mesh-external-services",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// when apply external service
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())

			// and you can also use .mesh on port of the provided host
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tcp-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// and flows through Egress
			Eventually(func(g Gomega) {
				stat, err := kubernetes.Cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat).To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("HTTPS", func() {
		tlsExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: tls-external-service
  namespace: %s
  labels:
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: tls-external-service.mesh-external-services.svc.cluster.local
      port: 80
  tls:
    enabled: true
    verification:
      mode: SkipCA # test-server certificate is not signed by a CA that is in the system trust store
`, Config.KumaNamespace, meshName)
		tlsVersionExternalService := func(tlsVersion string) string {
			return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: tls13-external-service
  namespace: %s
  labels:
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: tls13-external-service.mesh-external-services.svc.cluster.local
      port: 80
  tls:
    enabled: true
    verification:
      mode: SkipSAN
      caCert: # cat test/server/certs/server.crt | base64
        inline: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUROVENDQWgyZ0F3SUJBZ0lSQU9RUEdaRUNLS3hXQWs0dGgxQXBheXN3RFFZSktvWklodmNOQVFFTEJRQXcKR3pFWk1CY0dBMVVFQXhNUWRHVnpkQzF6WlhKMlpYSXViV1Z6YURBZUZ3MHlNakE0TURreE5UQXpOVE5hRncwegpNakE0TURZeE5UQXpOVE5hTUJzeEdUQVhCZ05WQkFNVEVIUmxjM1F0YzJWeWRtVnlMbTFsYzJnd2dnRWlNQTBHCkNTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFEY2VRWCs4aW9QQVlGbDFoVDduMGNmSE9MVmJZdDcKa013eFpmNzkzNm84djlRdWFQMlY0dW1pbFo1UHhYYmQwekI0Q3FNSWsvS1RkeWZZSkFIMWZtdGNrOWJZVHE5LwpUKzBXaTRqdEx3TmJjanA3eCtTa0I2WjJxSnliZE1XWFBaWDYzOGNOK1FRUEtCdDIzTDkrSm51OGo4dDkzbUpMCkhtTHRFY0JvQ2VpWnMrTnV1bWVsbVhWejJ0aGNvdGF3T3dUd3FoQ3BuT2NyNEtXR3VwQVdTOVl4RGNEV1p3azYKaUtjK1FVWUlKc2lMOWpVcXZYNWVFdisyQ3pQenE3STQybGk2dzNsaG9WVlpIdTVaNmh2UmJpSjVEME5RQldXMApEOVVmcE1POFBRaUlOSDlZbG8wM0VHMEQ2SzNxUmIzOFhKZzhoNU8rZzVxeFNuMG42MDRLTk85VkFnTUJBQUdqCmREQnlNQTRHQTFVZER3RUIvd1FFQXdJQ3BEQVRCZ05WSFNVRUREQUtCZ2dyQmdFRkJRY0RBVEFQQmdOVkhSTUIKQWY4RUJUQURBUUgvTUIwR0ExVWREZ1FXQkJRZnBYamwrckY0V0pEcTR5R1E5OWEvZkhiQXNqQWJCZ05WSFJFRQpGREFTZ2hCMFpYTjBMWE5sY25abGNpNXRaWE5vTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFDT3g3RHZIYThkClV2UmhxUGp2alRpWFZZMU52blNjVWE1THp3enpHWkY2ZHo3YXRIUFJMaFpJMFdrQk1INDdQYjdQMlVMYnhvakcKcmZuVmdxM3dLZ3Y5RTVKTjhMQ1NUanU0OTQvbWdnR2FRTXlFVG53WUYvUEdjd21JVXV2RVlqeXd6WnNxdTNPSQpBMXc4R3NoSUl6Z3VjVmFCNmlXbzBTdnpmMUNVTzF0WitkblU1QmQzTTYwL1NuQ1d4RmpuTmZ6WWhTTHRtZDlPCjhvaFV2RnJlTUovNzVCRWxJZUc0ZkQ2NWV3cjNDUS9tY0xOeXFsUVllQXFVT21BWStUekRtZ3RHc0JKc3NCUVcKdTFLbzk4ajBFdkxBYUxMOHZ4U0ZuN2hxVGhOY3lHcEFDdXBRbnlFbVM1aFEzQ2JNTFk4Y1N4aG5mbUx5a1NSUwpqNE4vcjJzVEJWekIKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
      clientCert: # cat test/server/certs/client.crt | base64
        inline: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURLekNDQWhPZ0F3SUJBZ0lVRndwVFBhb3pId0V1dVAyV3VZTVNIUGp5VC80d0RRWUpLb1pJaHZjTkFRRUwKQlFBd0d6RVpNQmNHQTFVRUF4TVFkR1Z6ZEMxelpYSjJaWEl1YldWemFEQWVGdzB5TkRFeU1EUXhNREV4TkROYQpGdzB6TkRFeU1ESXhNREV4TkROYU1FQXhDekFKQmdOVkJBWVRBbFZUTVJNd0VRWURWUVFJREFwVGIyMWxMVk4wCllYUmxNUTB3Q3dZRFZRUUtEQVJMZFcxaE1RMHdDd1lEVlFRTERBUjBaWE4wTUlJQklqQU5CZ2txaGtpRzl3MEIKQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBN0hwaTBvTDUxMzR4ZXAyeVI5bThsUGRTYmx1ZTg0QTU1RXhHWXJ6YQp2c2ovdlltTUE1OHFXMzNDYmNibFJFR0hCQTl1RXRTOTRTekQ0TDJoMS8wM0R6TFpSOW5jOTQ0V0REN0taeHZoCnU2eGhlcU5YbWE2ZkxVUEtYZVVaYjBjaUNteDlld0x5ekF1MUNBK0UwK3UyUDZYRWlaYWpXYVFpWnFENWJJa0gKRE1GZjZoMnlIc1BjTjRmWlBQYTMvUjA1QXpRMlZKSWhPSytYalZISnJ6dG9oVnlHTGR4d21yZVVhdWlVclk1dgpnNlhEOHNrZTdXNm9FUEFKQStDdStqKzV3S21NTloyZ21ZS255V3VYWlpnbGg1VVpaME5IQ3FOU25KWGI3RHUyClFQL0hGbzFGeUw1dTZEZXFTTjFMZzdKSFJicVgyYkw4d0V5S1h5MkV4SmwxaVFJREFRQUJvMEl3UURBZEJnTlYKSFE0RUZnUVVPV0RvT1RhdWRtYmZneVc3eFduUmoxM2I4Z1l3SHdZRFZSMGpCQmd3Rm9BVUg2VjQ1ZnF4ZUZpUQo2dU1oa1BmV3YzeDJ3TEl3RFFZSktvWklodmNOQVFFTEJRQURnZ0VCQUxodFRycFp5VFFIRllhQVVDc3lBOGtxCjByZlYrUCtXbmlqWStOdWJVWDVuNjlhbk1qbDBKQUZnTTM0cW5ybjhDNEtvUDJPOEx0ZUc3ZVZNeXFEWlhuZ24KcXJ1bHVrY3ZvQk15Sy9WM3JGR3RRSU1KbVVXcG5STFJrSXRDclA4c04yTlRNMGlKMEpQRjRWVlcyc1JwcjJmbgpXLzJwY0owVCtQT2dLeVZIMzZXRGp4QzVZMnRxUlRSWkhlWllieDFSekpaMzhHZ3U3NnF6aWVUaU81M0hRVFUrCld4Sld1RnE0TmJsbEx1MHIwdUVtMVZJazZSWTZTSC9IL0RxV0l2V0NNeFVnSEVYWSt1ZEc3a3lCN3JvRlJDZ1gKYjZCYlpUVGtFVnRCRUw2OE1kSkVlc0lxNVJUSFBXTFpYZUpXRnllZ3hDOXRHTnFaT0doOUZ5SWxBcEY3U2ZFPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
      clientKey: # cat test/server/certs/client.key | base64
        inline: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2QUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktZd2dnU2lBZ0VBQW9JQkFRRHNlbUxTZ3ZuWGZqRjYKbmJKSDJieVU5MUp1VzU3emdEbmtURVppdk5xK3lQKzlpWXdEbnlwYmZjSnR4dVZFUVljRUQyNFMxTDNoTE1QZwp2YUhYL1RjUE10bEgyZHozamhZTVBzcG5HK0c3ckdGNm8xZVpycDh0UThwZDVSbHZSeUlLYkgxN0F2TE1DN1VJCkQ0VFQ2N1kvcGNTSmxxTlpwQ0ptb1Bsc2lRY013Vi9xSGJJZXc5dzNoOWs4OXJmOUhUa0RORFpVa2lFNHI1ZU4KVWNtdk8yaUZYSVl0M0hDYXQ1UnE2SlN0am0rRHBjUHl5Ujd0YnFnUThBa0Q0Szc2UDduQXFZdzFuYUNaZ3FmSgphNWRsbUNXSGxSbG5RMGNLbzFLY2xkdnNPN1pBLzhjV2pVWEl2bTdvTjZwSTNVdURza2RGdXBmWnN2ekFUSXBmCkxZVEVtWFdKQWdNQkFBRUNnZ0VBQlFPRlE5eFdDcjBZdEhaU2VOYURlbzhSMXRnbmN4YzFZd05BL01mdlJWdEMKbk5TbFBOQnJsL3YvR3MrOFBhbThBSmlKSjJvT1NvOWw2Y1pyZjRaVlhBT2llclVDUzlkZDNVMlpnZjBqMkpSTApqc3VXeUdIYzZ4dEVWNkJMWFVJZlZTUSt0dFIxckdEVktrSVYrVjVHZzJ2eTBrMzQwYVk2dW4xUVBINWRRV1p2CndLaWFYWnhacHZzamRtQU1TMHdKblZoa2Zac1BXaWIvc1BZbDduR3A1QktqVUlRNjBkSFlaS2FzdTN3ZmZia1EKQTc0VUt2ZUM4L3hrb0hqa2NCYUdQZk5pRnBoaHh3ZC9qVlJqTm5JUHl1RG1FTmZPckpLb3RJWVRmdGRVVU5FTworQVh5L3hnbExtbUFXaFpkUXRZQlkrQ3lhdHBnZWJEZEJDL3B2TXUxd1FLQmdRRDN3ZFJYbk1ueVR5YVZhaDJhClZGNFR5c29zTnU2UmVCNmt0em14blJ0aWVjUzJqN2ZzSm8vTGJGd3BZZmVuMk1EOXJhc2dGRUF2T3JsRWx3TmQKMnRXeWxvTUhwMCtrbTdYRDFJenRzODhtUENseExWZ3NwOTdNZXpyVTQvTmYycFloeTQ2SS9XZHA3bkZzKzRWZworMVNvQ0ErcW4zYkJuQUQrMllUdXlueXZtUUtCZ1FEMFdIM2JiNE0zSW91K0VFbzFGZ2dITmdTTjJORDc5QmFrCmlzVWRObWNVQVA1ejdpK3hyTjdoSVZ0dTI4aFlPendtOEhFeUpvZGh6ckJBaVBVWnE2Q0FmWlp2am42T0JHSFYKS0RFQ2gwUjhUODZhakV2cTM4MzJCWk5tcXVmY0Nkc1ZraytGeW53WHdlcVQrRjZUb0l0RmliNElmRktsemJ1RwpLM3lSRG1OcmNRS0JnQ0dQQ3FFWFpxOUFrMXhYdEV6TU1yWUJtT0xtU2VoQVdmNDdwei9zcE9IdzFubFgvRFNyCmdIeXdYOGRuTXJGMGhhZVcxNEFQM2lYSGtZSzk1Y0hYdTJ4bVFMZFByVlVCbGx4Qk5SbVphbXltWjRLaC9yaUYKd0lMNENoNytCV0F0Ym5xRFpQb2ZRTnV6WlgrNmpmVjE5YUNRL3ZaQWhVaHlSaHcvQUdlTDI5bTVBb0dBY1hKVQpuUGxkVnMvM1NidU9lSzlOOHVzbG1pWThnWDZHdE1hcFZqTFlFUFdWTG9ZOEpxWTRwUll6dVhqWnYvMWdwRU9tCmlyNVF4UnlOd0tqV0E2RW4yQUIzUkR4SWplK0M3TkRJVUlBMVQvSk4zbnVkRStQdFlIaWVRMkMrWGU5RmhQSjEKY1l6ZHFMb2tDNmVaWWJsOGNFRFB0bWppaHBES3JEU3NsVHkwOUVFQ2dZQnVmR2pHNnMrQ3JYaDJyNC9lb29GegpjOTlZSFhVUi80a3ZwUFNqeW4weWlCbU8yNXl5c1N3RGM5SXVWU1pvSzUvV3I3cFRnaHNxeUE0M2tqdWxqUmlKCk9CQUQ0bkdnN0tIMDdTRGdIV2JRb1UwNGRCVWVwcEs2S09tbFNPT2cxaHhjdFJybmc4Z21qa3NPWDNsRXpRSTEKTHBxZ3p2OEkxdUVKOUdOOGN1OHJQQT09Ci0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K
    version:
      min: %s
      max: %s
`, Config.KumaNamespace, meshName, tlsVersion, tlsVersion)
		}
		filter := func(serviceName string) string {
			return fmt.Sprintf(
				"cluster.%s_%s_%s_default_extsvc_80.upstream_rq_total", // cx
				meshName,
				serviceName,
				Config.KumaNamespace,
			)
		}
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(Parallel(
					testserver.Install(
						testserver.WithNamespace(namespace),
						testserver.WithEchoArgs("--tls", "--crt=/kuma/server.crt", "--key=/kuma/server.key"),
						testserver.WithName("tls-external-service"),
						testserver.WithoutProbes()), // not compatible with TLS
					testserver.Install(
						testserver.WithNamespace(namespace),
						testserver.WithEchoArgs("--tls", "--crt=/kuma/server.crt", "--key=/kuma/server.key", "--tls13", "--instance", "tls13-server"),
						testserver.WithName("tls13-external-service"),
						testserver.WithoutProbes(), // not compatible with TLS
					))).
				Install(YamlK8s(tlsExternalService)).
				Install(YamlK8s(tlsVersionExternalService("TLS12"))).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should route to tls external-service", func() {
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tls-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// and flows through Egress
			Eventually(func(g Gomega) {
				stat, err := kubernetes.Cluster.GetZoneEgressEnvoyTunnel().GetStats(filter("tls-external-service"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat).To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})

		It("should route to tls 1.3 external-service", func() {
			// requests should fail because service is TLS13 but configuration uses TLS12
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client", "tls13-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(503))
			}, "30s", "1s", MustPassRepeatedly(3)).Should(Succeed())

			// when MES defined with correct TLS version
			Expect(kubernetes.Cluster.Install(YamlK8s(tlsVersionExternalService("TLS13")))).To(Succeed())

			Eventually(func(g Gomega) {
				resp, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "tls13-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("tls13-server"))
			}, "30s", "1s", MustPassRepeatedly(3)).Should(Succeed())

			// and flows through Egress
			Eventually(func(g Gomega) {
				stat, err := kubernetes.Cluster.GetZoneEgressEnvoyTunnel().GetStats(filter("tls13-external-service"))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat).To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("http service with MeshHTTPRoute", func() {
		meshExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: plain-external-service
  namespace: %s
  labels:
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: plain-external-service.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName)

		meshExternalService2 := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: external-service-with-httproute
  namespace: %s
  labels:
    kuma.io/mesh: %s
    hostname: "true"
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: external-service-with-httproute.mesh-external-services.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName)

		meshHttpRoutePolicy := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route-mes-policy
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshExternalService
        name: plain-external-service
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /
          default:
            backendRefs:
              - kind: MeshExternalService
                name: external-service-with-httproute
                port: 80
                weight: 100
`, Config.KumaNamespace, meshName)

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithName("plain-external-service"),
					testserver.WithEchoArgs("echo", "--instance", "plain-external-service"),
				)).
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithName("external-service-with-httproute"),
					testserver.WithEchoArgs("echo", "--instance", "external-service-with-httproute"),
				)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		E2EAfterEach(func() {
			Expect(DeleteMeshResources(kubernetes.Cluster, meshName,
				meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should route to http external-service", func() {
			// when external service added
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService))).To(Succeed())

			// then communication works
			Eventually(func(g Gomega) {
				resp, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "plain-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("plain-external-service"))
			}, "30s", "1s").Should(Succeed())

			// when
			Expect(kubernetes.Cluster.Install(YamlK8s(meshExternalService2))).To(Succeed())

			// and added route
			Expect(kubernetes.Cluster.Install(YamlK8s(meshHttpRoutePolicy))).To(Succeed())

			// then traffic is routed to the 2nd MeshExternalService
			Eventually(func(g Gomega) {
				resp, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client", "plain-external-service.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.Instance).To(Equal("external-service-with-httproute"))
			}, "30s", "1s").Should(Succeed())
		})
	})
}
