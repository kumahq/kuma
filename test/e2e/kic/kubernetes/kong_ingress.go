package kubernetes

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/kic"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func requestTestServerThroughKong(port int) error {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := netClient.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Ingress returned status code %d", resp.StatusCode)
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

const ingress = `
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: kuma-test
  name: k8s-ingress
  annotations:
    kubernetes.io/ingress.class: kong
spec:
  rules:
  - http:
      paths:
      - path: /
        backend:
          serviceName: test-server
          servicePort: 80
`

const ingressMeshDNS = `
---
apiVersion: v1
kind: Service
metadata:
  name: test-server-externalname
  namespace: kuma-test
spec:
  type: ExternalName
  externalName: test-server.kuma-test.svc.80.mesh
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: kuma-test
  name: k8s-ingress-dot-mesh
  annotations:
    kubernetes.io/ingress.class: kong
spec:
  rules:
  - http:
      paths:
      - path: /
        backend:
          serviceName: test-server-externalname
          servicePort: 80
`

func KICKubernetes() {
	// IPv6 curently not supported by Kong Ingress Controller
	// https://github.com/Kong/kubernetes-ingress-controller/issues/1017
	if IsIPv6() {
		fmt.Println("Test not supported on IPv6")
		return
	}
	if !IsK3D() {
		// KIC 2.0 when started with service type LoadBalancer requires external IP to be provisioned before it's healthy.
		// KIND cannot provision external IP, K3D can.
		fmt.Println("Test not supported on KIND")
		return
	}

	var ingressNamespace string
	var altIngressNamespace = "kuma-yawetag"
	var kubernetes Cluster
	var kubernetesOps = KumaK8sDeployOpts
	E2EBeforeSuite(func() {
		k8sClusters, err := NewK8sClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		kubernetes = k8sClusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(config_core.Standalone, kubernetesOps...)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(testserver.Install()).
			Setup(kubernetes)
		Expect(err).ToNot(HaveOccurred())
		err = kubernetes.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterEach(func() {
		err := k8s.RunKubectlE(kubernetes.GetTesting(), kubernetes.GetKubectlOptions(), "delete", "ingress", "--all", "-n", "kuma-test")
		Expect(err).ToNot(HaveOccurred())
		Expect(kubernetes.DeleteNamespace(ingressNamespace)).To(Succeed())
	})
	E2EAfterSuite(func() {
		Expect(kubernetes.DeleteKuma(kubernetesOps...)).To(Succeed())
		Expect(kubernetes.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(kubernetes.DismissCluster()).To(Succeed())
	})
	It("should install kong ingress into default namespace", func() {
		ingressNamespace = DefaultGatewayNamespace
		// given kong ingress
		err := NewClusterSetup().
			Install(kic.KongIngressController()).
			Install(kic.KongIngressNodePort()).
			Install(YamlK8s(ingress)).
			Setup(kubernetes)
		Expect(err).ToNot(HaveOccurred())

		retry.DoWithRetry(kubernetes.GetTesting(), "connect to test server via KIC",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				return "", requestTestServerThroughKong(kic.NodePortHTTP())
			})

		Expect(err).ToNot(HaveOccurred())
	})
	It("should install kong ingress into non-default namespace", func() {
		ingressNamespace = altIngressNamespace
		// given kong ingress
		err := NewClusterSetup().
			Install(kic.KongIngressController(kic.WithNamespace(ingressNamespace))).
			Install(kic.KongIngressNodePort(kic.WithNamespace(ingressNamespace))).
			Install(YamlK8s(ingressMeshDNS)).
			Setup(kubernetes)
		Expect(err).ToNot(HaveOccurred())

		retry.DoWithRetry(kubernetes.GetTesting(), "connect to test server via KIC",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				return "", requestTestServerThroughKong(kic.NodePortHTTP())
			})

		Expect(err).ToNot(HaveOccurred())
	})
}
