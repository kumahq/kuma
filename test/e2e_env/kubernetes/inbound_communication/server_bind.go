package inbound_communication

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ServerBind() {
	const namespace = "server-bind"
	const mesh = "server-bind"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("test-server-localhost"),
				testserver.WithEchoArgs("echo", "--instance", "bound-localhost", "--ip", "127.0.0.1"),
				testserver.WithoutProbe(),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("test-server-wildcard"),
				testserver.WithEchoArgs("echo", "--instance", "bound-wildcard", "--ip", "0.0.0.0"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("test-server-pod"),
				testserver.WithEchoArgs("echo", "--instance", "bound-pod", "--ip", "$(POD_IP)"),
			)).
			Install(DemoClientK8s(mesh, namespace)).
			Setup(env.Cluster)
		Expect(err).To(Succeed())
	})
	E2EAfterAll(func() {
		// Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		// Expect(env.Cluster.DeleteMesh(mesh))
	})

	FIt("should communicate with the applications", func() {
		// given
		podName, err := PodNameOfApp(env.Cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
		podIp, err := PodIPOfApp(env.Cluster, "test-server-pod", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		stdout, _, err := env.Cluster.Exec(namespace, podName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "http://"+net.JoinHostPort(podIp, "80")+"/")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("bound-pod"))
	})
}
