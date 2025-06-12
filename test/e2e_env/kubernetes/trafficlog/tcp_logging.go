package trafficlog

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func TCPLogging() {
	Describe("TrafficLog Logging to TCP", func() {
		meshName := "trafficlog-tcp-logging"
		trafficNamespace := "trafficlog-tcp-logging"
		tcpSinkNamespace := "tcp-sink-namespace"
		tcpSinkAppName := "tcp-sink"
		testServer := "test-server"

		args := struct {
			MeshName           string
			TrafficNamespace   string
			TcpSinkNamespace   string
			TcpSinkAppName     string
			KumaUniversalImage string
		}{
			meshName, trafficNamespace, tcpSinkNamespace,
			tcpSinkAppName, Config.GetUniversalImage(),
		}

		loggingBackend := &bytes.Buffer{}
		tmpl := template.Must(template.New("loggingBackend").Parse(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: {{ .MeshName }}
spec:
  logging:
    defaultBackend: netcat
    backends:
      - name: netcat
        format: '%START_TIME(%s)%,%KUMA_SOURCE_SERVICE%,%KUMA_DESTINATION_SERVICE%'
        type: tcp
        conf:
          address: {{.TcpSinkAppName}}.{{.TcpSinkNamespace}}.svc.cluster.local:9999
`))
		Expect(tmpl.Execute(loggingBackend, &args)).To(Succeed())

		trafficLog := &bytes.Buffer{}
		tmpl = template.Must(template.New("trafficLog").Parse(`
apiVersion: kuma.io/v1alpha1
kind: TrafficLog
mesh: {{ .MeshName }}
metadata:
  name: all-traffic
spec:
  sources:
    - match:
        kuma.io/service: "*"
  destinations:
    - match:
        kuma.io/service: "*"
`))
		Expect(tmpl.Execute(trafficLog, &args)).To(Succeed())

		tcpSink := &bytes.Buffer{}
		tmpl = template.Must(template.New("tcpSink").Parse(`
apiVersion: v1
kind: Service
metadata:
  name: {{.TcpSinkAppName}}
  namespace: {{.TcpSinkNamespace}}
spec:
  ports:
    - port: 9999
      name: netcat
  selector:
    app: {{.TcpSinkAppName}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.TcpSinkAppName}}
  namespace: {{.TcpSinkNamespace}}
  labels:
    app: {{.TcpSinkAppName}}
spec:
  selector:
    matchLabels:
      app: {{.TcpSinkAppName}}
  template:
    metadata:
      labels:
        app: {{.TcpSinkAppName}}
    spec:
      containers:
        - name: {{.TcpSinkAppName}}
          image: {{.KumaUniversalImage}}
          command: ["/bin/bash"]
          args: ["-c", "/bin/netcat -lk -p 9999 > /nc.out"]
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 9999
`))

		Expect(tmpl.Execute(tcpSink, &args)).To(Succeed())

		BeforeAll(func() {
			err := NewClusterSetup().
				Install(YamlK8s(loggingBackend.String())).
				Install(YamlK8s(trafficLog.String())).
				Install(NamespaceWithSidecarInjection(trafficNamespace)).
				Install(Namespace(tcpSinkNamespace)).
				Install(democlient.Install(democlient.WithNamespace(trafficNamespace), democlient.WithMesh(meshName))).
				Install(YamlK8s(tcpSink.String())).
				Install(WaitPodsAvailable(tcpSinkNamespace, tcpSinkAppName)).
				Install(testserver.Install(
					testserver.WithNamespace(trafficNamespace),
					testserver.WithMesh(meshName),
					testserver.WithName(testServer))).
				Install(TrafficRouteKubernetes(meshName)).
				Install(TrafficPermissionKubernetes(meshName)).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEachFailure(func() {
			DebugKube(kubernetes.Cluster, meshName, trafficNamespace, tcpSinkNamespace)
		})

		E2EAfterAll(func() {
			Expect(kubernetes.Cluster.TriggerDeleteNamespace(trafficNamespace)).To(Succeed())
			Expect(kubernetes.Cluster.TriggerDeleteNamespace(tcpSinkNamespace)).To(Succeed())
			Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
		})

		It("should send a traffic log to TCP port", func() {
			var startTimeStr, src, dst string
			var err error
			var stdout string
			tcpSinkPodName, err := PodNameOfApp(kubernetes.Cluster, tcpSinkAppName, tcpSinkNamespace)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, AppModeDemoClient, "test-server",
					client.FromKubernetesPod(trafficNamespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())

				stdout, _, err = kubernetes.Cluster.Exec(tcpSinkNamespace, tcpSinkPodName,
					"tcp-sink", "tail", "-1", "/nc.out")
				g.Expect(err).ToNot(HaveOccurred())
				parts := strings.Split(stdout, ",")
				g.Expect(parts).To(HaveLen(3))
				startTimeStr, src, dst = parts[0], parts[1], parts[2]
			}).Should(Succeed())
			startTimeInt, err := strconv.Atoi(startTimeStr)
			Expect(err).ToNot(HaveOccurred())
			startTime := time.Unix(int64(startTimeInt), 0)

			Expect(startTime).To(BeTemporally("~", time.Now(), time.Minute))

			Expect(src).To(Equal("demo-client_trafficlog-tcp-logging_svc"))
			Expect(strings.TrimSpace(dst)).To(Equal("test-server_trafficlog-tcp-logging_svc_80"))
		})
	})
}
