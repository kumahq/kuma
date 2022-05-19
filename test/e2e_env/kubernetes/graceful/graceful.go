package graceful

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/util/channels"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func Graceful() {
	if Config.Arch == "arm64" {
		return // K3D loadbalancer required for this test seems to not work with K3D
	}
	if Config.IPV6 {
		return // K3D cannot handle loadbalancer for IPV6
	}

	const name = "graceful"
	const namespace = "graceful"
	const mesh = "graceful"

	// Set up a gateway to be able to send requests constantly.
	// The alternative was to exec to the container, but this introduces a latency of getting into container for every curl
	gateway := `
---
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: edge-gateway
  namespace: graceful
  annotations:
    kuma.io/mesh: graceful
spec:
  replicas: 1
  serviceType: LoadBalancer
  tags:
    kuma.io/service: edge-gateway
---
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: edge-gateway
  namespace: graceful
mesh: graceful
spec:
  selectors:
  - match:
      kuma.io/service: edge-gateway
  conf:
    listeners:
    - port: 8080
      protocol: HTTP
---
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: edge-gateway
  namespace: graceful
mesh: graceful
spec:
  selectors:
  - match:
      kuma.io/service: edge-gateway
  conf:
    http:
      rules:
      - matches:
        - path:
            match: PREFIX
            value: /
        backends:
        - destination:
            kuma.io/service: graceful_graceful_svc_80
`

	var gatewayIP string

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}

	BeforeAll(func() {
		E2EDeferCleanup(func() {
			Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(mesh))
		})

		err := NewClusterSetup().
			Install(MeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(YamlK8s(gateway)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName(name),
			)).
			Setup(env.Cluster)
		Expect(err).To(Succeed())

		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				env.Cluster.GetTesting(),
				env.Cluster.GetKubectlOptions(namespace),
				"get", "service", "edge-gateway", "-ojsonpath={.status.loadBalancer.ingress[0].ip}",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(BeEmpty())
			gatewayIP = out
		}, "60s", "1s").Should(Succeed(), "could not get a LoadBalancer IP of the Gateway")
	})

	requestThroughGateway := func() error {
		resp, err := httpClient.Get("http://" + gatewayIP + ":8080")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.Errorf("status code: %d", resp.StatusCode)
		}
		_, err = io.Copy(ioutil.Discard, resp.Body)
		return err
	}

	It("should not drop a request when scaling up and down", func() {
		// given no retries
		err := k8s.RunKubectlE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(),
			"delete", "retry", "retry-all-graceful",
		)
		Expect(err).ToNot(HaveOccurred())

		// and constant traffic between client and server
		Eventually(requestThroughGateway, "30s", "1s").Should(Succeed())
		var failedErr error
		closeCh := make(chan struct{})
		defer close(closeCh)
		go func() {
			for {
				if err := requestThroughGateway(); err != nil {
					failedErr = err
					return
				}
				if channels.IsClosed(closeCh) {
					return
				}
			}
		}()

		// when
		err = k8s.RunKubectlE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(namespace),
			"scale", "deployment", name, "--replicas", "2",
		)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			g.Expect(WaitNumPods(namespace, 2, name)(env.Cluster)).To(Succeed())
			g.Expect(WaitPodsAvailable(namespace, name)(env.Cluster)).To(Succeed())
		}, "30s", "1s").Should(Succeed())
		Expect(failedErr).ToNot(HaveOccurred())

		// when
		err = k8s.RunKubectlE(
			env.Cluster.GetTesting(),
			env.Cluster.GetKubectlOptions(namespace),
			"scale", "deployment", name, "--replicas", "1",
		)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			g.Expect(WaitNumPods(namespace, 1, name)(env.Cluster)).To(Succeed())
		}, "60s", "1s").Should(Succeed())

		Expect(failedErr).ToNot(HaveOccurred())
	})
}
