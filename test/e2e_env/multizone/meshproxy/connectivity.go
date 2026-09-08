package meshproxy

import (
	"fmt"
	"strings"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_tls_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/config_dump"
	"github.com/kumahq/kuma/v3/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func Connectivity() {
	namespace := "meshproxy"
	externalNamespace := "meshproxy-ext"
	meshName := "meshproxy"
	identityName := "meshproxy-identity"

	ingressApp := zoneproxy.IngressName(meshName)
	egressApp := zoneproxy.EgressName(meshName)

	zoneProxies := func(opts ...zoneproxy.DeploymentOptsFn) InstallFunc {
		return zoneproxy.Install(append([]zoneproxy.DeploymentOptsFn{zoneproxy.WithMesh(meshName)}, opts...)...)
	}

	BeforeAll(func() {
		zones := []Cluster{multizone.KubeZone1, multizone.KubeZone2, multizone.UniZone1}
		Expect(NewClusterSetup().
			Install(Yaml(
				builders.Mesh().
					WithName(meshName),
			)).
			Install(MeshIdentityBundled(meshName, identityName)).
			Install(MeshTrafficPermissionAllowAllUniversalWorkloadIdentity(meshName, MeshIdentityTrustDomains(meshName, zones...)...)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(externalNamespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-1"),
				),
				democlient.Install(
					democlient.WithNamespace(namespace),
					democlient.WithMesh(meshName),
				),
				testserver.Install(
					testserver.WithNamespace(externalNamespace),
					testserver.WithName("external-service"),
					testserver.WithEchoArgs("echo", "--instance", "kube-external-service"),
				),
				zoneProxies(zoneproxy.WithNamespace(namespace)),
			)).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
				),
				zoneProxies(zoneproxy.WithNamespace(namespace)),
			)).
			SetupInGroup(multizone.KubeZone2, &group)

		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true), WithWorkload("demo-client")),
				TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"}), WithWorkload("test-server")),
				TestServerExternalServiceUniversal(fmt.Sprintf("external-service-%s", meshName), 8080, false, WithDockerContainerName("kuma-es-4_external-service-meshproxy")),
				zoneProxies(),
			)).
			SetupInGroup(multizone.UniZone1, &group)
		Expect(group.Wait()).To(Succeed())

		// Register Envoy admin tunnel for demo-client so tests can inspect its xDS config.
		Expect(multizone.UniZone1.RegisterAppEnvoyTunnel("demo-client")).To(Succeed())

		// MeshExternalService for the external server running in UniZone1.
		extServiceIP := multizone.UniZone1.GetApp("external-service-meshproxy").GetIP()
		Expect(NewClusterSetup().
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshExternalService
name: external-service-meshproxy
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 8080
    protocol: http
  endpoints:
    - address: %s
      port: 8080
`, meshName, extServiceIP))).
			Setup(multizone.UniZone1)).To(Succeed())

		// MeshExternalService for the external server running in KubeZone1.
		Expect(NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: external-service-kube
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: external-service.%s.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName, externalNamespace))).
			Setup(multizone.KubeZone1)).To(Succeed())

		Expect(DistributeMeshTrusts(multizone.Global, meshName, identityName, zones...)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(externalNamespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should route cross-zone traffic via new zone ingress proxies and record metrics on zone-proxy-ingress", func() {
		uniIngressFilter := fmt.Sprintf(
			"cluster.kri_msvc_%s_%s__test-server_80.upstream_rq_total",
			meshName, multizone.UniZone1.ZoneName(),
		)
		kubeZone1IngressFilter := fmt.Sprintf(
			"cluster.kri_msvc_%s_%s_%s_test-server_main.upstream_rq_total",
			meshName, multizone.KubeZone1.ZoneName(), namespace,
		)
		kubeZone2IngressFilter := fmt.Sprintf(
			"cluster.kri_msvc_%s_%s_%s_test-server_main.upstream_rq_total",
			meshName, multizone.KubeZone2.ZoneName(), namespace,
		)

		// Kubernetes client -> Universal zone via zone-proxy-ingress
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client",
				fmt.Sprintf("http://test-server.svc.%s.mesh.local:80", multizone.UniZone1.ZoneName()),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("uni-test-server"))
		}, "30s", "1s").Should(Succeed())

		// Verify traffic was routed through UniZone1's zone-proxy-ingress
		Eventually(func(g Gomega) {
			stat, err := multizone.UniZone1.GetAppEnvoyTunnel(ingressApp).GetStats(uniIngressFilter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// Universal client -> Kubernetes zone 1 via zone-proxy-ingress
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.UniZone1, "demo-client",
				fmt.Sprintf("http://test-server.%s.svc.%s.mesh.local:80", namespace, multizone.KubeZone1.ZoneName()),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kube-test-server-1"))
		}, "30s", "1s").Should(Succeed())

		// Verify traffic was routed through KubeZone1's zone-proxy-ingress
		Eventually(func(g Gomega) {
			stat, err := multizone.KubeZone1.GetEnvoyAdminTunnel(ingressApp, namespace).GetStats(kubeZone1IngressFilter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// Universal client -> Kubernetes zone 2 via zone-proxy-ingress
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.UniZone1, "demo-client",
				fmt.Sprintf("http://test-server.%s.svc.%s.mesh.local:80", namespace, multizone.KubeZone2.ZoneName()),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kube-test-server-2"))
		}, "30s", "1s").Should(Succeed())

		// Verify traffic was routed through KubeZone2's zone-proxy-ingress
		Eventually(func(g Gomega) {
			stat, err := multizone.KubeZone2.GetEnvoyAdminTunnel(ingressApp, namespace).GetStats(kubeZone2IngressFilter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// Verify UniZone1 ingress filterchain SNI
		uniIngressSNI := fmt.Sprintf("sni.msvc.%s.%s.test-server.80", meshName, multizone.UniZone1.ZoneName())
		Eventually(func(g Gomega) {
			xds, err := multizone.UniZone1.GetAppEnvoyTunnel(ingressApp).GetConfigDump()
			g.Expect(err).ToNot(HaveOccurred())
			snis, err := listenerFilterchainSNIs(xds, "zoneingress")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(snis).To(ContainElement(uniIngressSNI))
		}, "30s", "1s").Should(Succeed())

		// Verify KubeZone1 ingress filterchain SNI
		kubeZone1IngressSNI := fmt.Sprintf("sni.msvc.%s.%s.%s.test-server.main", meshName, multizone.KubeZone1.ZoneName(), namespace)
		Eventually(func(g Gomega) {
			xds, err := multizone.KubeZone1.GetEnvoyAdminTunnel(ingressApp, namespace).GetConfigDump()
			g.Expect(err).ToNot(HaveOccurred())
			snis, err := listenerFilterchainSNIs(xds, "zoneingress")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(snis).To(ContainElement(kubeZone1IngressSNI))
		}, "30s", "1s").Should(Succeed())

		// Verify KubeZone2 ingress filterchain SNI
		kubeZone2IngressSNI := fmt.Sprintf("sni.msvc.%s.%s.%s.test-server.main", meshName, multizone.KubeZone2.ZoneName(), namespace)
		Eventually(func(g Gomega) {
			xds, err := multizone.KubeZone2.GetEnvoyAdminTunnel(ingressApp, namespace).GetConfigDump()
			g.Expect(err).ToNot(HaveOccurred())
			snis, err := listenerFilterchainSNIs(xds, "zoneingress")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(snis).To(ContainElement(kubeZone2IngressSNI))
		}, "30s", "1s").Should(Succeed())

		// Verify UniZone1 demo-client cluster SNI for KubeZone1 test-server
		Eventually(func(g Gomega) {
			xds, err := multizone.UniZone1.GetAppEnvoyTunnel("demo-client").GetConfigDump()
			g.Expect(err).ToNot(HaveOccurred())
			sni, err := clusterSNI(xds, fmt.Sprintf("kri_msvc_%s_%s_%s_test-server_main", meshName, multizone.KubeZone1.ZoneName(), namespace))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(sni).To(Equal(kubeZone1IngressSNI))
		}, "30s", "1s").Should(Succeed())

		// Verify UniZone1 demo-client cluster SNI for KubeZone2 test-server
		Eventually(func(g Gomega) {
			xds, err := multizone.UniZone1.GetAppEnvoyTunnel("demo-client").GetConfigDump()
			g.Expect(err).ToNot(HaveOccurred())
			sni, err := clusterSNI(xds, fmt.Sprintf("kri_msvc_%s_%s_%s_test-server_main", meshName, multizone.KubeZone2.ZoneName(), namespace))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(sni).To(Equal(kubeZone2IngressSNI))
		}, "30s", "1s").Should(Succeed())
	})

	It("should route traffic to MeshExternalService and record metrics on zone-proxy-egress", func() {
		uniEgressFilter := fmt.Sprintf(
			"cluster.kri_extsvc_%s_%s__external-service-meshproxy_8080.upstream_rq_200",
			meshName, multizone.UniZone1.ZoneName(),
		)
		kubeEgressFilter := fmt.Sprintf(
			"cluster.kri_extsvc_%s_%s_%s_external-service-kube_80.upstream_rq_200",
			meshName, multizone.KubeZone1.ZoneName(), Config.KumaNamespace,
		)

		// Kubernetes client -> external service
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client",
				"http://external-service-kube.extsvc.mesh.local",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kube-external-service"))
		}, "30s", "1s").Should(Succeed())

		// Verify traffic was routed through KubeZone1's zone-proxy-egress
		Eventually(func(g Gomega) {
			stat, err := multizone.KubeZone1.GetEnvoyAdminTunnel(egressApp, namespace).GetStats(kubeEgressFilter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// Universal client -> external service
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.UniZone1, "demo-client",
				"http://external-service-meshproxy.extsvc.mesh.local:8080",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("external-service-meshproxy"))
		}, "30s", "1s").Should(Succeed())

		// Verify traffic was routed through UniZone1's zone-proxy-egress
		Eventually(func(g Gomega) {
			stat, err := multizone.UniZone1.GetAppEnvoyTunnel(egressApp).GetStats(uniEgressFilter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())

		// Verify KubeZone1 egress filterchain SNI
		kubeEgressSNI := fmt.Sprintf("sni.extsvc.%s.%s.%s.external-service-kube.80", meshName, multizone.KubeZone1.ZoneName(), Config.KumaNamespace)
		Eventually(func(g Gomega) {
			xds, err := multizone.KubeZone1.GetEnvoyAdminTunnel(egressApp, namespace).GetConfigDump()
			g.Expect(err).ToNot(HaveOccurred())
			snis, err := listenerFilterchainSNIs(xds, "zoneegress")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(snis).To(ContainElement(kubeEgressSNI))
		}, "30s", "1s").Should(Succeed())

		// Verify UniZone1 egress filterchain SNI
		uniEgressSNI := fmt.Sprintf("sni.extsvc.%s.%s.external-service-meshproxy.8080", meshName, multizone.UniZone1.ZoneName())
		Eventually(func(g Gomega) {
			xds, err := multizone.UniZone1.GetAppEnvoyTunnel(egressApp).GetConfigDump()
			g.Expect(err).ToNot(HaveOccurred())
			snis, err := listenerFilterchainSNIs(xds, "zoneegress")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(snis).To(ContainElement(uniEgressSNI))
		}, "30s", "1s").Should(Succeed())

		// Verify UniZone1 demo-client cluster SNI for external-service-meshproxy
		Eventually(func(g Gomega) {
			xds, err := multizone.UniZone1.GetAppEnvoyTunnel("demo-client").GetConfigDump()
			g.Expect(err).ToNot(HaveOccurred())
			sni, err := clusterSNI(xds, fmt.Sprintf("kri_extsvc_%s_%s__external-service-meshproxy_8080", meshName, multizone.UniZone1.ZoneName()))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(sni).To(Equal(uniEgressSNI))
		}, "30s", "1s").Should(Succeed())
	})
}

func listenerFilterchainSNIs(cfg *config_dump.EnvoyConfig, listenerNameFragment string) ([]string, error) {
	for _, dl := range cfg.Listeners.DynamicListeners {
		if !strings.Contains(dl.Name, listenerNameFragment) {
			continue
		}
		if dl.ActiveState == nil {
			continue
		}
		var listener envoy_listener_v3.Listener
		if err := util_proto.UnmarshalAnyTo(dl.ActiveState.Listener, &listener); err != nil {
			return nil, err
		}
		var snis []string
		for _, fc := range listener.FilterChains {
			if fc.FilterChainMatch != nil {
				snis = append(snis, fc.FilterChainMatch.ServerNames...)
			}
		}
		return snis, nil
	}
	return nil, fmt.Errorf("no listener containing %q found in config dump", listenerNameFragment)
}

func clusterSNI(cfg *config_dump.EnvoyConfig, nameFragment string) (string, error) {
	for _, dc := range cfg.Cluster.DynamicActiveClusters {
		var cluster envoy_cluster_v3.Cluster
		if err := util_proto.UnmarshalAnyTo(dc.Cluster, &cluster); err != nil {
			return "", err
		}
		if !strings.Contains(cluster.Name, nameFragment) {
			continue
		}
		if cluster.TransportSocket == nil {
			continue
		}
		var tlsCtx envoy_tls_v3.UpstreamTlsContext
		if err := util_proto.UnmarshalAnyTo(cluster.GetTransportSocket().GetTypedConfig(), &tlsCtx); err != nil {
			return "", err
		}
		return tlsCtx.Sni, nil
	}
	return "", fmt.Errorf("no cluster containing %q found in config dump", nameFragment)
}
