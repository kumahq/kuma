package kic

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	client "github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/kic"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
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

	getKICIP := func() string {
		var ip string
		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				env.Cluster.GetTesting(),
				env.Cluster.GetKubectlOptions(namespace),
				"get", "service", "gateway", "-ojsonpath={.spec.clusterIP}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(BeEmpty())
			ip = out
		}, "60s", "1s").Should(Succeed(), "could not get the clusterIP of the Service")
		return ip
	}

	var kicIP string

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceOutsideMesh)).
			Install(DemoClientK8s(mesh, namespaceOutsideMesh)). // this will not be in the mesh
			Install(kic.KongIngressController(
				kic.WithNamespace(namespace),
				kic.WithMesh(mesh),
			)).
			Install(kic.KongIngressService(kic.WithNamespace(namespace))).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("test-server"),
			)).
			Setup(env.Cluster)).To(Succeed())

		kicIP = getKICIP()
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.TriggerDeleteNamespace(namespaceOutsideMesh)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should route to service using Kube DNS", func() {
		ingress := `
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: kic
  name: kube-dns-ingress
  annotations:
    kubernetes.io/ingress.class: kong
spec:
  rules:
  - http:
      paths:
      - path: /test-server
        backend:
          serviceName: test-server
          servicePort: 80
`
		Expect(env.Cluster.Install(YamlK8s(ingress))).To(Succeed())

		Eventually(func(g Gomega) {
			_, err := client.CollectResponse(
				env.Cluster, "demo-client", fmt.Sprintf("http://%s/test-server", kicIP),
				client.FromKubernetesPod(namespaceOutsideMesh, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})

	It("should route to service using Kuma DNS", func() {
		const ingressMeshDNS = `
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
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: kic
  name: k8s-ingress-dot-mesh
  annotations:
    kubernetes.io/ingress.class: kong
spec:
  rules:
  - http:
      paths:
      - path: /dot-mesh
        backend:
          serviceName: test-server-externalname
          servicePort: 80
`

		Expect(env.Cluster.Install(YamlK8s(ingressMeshDNS))).To(Succeed())

		Eventually(func(g Gomega) {
			_, err := client.CollectResponse(
				env.Cluster, "demo-client", fmt.Sprintf("http://%s/dot-mesh", kicIP),
				client.FromKubernetesPod(namespaceOutsideMesh, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
