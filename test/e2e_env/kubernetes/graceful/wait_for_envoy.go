package graceful

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func WaitForEnvoyReady() {
	namespace := "wait-for-envoy"
	meshName := "wait-for-envoy"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(MeshKubernetes(meshName)).
			Install(testserver.Install(
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	// different name to avoid snapshot cache
	pod := func(id int, waitForDataplane string) InstallFunc {
		return YamlK8s(fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: wait-for-envoy-%d
  namespace: %s
  annotations:
    kuma.io/mesh: %s
    kuma.io/wait-for-dataplane-ready: "%s"
spec:
  restartPolicy: Never
  containers:
  - name: alpine
    image: %s
    args:
    - /bin/bash
    - -c
    - --
    - 'while true; do curl --fail test-server_wait-for-envoy_svc_80.mesh && echo "succeeded" || { echo "failed" ; exit 1; }; done'
    resources:
      limits:
        cpu: 50m
        memory: 64Mi`, id, namespace, meshName, waitForDataplane, framework.Config.GetUniversalImage()))
	}

	It("remove Dataplane of evicted Pod", func() {
		// TODO: flaky because config is delivered too fast, how to slow it down?
		// when not waiting for envoy to be ready
		Expect(kubernetes.Cluster.Install(pod(1, "false"))).ToNot(Succeed())

		// when waiting for envoy to be ready
		Expect(kubernetes.Cluster.Install(pod(2, "true"))).To(Succeed())
	})
}
