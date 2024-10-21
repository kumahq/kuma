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

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should setup connectivity before app starts", func() {
		// the pod simulates what many app does which is connecting to external destination (like a database) immediately
		// restartPolicy is Never so if we fail we won't restart and the test fails.
		err := NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: wait-for-envoy
  namespace: %s
  labels:
    app: wait-for-envoy
    kuma.io/mesh: %s
  annotations:
    kuma.io/wait-for-dataplane-ready: "true"
spec:
  restartPolicy: Never
  containers:
  - name: alpine
    image: %s
    args:
    - /bin/bash
    - -c
    - --
    - 'curl --max-time 3 --fail test-server_wait-for-envoy_svc_80.mesh && test-server echo --port 80'
    readinessProbe:
      httpGet:
        path: /
        port: 80
      successThreshold: 1
    resources:
      limits:
        cpu: 50m
        memory: 64Mi`, namespace, meshName, framework.Config.GetUniversalImage()))).
			Install(framework.WaitNumPods(namespace, 1, "wait-for-envoy")).
			Install(framework.WaitPodsAvailable(namespace, "wait-for-envoy")).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})
}
