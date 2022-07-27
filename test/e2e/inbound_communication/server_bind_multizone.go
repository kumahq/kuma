package inbound_communication

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func MeshMTLSOnAndZoneEgress(mesh string) InstallFunc {
	return YamlUniversal(fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  zoneEgress: true
`, mesh))
}

func ServerBind() {
	const mesh = "server-bind"
	const meshEgress = "server-bind-egress"
	const namespace = "server-bind"
	const namespaceEgress = "server-bind-egress"
	var global, zone1, zone4 Cluster
	// var zone4 *UniversalCluster

	BeforeAll(func() {
		k8sClusters, err := NewK8sClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters(
			[]string{Kuma4, Kuma5},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		global = universalClusters.GetCluster(Kuma5)

		Expect(NewClusterSetup().
			Install(Kuma(config_core.Global,
				WithEnv("KUMA_DEFAULTS_ENABLE_INBOUND_PASSTHROUGH", "false"),
			)).
			Install(MTLSMeshUniversal(mesh)).
			Install(MeshMTLSOnAndZoneEgress(meshEgress)).
			Setup(global)).To(Succeed())
		globalCP := global.GetKuma()

		// K8s Cluster 1
		zone1 = k8sClusters.GetCluster(Kuma1)
		Expect(NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithEgress(),
				WithIngress(),
				WithEgressEnvoyAdminTunnel(),
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithEnv("KUMA_DEFAULTS_ENABLE_INBOUND_PASSTHROUGH", "false"),
			)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(NamespaceWithSidecarInjection(namespaceEgress)).
			Install(DemoClientK8s(mesh, namespace)).
			Install(DemoClientK8s(meshEgress, namespaceEgress)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("k8s-test-server-localhost"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-localhost", "--ip", "127.0.0.1"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("k8s-test-server-wildcard"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-wildcard", "--ip", "0.0.0.0"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("k8s-test-server-pod"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-pod", "--ip", "$(POD_IP)"),
				testserver.WithoutProbe(),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespaceEgress),
				testserver.WithMesh(meshEgress),
				testserver.WithName("k8s-test-server-localhost-egress"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-localhost-egress", "--ip", "127.0.0.1"),
				testserver.WithoutProbe(),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespaceEgress),
				testserver.WithMesh(meshEgress),
				testserver.WithName("k8s-test-server-wildcard-egress"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-wildcard-egress", "--ip", "0.0.0.0"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespaceEgress),
				testserver.WithMesh(meshEgress),
				testserver.WithName("k8s-test-server-pod-egress"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-pod-egress", "--ip", "$(POD_IP)"),
				testserver.WithoutProbe(),
			)).
			Setup(zone1)).To(Succeed())

		// Universal Cluster 4
		zone4 = universalClusters.GetCluster(Kuma4).(*UniversalCluster)
		Expect(err).ToNot(HaveOccurred())

		Expect(NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithEnv("KUMA_DEFAULTS_ENABLE_INBOUND_PASSTHROUGH", "false"),
			)).
			Install(DemoClientUniversal(
				"uni-demo-client",
				mesh,
				WithTransparentProxy(true),
			)).
			Install(DemoClientUniversal(
				"uni-demo-client-egress",
				meshEgress,
				WithTransparentProxy(true),
			)).
			// TODO: bind to docker ip
			Install(TestServerUniversal("uni-test-server-localhost", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-localhost", "--ip", "127.0.0.1"}),
				ServiceProbe(),
				WithServiceName("uni-test-server-localhost"),
			)).
			Install(TestServerUniversal("uni-test-server-wildcard", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-wildcard", "--ip", "0.0.0.0"}),
				ServiceProbe(),
				WithServiceName("uni-test-server-wildcard"),
			)).
			Install(TestServerUniversal("uni-test-server-wildcard-no-tp", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-wildcard-no-tp", "--ip", "0.0.0.0"}),
				ServiceProbe(),
				WithTransparentProxy(false),
				WithServiceName("uni-test-server-wildcard-no-tp"),
			)).
			Install(TestServerUniversal("uni-test-server-localhost-egress", meshEgress,
				WithArgs([]string{"echo", "--instance", "uni-bound-localhost-egress", "--ip", "127.0.0.1"}),
				ServiceProbe(),
				WithServiceName("uni-test-server-localhost-egress"),
			)).
			Install(TestServerUniversal("uni-test-server-wildcard-egress", meshEgress,
				WithArgs([]string{"echo", "--instance", "uni-bound-wildcard-egress", "--ip", "0.0.0.0"}),
				ServiceProbe(),
				WithServiceName("uni-test-server-wildcard-egress"),
			)).
			Install(TestServerUniversal("uni-test-server-wildcard-no-tp-egress", meshEgress,
				WithArgs([]string{"echo", "--instance", "uni-bound-wildcard-no-tp-egress", "--ip", "0.0.0.0"}),
				ServiceProbe(),
				WithTransparentProxy(false),
				WithServiceName("uni-test-server-wildcard-no-tp-egress"),
			)).
			Install(EgressUniversal(globalCP.GenerateZoneEgressToken)).
			Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
			Setup(zone4),
		).To(Succeed())
	})
	// E2EAfterAll(func() {
	// 	Expect(zone1.TriggerDeleteNamespace(namespace)).To(Succeed())
	// 	Expect(zone1.DeleteMesh(mesh))
	// 	Expect(zone4.DeleteMeshApps(mesh)).To(Succeed())
	// 	Expect(env.Global.DeleteMeshApps(mesh)).To(Succeed())
	// })

	It("should check communication k8s to k8s", func() {
		// given
		podName, err := PodNameOfApp(zone1, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		response, err := client.CollectResponse(
			zone1, "demo-client", "k8s-test-server-wildcard.server-bind.svc.80.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-wildcard"))

		// when
		response, err = client.CollectResponse(
			zone1, "demo-client", "k8s-test-server-localhost.server-bind.svc.80.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-localhost"))

		// when
		_, _, err = zone1.Exec(namespace, podName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "k8s-test-server-pod.server-bind.svc.80.mesh")

		// then
		Expect(err).To(HaveOccurred())
	})

	It("should check communication universal to universal", func() {
		// when
		response, err := client.CollectResponse(
			zone4, "uni-demo-client", "uni-test-server-wildcard.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard"))

		// when
		response, err = client.CollectResponse(
			zone4, "uni-demo-client", "uni-test-server-wildcard-no-tp.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard-no-tp"))

		// when
		response, err = client.CollectResponse(
			zone4, "uni-demo-client", "uni-test-server-localhost.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-localhost"))
	})

	It("should check communication k8s to universal", func() {
		// when
		response, err := client.CollectResponse(
			zone1, "demo-client", "uni-test-server-wildcard.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard"))

		// when
		response, err = client.CollectResponse(
			zone1, "demo-client", "uni-test-server-localhost.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-localhost"))

		// when
		response, err = client.CollectResponse(
			zone1, "demo-client", "uni-test-server-wildcard-no-tp.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard-no-tp"))
	})

	It("should check communication universal to k8s", func() {
		// when
		response, err := client.CollectResponse(
			zone4, "uni-demo-client", "k8s-test-server-wildcard.server-bind.svc.80.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-wildcard"))

		// when
		response, err = client.CollectResponse(
			zone4, "uni-demo-client", "k8s-test-server-localhost.server-bind.svc.80.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-localhost"))

		// when
		_, _, err = zone4.Exec("", "", "uni-demo-client",
			"curl", "-v", "-m", "3", "--fail", "k8s-test-server-pod.server-bind.svc.80.mesh")

		// then
		Expect(err).To(HaveOccurred())
	})

	It("should check communication universal to k8s with egress", func() {
		// when
		response, err := client.CollectResponse(
			zone4, "uni-demo-client-egress", "k8s-test-server-wildcard-egress.server-bind-egress.svc.80.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-wildcard-egress"))

		// when
		response, err = client.CollectResponse(
			zone4, "uni-demo-client-egress", "k8s-test-server-localhost-egress.server-bind-egress.svc.80.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-localhost-egress"))

		// when
		_, _, err = zone4.Exec("", "", "uni-demo-client-egress",
			"curl", "-v", "-m", "3", "--fail", "k8s-test-server-pod-egress.server-bind-egress.svc.80.mesh")

		// then
		Expect(err).To(HaveOccurred())
	})

	It("should check communication k8s to universal with egress", func() {
		// when
		response, err := client.CollectResponse(
			zone1, "demo-client", "uni-test-server-wildcard-egress.mesh",
			client.FromKubernetesPod(namespaceEgress, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard-egress"))

		// when
		response, err = client.CollectResponse(
			zone1, "demo-client", "uni-test-server-wildcard-no-tp-egress.mesh",
			client.FromKubernetesPod(namespaceEgress, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard-no-tp-egress"))

		// when
		response, err = client.CollectResponse(
			zone1, "demo-client", "uni-test-server-localhost-egress.mesh",
			client.FromKubernetesPod(namespaceEgress, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-localhost-egress"))
	})
}
