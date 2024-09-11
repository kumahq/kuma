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

	"github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func TestPlugin() {
	meshName := "meshaccesslog"
	var externalServiceDockerName string
	var tcpSinkDockerName string

	GatewayAddressPort := func(appName string, port int) string {
		ip := universal.Cluster.GetApp(appName).GetIP()
		return net.JoinHostPort(ip, strconv.Itoa(port))
	}

	uniServiceYAML := fmt.Sprintf(`
type: MeshService
name: test-server
mesh: %s
labels:
  kuma.io/origin: zone
  kuma.io/env: universal
spec:
  selector:
    dataplaneTags:
      kuma.io/service: test-server
  ports:
  - port: 80
    targetPort: 80
    appProtocol: http
    name: main-port
`, meshName)

	BeforeAll(func() {
		externalServiceDockerName = fmt.Sprintf("%s_%s-%s", universal.Cluster.Name(), meshName, "test-server")
		tcpSinkDockerName = fmt.Sprintf("%s_%s_%s", universal.Cluster.Name(), meshName, AppModeTcpSink)
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TestServerUniversal(
				"test-server", meshName, WithArgs([]string{"echo", "--instance", "echo-v1"}), WithDockerContainerName(externalServiceDockerName)),
			).
			Install(YamlUniversal(uniServiceYAML)).
			Install(YamlUniversal(`
type: HostnameGenerator
name: uni-ms
spec:
  template: '{{ .DisplayName }}.universal.ms'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: zone
        kuma.io/env: universal`)).
			Install(GatewayProxyUniversal(meshName, "edge-gateway")).
			Install(YamlUniversal(gateway.MkGateway("edge-gateway", meshName, "edge-gateway", false, "example.kuma.io", "test-server", 8080))).
			Install(gateway.GatewayClientAppUniversal("gateway-client")).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteApp("gateway-client")).To(Succeed())
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	// Always have new MeshAccessLog resources and log sink
	BeforeEach(func() {
		Expect(NewClusterSetup().
			Install(TcpSinkUniversal(AppModeTcpSink, WithDockerContainerName(tcpSinkDockerName))).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster),
		).To(Succeed())
	})
	E2EAfterEach(func() {
		items, err := universal.Cluster.GetKumactlOptions().KumactlList("meshaccesslogs", meshName)
		Expect(err).ToNot(HaveOccurred())

		for _, item := range items {
			err := universal.Cluster.GetKumactlOptions().KumactlDelete("meshaccesslog", item, meshName)
			Expect(err).ToNot(HaveOccurred())
		}

		Expect(universal.Cluster.DeleteApp(AppModeDemoClient)).To(Succeed())
		Expect(universal.Cluster.DeleteApp(AppModeTcpSink)).To(Succeed())
	})

	trafficLogFormat := "%START_TIME(%s)%,%KUMA_SOURCE_SERVICE%,%KUMA_DESTINATION_SERVICE%"
	expectTrafficLogged := func(makeRequest func(g Gomega)) (string, string) {
		var src, dst string

		Eventually(func(g Gomega) {
			makeRequest(g)

			stdout, _, err := universal.Cluster.Exec("", "", AppModeTcpSink, "head", "-1", "/nc.out")
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
       - type: Tcp
         tcp:
           format:
             type: Plain
             plain: '%s'
           address: "%s:9999"
`, trafficLogFormat, tcpSinkDockerName)
		Expect(YamlUniversal(yaml)(universal.Cluster)).To(Succeed())

		makeRequest := func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, AppModeDemoClient, "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}
		src, dst := expectTrafficLogged(makeRequest)

		Expect(src).To(Equal(AppModeDemoClient))
		Expect(dst).To(Equal("test-server"))
	})

	It("should log outgoing traffic to real MeshService", func() {
		yaml := fmt.Sprintf(`
type: MeshAccessLog
name: client-outgoing-real-ms
mesh: meshaccesslog
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
        sectionName: main-port
      default:
        backends:
          - type: Tcp
            tcp:
              format:
                type: Plain
                plain: '%s'
              address: "%s:9999"
`, trafficLogFormat, tcpSinkDockerName)
		Expect(YamlUniversal(yaml)(universal.Cluster)).To(Succeed())

		makeRequest := func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, AppModeDemoClient, "test-server.universal.ms",
			)
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
       - type: Tcp
         tcp:
           format:
             type: Json
             json:
             - key: Source
               value: '%%KUMA_SOURCE_SERVICE%%'
             - key: Destination
               value: '%%KUMA_DESTINATION_SERVICE%%'
             - key: Start
               value: '%%START_TIME(%%s)%%'
             - key: HeaderCamel
               value: '%%REQ(X-Test)%%'
             - key: HeaderLower
               value: '%%REQ(x-test)%%'
             - key: HeaderCrazy
               value: '%%REQ(X-TeSt)%%'
           address: "%s:9999"
`, tcpSinkDockerName)
		Expect(YamlUniversal(yaml)(universal.Cluster)).To(Succeed())

		var log struct {
			Source      string
			Destination string
			Start       string
			HeaderCamel string
			HeaderLower string
			HeaderCrazy string
		}
		headerValue := "headervalue"
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, AppModeDemoClient, "test-server.mesh",
				client.WithHeader("X-TeSt", headerValue),
			)
			g.Expect(err).ToNot(HaveOccurred())

			stdout, _, err := universal.Cluster.Exec("", "", AppModeTcpSink, "head", "-1", "/nc.out")
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(json.Unmarshal([]byte(stdout), &log)).To(Succeed())
		}, "30s", "1s").Should(Succeed())

		Expect(log.Source).To(Equal(AppModeDemoClient))
		Expect(log.Destination).To(Equal("test-server"))
		Expect(log.HeaderCamel).To(Equal(headerValue))
		Expect(log.HeaderLower).To(Equal(headerValue))
		Expect(log.HeaderCrazy).To(Equal(headerValue))
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
       - type: Tcp
         tcp:
           format:
             type: Plain
             plain: '%s'
           address: "%s:9999"
