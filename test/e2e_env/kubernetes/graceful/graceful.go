package graceful

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/util/channels"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
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
	gatewayInstnace := func(replicas int) string {
		return fmt.Sprintf(`
---
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: edge-gateway
  namespace: graceful
  labels:
    kuma.io/mesh: graceful
spec:
  replicas: %d
  serviceType: LoadBalancer
`, replicas)
	}

	gateway := `
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
      kuma.io/service: edge-gateway_graceful_svc
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
      kuma.io/service: edge-gateway_graceful_svc
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

	var gwIP string

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(YamlK8s(gateway)).
			Install(YamlK8s(gatewayInstnace(1))).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName(name),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).To(Succeed())

		Eventually(func(g Gomega) {
			address, err := kubernetes.Cluster.GetLBIngressIP("edge-gateway", namespace)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(address).ToNot(BeEmpty())
			gwIP = address
		}, "60s", "1s").Should(Succeed(), "could not get a LoadBalancer IP of the Gateway")

		// remove retries to avoid covering failed request
		Expect(DeleteMeshResources(kubernetes.Cluster, mesh, core_mesh.RetryResourceTypeDescriptor)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, namespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	requestThroughGateway := func() error {
		resp, err := httpClient.Get("http://" + gwIP + ":8080")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return errors.Errorf("status code: %d body: %s", resp.StatusCode, string(bytes))
		}
		_, err = io.Copy(io.Discard, resp.Body)
		return err
	}

	type testCase struct {
		deploymentName string
		scaleFn        func(int) error
	}

	DescribeTable("should not drop a request when scaling up and down",
		func(given testCase) {
			// given constant traffic between client and server
			Eventually(requestThroughGateway, "30s", "1s").Should(Succeed())
			var failedErr error
			closeCh := make(chan struct{})
			defer close(closeCh)
			go func() {
				for {
					if channels.IsClosed(closeCh) {
						return
					}
					if err := requestThroughGateway(); err != nil {
						failedErr = err
						return
					}
					// add a slight delay to not overwhelm completely the host running this test and leave more resources to other tests running in parallel.
					time.Sleep(50 * time.Millisecond)
				}
			}()

			// when
			Expect(given.scaleFn(2)).To(Succeed())

			// then
			Eventually(func(g Gomega) {
				g.Expect(WaitNumPods(namespace, 2, given.deploymentName)(kubernetes.Cluster)).To(Succeed())
				g.Expect(WaitPodsAvailable(namespace, given.deploymentName)(kubernetes.Cluster)).To(Succeed())
			}, "30s", "1s").Should(Succeed())
			Expect(failedErr).ToNot(HaveOccurred())

			// when
			Expect(given.scaleFn(1)).To(Succeed())

			// then
			Eventually(func(g Gomega) {
				g.Expect(WaitNumPods(namespace, 1, given.deploymentName)(kubernetes.Cluster)).To(Succeed())
			}, "60s", "1s").Should(Succeed())

			Expect(failedErr).ToNot(HaveOccurred())
		},
		Entry("a service", testCase{
			deploymentName: name,
			scaleFn: func(replicas int) error {
				return k8s.RunKubectlE(
					kubernetes.Cluster.GetTesting(),
					kubernetes.Cluster.GetKubectlOptions(namespace),
					"scale", "deployment", name, "--replicas", strconv.Itoa(replicas),
				)
			},
		}),
		Entry("a gateway", testCase{
			deploymentName: "edge-gateway",
			scaleFn: func(replicas int) error {
				return kubernetes.Cluster.Install(YamlK8s(gatewayInstnace(replicas)))
			},
		}),
	)
}
