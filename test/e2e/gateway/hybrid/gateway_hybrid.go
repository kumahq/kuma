package hybrid

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/e2e/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func GatewayHybrid() {
	const serviceName = "test-server_kuma-test_svc_80"

	var global, k8sZone, zone1, zone2 Cluster

	E2EBeforeSuite(func() {
		global = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
		err := NewClusterSetup().
			Install(Kuma(config_core.Global,
				WithEnv("KUMA_EXPERIMENTAL_GATEWAY", "true"),
				gateway.OptEnableMeshMTLS),
			).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		Expect(global.VerifyKuma()).To(Succeed())

		k8sZone = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		err = NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithIngress(),
				WithCtlOpts(map[string]string{"--experimental-gateway": "true"}),
				WithGlobalAddress(global.GetKuma().GetKDSServerAddress()),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(testserver.Install(
				testserver.WithArgs("echo", "--instance", Kuma1),
			)).
			Setup(k8sZone)
		Expect(err).ToNot(HaveOccurred())

		zone1 = NewUniversalCluster(NewTestingT(), Kuma2, Silent)
		err = NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithGlobalAddress(global.GetKuma().GetKDSServerAddress()),
				WithEnv("KUMA_EXPERIMENTAL_GATEWAY", "true"),
			)).
			Install(gateway.EchoServerApp("echo-server", serviceName, Kuma2)).
			Install(gateway.GatewayProxyUniversal("gateway-proxy")).
			Install(gateway.GatewayClientAppUniversal("gateway-client")).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())

		zoneIngressToken, err := global.GetKuma().GenerateZoneIngressToken(Kuma3)
		Expect(err).ToNot(HaveOccurred())

		zone2 = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
		err = NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithGlobalAddress(global.GetKuma().GetKDSServerAddress()),
				WithEnv("KUMA_EXPERIMENTAL_GATEWAY", "true"),
			)).
			Install(gateway.EchoServerApp("echo-server", serviceName, Kuma3)).
			Install(IngressUniversal(zoneIngressToken)).
			Setup(zone2)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterSuite(func() {
		Expect(k8sZone.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(k8sZone.DeleteKuma()).To(Succeed())
		Expect(k8sZone.DismissCluster()).To(Succeed())

		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())

		Expect(zone2.DeleteKuma()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())

		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	type testCase struct {
		path              string
		expectedInstances []string
	}

	Context("proxying through instances across all zones", func() {
		DescribeTable("gateway proxies the traffic to echo service",
			func(given testCase) {
				err := NewClusterSetup().
					Install(YamlUniversal(`
type: Gateway
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  listeners:
  - port: 8080
    protocol: HTTP
    hostname: example.kuma.io
    tags:
      hostname: example.kuma.io
`,
					)).
					Install(YamlUniversal(`
type: GatewayRoute
mesh: default
name: edge-gateway
selectors:
- match:
    kuma.io/service: edge-gateway
conf:
  http:
    rules:
    - matches:
      - path:
          match: PREFIX
          value: /kuma-1
      backends:
      - destination:
          kuma.io/service: test-server_kuma-test_svc_80
          kuma.io/zone: kuma-1-zone
    - matches:
      - path:
          match: PREFIX
          value: /kuma-2
      backends:
      - destination:
          kuma.io/service: test-server_kuma-test_svc_80
          kuma.io/zone: kuma-2
    - matches:
      - path:
          match: PREFIX
          value: /kuma-3
      backends:
      - destination:
          kuma.io/service: test-server_kuma-test_svc_80
          kuma.io/zone: kuma-3
    - matches:
      - path:
          match: PREFIX
          value: /all
      backends:
      - destination:
          kuma.io/service: test-server_kuma-test_svc_80
`)).
					Setup(global)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func(g Gomega) {
					target := fmt.Sprintf("http://%s%s",
						net.JoinHostPort(zone1.(*UniversalCluster).GetApp("gateway-proxy").GetIP(), "8080"),
						given.path,
					)
					responses, err := client.CollectResponsesByInstance(zone1, "gateway-client", target, client.WithHeader("Host", "example.kuma.io"))
					g.Expect(err).To(Succeed())

					g.Expect(responses).To(HaveLen(len(given.expectedInstances)))
					for _, expectedInstance := range given.expectedInstances {
						g.Expect(responses).To(HaveKey(expectedInstance))
					}
				}, "30s", "1s").Should(Succeed())
			},
			Entry("should proxy between all instances", testCase{
				path:              "/all",
				expectedInstances: []string{Kuma1, Kuma2, Kuma3},
			}),
			Entry("should proxy to the zone-1", testCase{
				path:              "/kuma-1",
				expectedInstances: []string{Kuma1},
			}),
			Entry("should proxy to the zone-2", testCase{
				path:              "/kuma-2",
				expectedInstances: []string{Kuma2},
			}),
			Entry("should proxy to the zone-3", testCase{
				path:              "/kuma-3",
				expectedInstances: []string{Kuma3},
			}),
		)
	})
}