`, trafficLogFormat, tcpSinkDockerName)
		Expect(YamlUniversal(yaml)(universal.Cluster)).To(Succeed())

		// 52 is empty response but the TCP connection succeeded
		makeRequest := func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, AppModeDemoClient, externalServiceDockerName,
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Equal(52))
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
       - type: Tcp
         tcp:
           format:
             type: Plain
             plain: '%s'
           address: "%s:9999"
`, trafficLogFormat, tcpSinkDockerName)
		Expect(YamlUniversal(externalService)(universal.Cluster)).To(Succeed())
		Expect(YamlUniversal(accesslog)(universal.Cluster)).To(Succeed())

		// 52 is empty response but the TCP connection succeeded
		makeRequest := func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, AppModeDemoClient, "ext-service.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Equal(52))
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
       - type: Tcp
         tcp:
           format:
             type: Plain
             plain: '%s'
           address: "%s:9999"
`, trafficLogFormat, tcpSinkDockerName)

		Expect(YamlUniversal(yaml)(universal.Cluster)).To(Succeed())

		makeRequest := func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, AppModeDemoClient, "test-server.mesh",
			)
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
       - type: Tcp
         tcp:
           format:
             type: Plain
             plain: '%s'
           address: "%s:9999"
`, trafficLogFormat, tcpSinkDockerName)
		Expect(YamlUniversal(yaml)(universal.Cluster)).To(Succeed())

		makeRequest := func(g Gomega) {
			gateway.ProxySimpleRequests(universal.Cluster, "echo-v1",
				GatewayAddressPort("edge-gateway", 8080), "example.kuma.io")
		}
		src, dst := expectTrafficLogged(makeRequest)

		Expect(src).To(Equal("edge-gateway"))
		Expect(dst).To(Equal("*"))
	})
}
