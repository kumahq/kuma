package meshexternalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func MeshExternalServices() {
	meshName := "mesh-external-services"
	identityName := "mes-identity"
	namespace := "mesh-external-services"
	clientNamespace := "client-mesh-external-services"

	// The standalone zone CP runs under the "default" zone name.
	trustDomain := fmt.Sprintf("%s.default.mesh.local", meshName)

	egressApp := zoneproxy.EgressName(meshName)
	egressTunnel := func() envoy_admin.Tunnel {
		return kubernetes.Cluster.GetEnvoyAdminTunnel(egressApp, clientNamespace)
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlK8s(samples.MeshDefaultBuilder().
				WithName(meshName).
				KubeYaml())).
			Install(MeshIdentityBundledKubernetes(meshName, identityName)).
			Install(MeshTrafficPermissionAllowAllKubernetesWorkloadIdentity(meshName, trustDomain)).
			Install(Namespace(namespace)).
			Install(NamespaceWithSidecarInjection(clientNamespace)).
			Install(democlient.Install(democlient.WithNamespace(clientNamespace), democlient.WithMesh(meshName))).
			Install(zoneproxy.Install(
				zoneproxy.WithNamespace(clientNamespace),
				zoneproxy.WithMesh(meshName),
				zoneproxy.WithEgress(),
			)).
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
			"cluster.kri_extsvc_%s_default_%s_http-external-service_80.upstream_rq_total",
			meshName,
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
				stat, err := egressTunnel().GetStats(filter)
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
			"cluster.kri_extsvc_%s_default_%s_mesh-external-service-rbac_80.upstream_rq_total",
			meshName,
			Config.KumaNamespace,
		)

		disableMeshPassthrough := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshPassthrough
