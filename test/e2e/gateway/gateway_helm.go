package gateway

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func GatewayHELM() {
	var cluster Cluster

	// We deploy GatewayInstance with Service of type NodePort and port 30080
	// This will expose it on Kubernetes node, but the node is hidden in docker, so it has to be exposed further
	// to the VM on which the test is run. In k3d.mk we explicitly expose a range of ports including this one.
	const GatewayNodePort = "30080"

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		// todo(jakubdyszkiewicz) when we release Kuma with experimental Gateway, change HELM version to released one that does not contain Gateway CRDs
		Expect(NewClusterSetup().
			Install(Kuma(core.Standalone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmOpt("experimental.meshGateway", "true"),
				WithHelmReleaseName(fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueId()))),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(testserver.Install()).
			Setup(cluster)).To(Succeed())

		E2EDeferCleanup(func() {
			Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
			Expect(cluster.DeleteKuma()).To(Succeed())
			Expect(cluster.DismissCluster()).To(Succeed())
		})
	})

	It("should check if Gateway CRD is installed", func() {
		// given gateway with route to the deployed test server
		err := NewClusterSetup().
			Install(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: edge-gateway
spec:
  replicas: 1
  serviceType: NodePort
  tags:
    kuma.io/service: edge-gateway`)).
			Install(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: edge-gateway
mesh: default
spec:
  selectors:
  - match:
      kuma.io/service: edge-gateway
  conf:
    listeners:
    - port: 30080
      protocol: HTTP
      hostname: example.kuma.io
      tags:
        hostname: example.kuma.io`)).
			Install(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: edge-gateway
mesh: default
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
            kuma.io/service: test-server_kuma-test_svc_80`)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			// given
			address := "http://" + net.JoinHostPort("localhost", GatewayNodePort)
			if Config.IPV6 {
				// With IPV6, KIND forwards to the host but it works only on IPV6.
				// Just localhost:30800 will resolve to 127.0.0.1:30800 therefore we need explicit IPV6
				address = "http://" + net.JoinHostPort("::1", GatewayNodePort)
			}

			// when
			req, err := http.NewRequest("GET", address, nil)
			g.Expect(err).ToNot(HaveOccurred())
			req.Host = "example.kuma.io"
			resp, err := http.DefaultClient.Do(req)

			// then
			g.Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			g.Expect(resp.StatusCode).To(Equal(200))
		}, "30s", "1s").Should(Succeed())
	})
}
