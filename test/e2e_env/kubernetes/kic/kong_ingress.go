package kic

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/kic"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func KICKubernetes() {
	// IPv6 currently not supported by Kong Ingress Controller
	// https://github.com/Kong/kubernetes-ingress-controller/issues/1017
	if Config.IPV6 {
		fmt.Println("Test not supported on IPv6")
		return
	}
	if Config.K8sType == KindK8sType {
		// KIC 2.0 when started with service type LoadBalancer requires external IP to be provisioned before it's healthy.
		// KIND cannot provision external IP, K3D can.
		fmt.Println("Test not supported on KIND")
		return
	}

	namespace := "kic"
	mesh := "kic"
	namespaceOutsideMesh := "kic-external"

	var kicIP string

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshKubernetes(mesh)).
			Install(MeshTrafficPermissionAllowAllKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceOutsideMesh)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(namespaceOutsideMesh)), // this will not be in the mesh
				kic.KongIngressController(
					kic.WithNamespace(namespace),
					kic.WithName("kic"),
					kic.WithMesh(mesh),
				),
				kic.KongIngressService(
					kic.WithNamespace(namespace),
					kic.WithName("kic"),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(mesh),
					testserver.WithName("test-server"),
				),
			)).
			Setup(kubernetes.Cluster)).To(Succeed())

		ip, err := kic.From(kubernetes.Cluster).IP(namespace)
		Expect(err).To(Succeed())

		kicIP = ip
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, namespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespaceOutsideMesh)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should route to service using Kube DNS", func() {
		ingress := `
---
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: kic
  annotations:
    konghq.com/gatewayclass-unmanaged: 'true'
spec:
  controllerName: konghq.com/kic-gateway-controller
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: kong
  namespace: kic
spec:
  gatewayClassName: kic
  listeners:
  - name: proxy
    port: 80
    protocol: HTTP
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: echo
  namespace: kic
  annotations:
    konghq.com/strip-path: 'true'
spec:
  parentRefs:
  - name: kong
    namespace: kic
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /test-server
    backendRefs:
    - name: test-server
      kind: Service
      port: 80
`
		Expect(kubernetes.Cluster.Install(YamlK8s(ingress))).To(Succeed())

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", fmt.Sprintf("http://%s/test-server", kicIP),
				client.FromKubernetesPod(namespaceOutsideMesh, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})

	It("should route to service using Kuma DNS", func() {
		const ingressMeshDNS = `
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: echo
  namespace: kic
  annotations:
    konghq.com/strip-path: 'true'
spec:
  parentRefs:
  - name: kong
    namespace: kic
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /test-server
    backendRefs:
    - name: test-server
      kind: Service
      port: 80
---
apiVersion: v1
kind: Service
metadata:
  name: test-server-externalname
  namespace: kic
spec:
  type: ExternalName
  externalName: test-server.kic.svc.80.mesh
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: k8s-ingress-dot-mesh
  namespace: kic
  annotations:
    konghq.com/strip-path: 'true'
spec:
  parentRefs:
  - name: kong
    namespace: kic
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /dot-mesh
    backendRefs:
    - name: test-server-externalname
      kind: Service
      port: 80
`

		Expect(kubernetes.Cluster.Install(YamlK8s(ingressMeshDNS))).To(Succeed())

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", fmt.Sprintf("http://%s/dot-mesh", kicIP),
				client.FromKubernetesPod(namespaceOutsideMesh, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
