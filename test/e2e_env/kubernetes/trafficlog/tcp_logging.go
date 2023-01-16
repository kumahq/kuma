package trafficlog

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	. "github.com/kumahq/kuma/test/framework"
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
				Install(DemoClientK8s(meshName, trafficNamespace)).
				Install(YamlK8s(tcpSink.String())).
				Install(WaitPodsAvailable(tcpSinkNamespace, tcpSinkAppName)).
				Install(testserver.Install(
					testserver.WithNamespace(trafficNamespace),
					testserver.WithMesh(meshName),
					testserver.WithName(testServer))).
				Setup(kubernetes.Cluster)
			Expect(err).ToNot(HaveOccurred())
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
			clientPodName, err := PodNameOfApp(kubernetes.Cluster, "demo-client", trafficNamespace)
			Expect(err).ToNot(HaveOccurred())
			tcpSinkPodName, err := PodNameOfApp(kubernetes.Cluster, tcpSinkAppName, tcpSinkNamespace)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				_, _, err = kubernetes.Cluster.Exec(trafficNamespace, clientPodName,
					AppModeDemoClient, "curl", "-v", "--fail", "test-server")
				if err != nil {
					return err
				}
				stdout, _, err = kubernetes.Cluster.Exec(tcpSinkNamespace, tcpSinkPodName,
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
			}).ShouldNot(HaveOccurred())
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