metadata:
  name: disable-default-passthrough
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    passthroughMode: None
`, Config.KumaNamespace, meshName)

		// A mesh-scoped zone egress enforces access to a MeshExternalService
		// through MeshTrafficPermission SNI rules, so forbidding the traffic is
		// a deny rule on the egress rather than a flag on the Mesh.
		denyMeshExternalService := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: deny-mesh-external-service-rbac
  namespace: %[1]s
  labels:
    kuma.io/mesh: %[2]s
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        deny:
          - sni:
              type: Exact
              value: "sni.extsvc.%[2]s.default.%[1]s.mesh-external-service-rbac.80"
`, Config.KumaNamespace, meshName)

		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithName("external-service-rbac"),
			))
			Expect(err).ToNot(HaveOccurred())
		})

		E2EAfterAll(func() {
			Expect(kubernetes.Cluster.Install(DeleteYamlK8s(disableMeshPassthrough))).To(Succeed())
			Expect(kubernetes.Cluster.Install(DeleteYamlK8s(denyMeshExternalService))).To(Succeed())
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
				stat, err := egressTunnel().GetStats(filter)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat).To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())

			// when disable all traffic
			Expect(kubernetes.Cluster.Install(YamlK8s(denyMeshExternalService))).To(Succeed())
			Expect(kubernetes.Cluster.Install(YamlK8s(disableMeshPassthrough))).To(Succeed())

			// then traffic doesn't work
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client", "mesh-external-service-rbac.extsvc.mesh.local",
					client.FromKubernetesPod(clientNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(403))
			}, "60s", "1s").Should(Succeed())
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
			"cluster.kri_extsvc_%s_default_%s_tcp-external-service_80.upstream_rq_total",
			meshName,
			Config.KumaNamespace,
		)
		BeforeAll(func() {
			err := kubernetes.Cluster.Install(testserver.Install(
				testserver.WithName("tcp-external-service"),
				testserver.WithServicePortAppProtocol("tcp"),
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
				stat, err := egressTunnel().GetStats(filter)
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
      caCert: # cat test/server/certs/server.crt
        type: InsecureInline
        insecureInline:
          value: |
            -----BEGIN CERTIFICATE-----
            MIIDNTCCAh2gAwIBAgIRAOQPGZECKKxWAk4th1ApayswDQYJKoZIhvcNAQELBQAw
            GzEZMBcGA1UEAxMQdGVzdC1zZXJ2ZXIubWVzaDAeFw0yMjA4MDkxNTAzNTNaFw0z
            MjA4MDYxNTAzNTNaMBsxGTAXBgNVBAMTEHRlc3Qtc2VydmVyLm1lc2gwggEiMA0G
            CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDceQX+8ioPAYFl1hT7n0cfHOLVbYt7
            kMwxZf7936o8v9QuaP2V4umilZ5PxXbd0zB4CqMIk/KTdyfYJAH1fmtck9bYTq9/
            T+0Wi4jtLwNbcjp7x+SkB6Z2qJybdMWXPZX638cN+QQPKBt23L9+Jnu8j8t93mJL
            HmLtEcBoCeiZs+NuumelmXVz2thcotawOwTwqhCpnOcr4KWGupAWS9YxDcDWZwk6
            iKc+QUYIJsiL9jUqvX5eEv+2CzPzq7I42li6w3lhoVVZHu5Z6hvRbiJ5D0NQBWW0
            D9UfpMO8PQiINH9Ylo03EG0D6K3qRb38XJg8h5O+g5qxSn0n604KNO9VAgMBAAGj
            dDByMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMB
            Af8EBTADAQH/MB0GA1UdDgQWBBQfpXjl+rF4WJDq4yGQ99a/fHbAsjAbBgNVHREE
            FDASghB0ZXN0LXNlcnZlci5tZXNoMA0GCSqGSIb3DQEBCwUAA4IBAQCOx7DvHa8d
            UvRhqPjvjTiXVY1NvnScUa5LzwzzGZF6dz7atHPRLhZI0WkBMH47Pb7P2ULbxojG
            rfnVgq3wKgv9E5JN8LCSTju494/mggGaQMyETnwYF/PGcwmIUuvEYjywzZsqu3OI
            A1w8GshIIzgucVaB6iWo0Svzf1CUO1tZ+dnU5Bd3M60/SnCWxFjnNfzYhSLtmd9O
            8ohUvFreMJ/75BElIeG4fD65ewr3CQ/mcLNyqlQYeAqUOmAY+TzDmgtGsBJssBQW
            u1Ko98j0EvLAaLL8vxSFn7hqThNcyGpACupQnyEmS5hQ3CbMLY8cSxhnfmLykSRS
            j4N/r2sTBVzB
            -----END CERTIFICATE-----
      clientCert: # cat test/server/certs/client.crt
        type: InsecureInline
        insecureInline:
          value: |
            -----BEGIN CERTIFICATE-----
            MIIDKzCCAhOgAwIBAgIUFwpTPaozHwEuuP2WuYMSHPjyT/4wDQYJKoZIhvcNAQEL
            BQAwGzEZMBcGA1UEAxMQdGVzdC1zZXJ2ZXIubWVzaDAeFw0yNDEyMDQxMDExNDNa
            Fw0zNDEyMDIxMDExNDNaMEAxCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApTb21lLVN0
            YXRlMQ0wCwYDVQQKDARLdW1hMQ0wCwYDVQQLDAR0ZXN0MIIBIjANBgkqhkiG9w0B
            AQEFAAOCAQ8AMIIBCgKCAQEA7Hpi0oL5134xep2yR9m8lPdSblue84A55ExGYrza
            vsj/vYmMA58qW33CbcblREGHBA9uEtS94SzD4L2h1/03DzLZR9nc944WDD7KZxvh
            u6xheqNXma6fLUPKXeUZb0ciCmx9ewLyzAu1CA+E0+u2P6XEiZajWaQiZqD5bIkH
            DMFf6h2yHsPcN4fZPPa3/R05AzQ2VJIhOK+XjVHJrztohVyGLdxwmreUauiUrY5v
            g6XD8ske7W6oEPAJA+Cu+j+5wKmMNZ2gmYKnyWuXZZglh5UZZ0NHCqNSnJXb7Du2
            QP/HFo1FyL5u6DeqSN1Lg7JHRbqX2bL8wEyKXy2ExJl1iQIDAQABo0IwQDAdBgNV
            HQ4EFgQUOWDoOTaudmbfgyW7xWnRj13b8gYwHwYDVR0jBBgwFoAUH6V45fqxeFiQ
            6uMhkPfWv3x2wLIwDQYJKoZIhvcNAQELBQADggEBALhtTrpZyTQHFYaAUCsyA8kq
            0rfV+P+WnijY+NubUX5n69anMjl0JAFgM34qnrn8C4KoP2O8LteG7eVMyqDZXngn
            qrulukcvoBMyK/V3rFGtQIMJmUWpnRLRkItCrP8sN2NTM0iJ0JPF4VVW2sRpr2fn
            W/2pcJ0T+POgKyVH36WDjxC5Y2tqRTRZHeZYbx1RzJZ38Ggu76qzieTiO53HQTU+
            WxJWuFq4NbllLu0r0uEm1VIk6RY6SH/H/DqWIvWCMxUgHEXY+udG7kyB7roFRCgX
            b6BbZTTkEVtBEL68MdJEesIq5RTHPWLZXeJWFyegxC9tGNqZOGh9FyIlApF7SfE=
            -----END CERTIFICATE-----
      clientKey: # cat test/server/certs/client.key
        type: InsecureInline
        insecureInline:
          value: |
            -----BEGIN PRIVATE KEY-----
            MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDsemLSgvnXfjF6
            nbJH2byU91JuW57zgDnkTEZivNq+yP+9iYwDnypbfcJtxuVEQYcED24S1L3hLMPg
            vaHX/TcPMtlH2dz3jhYMPspnG+G7rGF6o1eZrp8tQ8pd5RlvRyIKbH17AvLMC7UI
            D4TT67Y/pcSJlqNZpCJmoPlsiQcMwV/qHbIew9w3h9k89rf9HTkDNDZUkiE4r5eN
            UcmvO2iFXIYt3HCat5Rq6JStjm+DpcPyyR7tbqgQ8AkD4K76P7nAqYw1naCZgqfJ
            a5dlmCWHlRlnQ0cKo1KcldvsO7ZA/8cWjUXIvm7oN6pI3UuDskdFupfZsvzATIpf
            LYTEmXWJAgMBAAECggEABQOFQ9xWCr0YtHZSeNaDeo8R1tgncxc1YwNA/MfvRVtC
            nNSlPNBrl/v/Gs+8Pam8AJiJJ2oOSo9l6cZrf4ZVXAOierUCS9dd3U2Zgf0j2JRL
            jsuWyGHc6xtEV6BLXUIfVSQ+ttR1rGDVKkIV+V5Gg2vy0k340aY6un1QPH5dQWZv
            wKiaXZxZpvsjdmAMS0wJnVhkfZsPWib/sPYl7nGp5BKjUIQ60dHYZKasu3wffbkQ
            A74UKveC8/xkoHjkcBaGPfNiFphhxwd/jVRjNnIPyuDmENfOrJKotIYTftdUUNEO
            +AXy/xglLmmAWhZdQtYBY+CyatpgebDdBC/pvMu1wQKBgQD3wdRXnMnyTyaVah2a
            VF4TysosNu6ReB6ktzmxnRtiecS2j7fsJo/LbFwpYfen2MD9rasgFEAvOrlElwNd
            2tWyloMHp0+km7XD1Izts88mPClxLVgsp97MezrU4/Nf2pYhy46I/Wdp7nFs+4Vg
            +1SoCA+qn3bBnAD+2YTuynyvmQKBgQD0WH3bb4M3Iou+EEo1FggHNgSN2ND79Bak
            isUdNmcUAP5z7i+xrN7hIVtu28hYOzwm8HEyJodhzrBAiPUZq6CAfZZvjn6OBGHV
            KDECh0R8T86ajEvq3832BZNmqufcCdsVkk+FynwXweqT+F6ToItFib4IfFKlzbuG
            K3yRDmNrcQKBgCGPCqEXZq9Ak1xXtEzMMrYBmOLmSehAWf47pz/spOHw1nlX/DSr
            gHywX8dnMrF0haeW14AP3iXHkYK95cHXu2xmQLdPrVUBllxBNRmZamymZ4Kh/riF
            wIL4Ch7+BWAtbnqDZPofQNuzZX+6jfV19aCQ/vZAhUhyRhw/AGeL29m5AoGAcXJU
            nPldVs/3SbuOeK9N8uslmiY8gX6GtMapVjLYEPWVLoY8JqY4pRYzuXjZv/1gpEOm
            ir5QxRyNwKjWA6En2AB3RDxIje+C7NDIUIA1T/JN3nudE+PtYHieQ2C+Xe9FhPJ1
            cYzdqLokC6eZYbl8cEDPtmjihpDKrDSslTy09EECgYBufGjG6s+CrXh2r4/eooFz
            c99YHXUR/4kvpPSjyn0yiBmO25yysSwDc9IuVSZoK5/Wr7pTghsqyA43kjuljRiJ
            OBAD4nGg7KH07SDgHWbQoU04dBUeppK6KOmlSOOg1hxctRrng8gmjksOX3lEzQI1
            Lpqgzv8I1uEJ9GN8cu8rPA==
            -----END PRIVATE KEY-----
    version:
      min: %s
      max: %s
`, Config.KumaNamespace, meshName, tlsVersion, tlsVersion)
		}
		filter := func(serviceName string) string {
			return fmt.Sprintf(
				"cluster.kri_extsvc_%s_default_%s_%s_80.upstream_rq_total", // cx
				meshName,
				Config.KumaNamespace,
				serviceName,
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
				stat, err := egressTunnel().GetStats(filter("tls-external-service"))
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
				stat, err := egressTunnel().GetStats(filter("tls13-external-service"))
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
        labels:
          kuma.io/display-name: plain-external-service
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /
          default:
            backendRefs:
              - kind: MeshExternalService
                labels:
                  kuma.io/display-name: external-service-with-httproute
                  k8s.kuma.io/namespace: %s
                port: 80
                weight: 100
`, Config.KumaNamespace, meshName, Config.KumaNamespace)

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
