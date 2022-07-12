package trafficlog

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func TCPLogging() {
	Describe("TrafficLog Logging to TCP", func() {
		meshName := "trafficlog-tcp-logging"
		namespace := "trafficlog-tcp-logging"
		tcpSinkNamespace := "externalservice-namespace"
		testServer := "test-server"
		loggingBackend := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
`, meshName) + `spec:
  logging:
    defaultBackend: netcat
    backends:
    - name: netcat
      format: '%START_TIME(%s)%,%KUMA_SOURCE_SERVICE%,%KUMA_DESTINATION_SERVICE%'
      type: tcp
      conf:
        address: externalservice-tcp-sink.externalservice-namespace.svc.cluster.local:9999
`

		trafficLog := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: TrafficLog
mesh: %s
metadata:
  name: all-traffic
spec:
  sources:
  - match:
      kuma.io/service: "*"
  destinations:
   - match:
      kuma.io/service: "*"
`, meshName)
		BeforeAll(func() {
			err := NewClusterSetup().
				Install(YamlK8s(loggingBackend)).
				Install(YamlK8s(trafficLog)).
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithName(testServer))).
				Install(DemoClientK8s(meshName, namespace)).
				Install(externalservice.Install(externalservice.TcpSink, []string{})).
				Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		E2EAfterAll(func() {
			Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
			Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should send a traffic log to TCP port", func() {
			var startTimeStr, src, dst string
			var err error
			var stdout string
			clientPodName, err := PodNameOfApp(env.Cluster, "demo-client", namespace)
			Expect(err).ToNot(HaveOccurred())
			tcpSinkPodName, err := PodNameOfApp(env.Cluster, "externalservice-tcp-sink", tcpSinkNamespace)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				_, _, err = env.Cluster.ExecWithRetries(namespace, clientPodName,
					AppModeDemoClient, "curl", "-v", "--fail", "test-server")
				if err != nil {
					return err
				}
				stdout, _, err = env.Cluster.Exec(tcpSinkNamespace, tcpSinkPodName,
					"tcp-sink", "head", "-1", "/nc.out")
				if err != nil {
					return err
				}
				parts := strings.Split(stdout, ",")
				if len(parts) != 3 {
					return errors.Errorf("unexpected number of fields in: %s", stdout)
				}
				startTimeStr, src, dst = parts[0], parts[1], parts[2]
				return nil
			}, "30s", "1ms").ShouldNot(HaveOccurred())
			startTimeInt, err := strconv.Atoi(startTimeStr)
			Expect(err).ToNot(HaveOccurred())
			startTime := time.Unix(int64(startTimeInt), 0)

			// Just testing that it is a timestamp, not accuracy. If it's
			// an int that would represent Unix time within an hour of now
			// it's probably a timestamp substitution.
			Expect(startTime).To(BeTemporally("~", time.Now(), time.Hour))

			Expect(src).To(Equal("demo-client_trafficlog-tcp-logging_svc"))
			Expect(strings.TrimSpace(dst)).To(Equal("test-server_trafficlog-tcp-logging_svc_80"))
		})
	})
}
