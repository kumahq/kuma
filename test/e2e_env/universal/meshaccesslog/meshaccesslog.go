package meshaccesslog

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	"github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
)

func TestPlugin() {
	meshName := "meshaccesslog"
	externalServiceName := "meshaccesslog-" + externalservice.TcpSink
	externalServiceDeployment := "externalservice-" + externalServiceName
	var externalServiceDockerName string

	GatewayAddressPort := func(appName string, port int) string {
		ip := env.Cluster.GetApp(appName).GetIP()
		return net.JoinHostPort(ip, strconv.Itoa(port))
	}

	BeforeAll(func() {
		externalServiceDockerName = fmt.Sprintf("%s_%s-%s", env.Cluster.Name(), meshName, "test-server")
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TestServerUniversal(
				"test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}), WithDockerContainerName(externalServiceDockerName)),
			).
			Install(GatewayProxyUniversal(meshName, "edge-gateway")).
			Install(YamlUniversal(gateway.MkGateway("edge-gateway", meshName, false, "example.kuma.io", "test-server", 8080))).
			Install(gateway.GatewayClientAppUniversal("gateway-client")).
			Setup(env.Cluster)).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteApp("gateway-client")).To(Succeed())
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	// Always have new MeshAccessLog resources and log sink
	BeforeEach(func() {
		Expect(NewClusterSetup().
			Install(externalservice.Install(externalServiceName, externalservice.UniversalTCPSink)).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(env.Cluster),
		).To(Succeed())
	})
	E2EAfterEach(func() {
		items, err := env.Cluster.GetKumactlOptions().KumactlList("meshaccesslogs", meshName)
		Expect(err).ToNot(HaveOccurred())

		for _, item := range items {
			err := env.Cluster.GetKumactlOptions().KumactlDelete("meshaccesslog", item, meshName)
			Expect(err).ToNot(HaveOccurred())
		}

		Expect(env.Cluster.DeleteApp(AppModeDemoClient)).To(Succeed())
		Expect(env.Cluster.DeleteDeployment(externalServiceDeployment)).To(Succeed())
	})

	trafficLogFormat := "%START_TIME(%s)%,%KUMA_SOURCE_SERVICE%,%KUMA_DESTINATION_SERVICE%"
	expectTrafficLogged := func(makeRequest func(g Gomega)) (string, string) {
		var src, dst string
		sinkDeployment := env.Cluster.Deployment(externalServiceDeployment).(*externalservice.UniversalDeployment)

		Eventually(func(g Gomega) {
			makeRequest(g)

			stdout, _, err := sinkDeployment.Exec("", "", "head", "-1", "/nc.out")
			g.Expect(err).ToNot(HaveOccurred())
			parts := strings.Split(stdout, ",")
			g.Expect(parts).To(HaveLen(3))

			startTimeInt, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			Expect(err).ToNot(HaveOccurred())
			startTime := time.Unix(int64(startTimeInt), 0)
			Expect(startTime).To(BeTemporally("~", time.Now(), time.Hour))

			src, dst = parts[1], parts[2]
		}, "30s", "1s").Should(Succeed())

		return strings.TrimSpace(src), strings.TrimSpace(dst)
	}

	It("should log outgoing traffic", func() {
		yaml := fmt.Sprintf(`
type: MeshAccessLog
name: client-outgoing
mesh: meshaccesslog
spec:
 targetRef:
   kind: MeshService
   name: demo-client
 to:
   - targetRef:
       kind: Mesh
     default:
       backends:
       - tcp:
           format:
             plain: '%s'
           address: "%s_%s:9999"
`, trafficLogFormat, env.Cluster.Name(), externalServiceDeployment)
		Expect(YamlUniversal(yaml)(env.Cluster)).To(Succeed())

		makeRequest := func(g Gomega) {
			_, _, err := env.Cluster.Exec("", "", AppModeDemoClient,
				"curl", "-v", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
		}
		src, dst := expectTrafficLogged(makeRequest)

		Expect(src).To(Equal(AppModeDemoClient))
		Expect(dst).To(Equal("test-server"))
	})

	It("should log outgoing traffic with JSON formatting", func() {
		yaml := fmt.Sprintf(`
type: MeshAccessLog
name: client-outgoing
mesh: meshaccesslog
spec:
 targetRef:
   kind: MeshService
   name: demo-client
 to:
   - targetRef:
       kind: Mesh
     default:
       backends:
       - tcp:
           format:
             json:
             - key: Source
               value: '%%KUMA_SOURCE_SERVICE%%'
             - key: Destination
               value: '%%KUMA_DESTINATION_SERVICE%%'
             - key: Start
               value: '%%START_TIME(%%s)%%'
           address: "%s_%s:9999"
`, env.Cluster.Name(), externalServiceDeployment)
		Expect(YamlUniversal(yaml)(env.Cluster)).To(Succeed())

		var src, dst string
		sinkDeployment := env.Cluster.Deployment(externalServiceDeployment).(*externalservice.UniversalDeployment)
		Eventually(func(g Gomega) {
			_, _, err := env.Cluster.Exec("", "", AppModeDemoClient,
				"curl", "-v", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())

			stdout, _, err := sinkDeployment.Exec("", "", "head", "-1", "/nc.out")
			g.Expect(err).ToNot(HaveOccurred())

			type log struct {
				Source      string
				Destination string
				Start       string
			}
			var line log
			g.Expect(json.Unmarshal([]byte(stdout), &line)).To(Succeed())

			src = line.Source
			dst = line.Destination
		}, "30s", "1s").Should(Succeed())

		Expect(src).To(Equal(AppModeDemoClient))
		Expect(dst).To(Equal("test-server"))
	})

	// This is flaky if we don't redeploy demo-client in BeforeEach/E2EAfterEach
	// This may have something to do with access-log-streamer
	It("should log outgoing passthrough traffic", func() {
		yaml := fmt.Sprintf(`
type: MeshAccessLog
name: client-outgoing
mesh: meshaccesslog
spec:
 targetRef:
   kind: MeshService
   name: demo-client
 to:
   - targetRef:
       kind: Mesh
     default:
       backends:
       - tcp:
           format:
             plain: '%s'
           address: "%s_%s:9999"
`, trafficLogFormat, env.Cluster.Name(), externalServiceDeployment)
		Expect(YamlUniversal(yaml)(env.Cluster)).To(Succeed())

		// 52 is empty response but the TCP connection succeeded
		makeRequest := func(g Gomega) {
			_, _, err := env.Cluster.Exec("", "", AppModeDemoClient,
				"curl", "-v", "--fail", externalServiceDockerName)
			g.Expect(err).To(ContainSubstring("exit status 52"))
		}
		src, dst := expectTrafficLogged(makeRequest)

		Expect(src).To(Equal(AppModeDemoClient))
		Expect(dst).To(Equal("external"))
	})

	It("supports logging traffic to an ExternalService using MeshService (without ZoneIngress)", func() {
		externalService := fmt.Sprintf(`
type: ExternalService
name: external-service
mesh: meshaccesslog
tags:
  kuma.io/service: ext-service
  kuma.io/protocol: tcp
networking:
  address: "%s:80"
`, externalServiceDockerName)
		accesslog := fmt.Sprintf(`
type: MeshAccessLog
name: client-outgoing
mesh: meshaccesslog
spec:
 targetRef:
   kind: MeshService
   name: demo-client
 to:
   - targetRef:
       kind: MeshService
       name: ext-service
     default:
       backends:
       - tcp:
           format:
             plain: '%s'
           address: "%s_%s:9999"
`, trafficLogFormat, env.Cluster.Name(), externalServiceDeployment)
		Expect(YamlUniversal(externalService)(env.Cluster)).To(Succeed())
		Expect(YamlUniversal(accesslog)(env.Cluster)).To(Succeed())

		// 52 is empty response but the TCP connection succeeded
		makeRequest := func(g Gomega) {
			_, _, err := env.Cluster.Exec("", "", AppModeDemoClient,
				"curl", "-v", "--fail", "ext-service.mesh")
			g.Expect(err).To(ContainSubstring("exit status 52"))
		}
		src, dst := expectTrafficLogged(makeRequest)

		Expect(src).To(Equal(AppModeDemoClient))
		Expect(dst).To(Equal("ext-service"))
	})

	It("should log incoming traffic", func() {
		yaml := fmt.Sprintf(`
type: MeshAccessLog
name: server-outgoing
mesh: meshaccesslog
spec:
 targetRef:
   kind: MeshService
   name: test-server
 from:
   - targetRef:
       kind: Mesh
     default:
       backends:
       - tcp:
           format:
             plain: '%s'
           address: "%s_%s:9999"
`, trafficLogFormat, env.Cluster.Name(), externalServiceDeployment)

		Expect(YamlUniversal(yaml)(env.Cluster)).To(Succeed())

		makeRequest := func(g Gomega) {
			_, _, err := env.Cluster.Exec("", "", AppModeDemoClient,
				"curl", "-v", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
		}
		src, dst := expectTrafficLogged(makeRequest)

		Expect(src).To(Equal("unknown"))
		Expect(dst).To(Equal("test-server"))
	})

	It("should log traffic from MeshGateway", func() {
		yaml := fmt.Sprintf(`
type: MeshAccessLog
name: gateway-outgoing
mesh: meshaccesslog
spec:
 targetRef:
   kind: MeshService
   name: edge-gateway
 to:
   - targetRef:
       kind: Mesh
     default:
       backends:
       - tcp:
           format:
             plain: '%s'
           address: "%s_%s:9999"
`, trafficLogFormat, env.Cluster.Name(), externalServiceDeployment)
		Expect(YamlUniversal(yaml)(env.Cluster)).To(Succeed())

		makeRequest := func(g Gomega) {
			gateway.ProxySimpleRequests(env.Cluster, "echo-v1",
				GatewayAddressPort("edge-gateway", 8080), "example.kuma.io")
		}
		src, dst := expectTrafficLogged(makeRequest)

		Expect(src).To(Equal("edge-gateway"))
		Expect(dst).To(Equal("*"))
	})
}
