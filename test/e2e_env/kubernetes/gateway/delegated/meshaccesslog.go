package delegated

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshAccessLog(config *Config) func() {
	GinkgoHelper()

	return func() {
		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshAccessLogResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should log incoming traffic", func() {
			// given
			meshAccessLog := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshAccessLog
metadata:
  name: mal
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
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
              address: demo-client.%s.svc.cluster.local:3000
`, config.CpNamespace, config.Mesh, config.NamespaceOutsideMesh)

			Expect(framework.YamlK8s(meshAccessLog)(kubernetes.Cluster)).To(Succeed())

			// when
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			// then
			Eventually(func(g Gomega) {
				demoClientPod, err := framework.PodOfApp(
					kubernetes.Cluster,
					"demo-client",
					config.NamespaceOutsideMesh,
				)
				g.Expect(err).ToNot(HaveOccurred())

				logs, err := kubernetes.Cluster.GetPodLogs(demoClientPod)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(parseLogs(logs)).To(ContainElement(
					And(
						HaveField("Start", BeTemporally("~", time.Now(), time.Hour)),
						HaveField("Source", fmt.Sprintf("delegated-gateway-admin_%s_svc_8444", config.Namespace)),
						HaveField("Destination", fmt.Sprintf("test-server_%s_svc_80", config.Namespace)),
					),
				))
			}, "30s", "1s").Should(Succeed())
		})
	}
}

type log struct {
	Source      string
	Destination string
	Start       time.Time
}

func (u *log) UnmarshalJSON(b []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	if raw["Source"] == "" || raw["Destination"] == "" || raw["Start"] == "" {
		return fmt.Errorf("missing expected fields")
	}

	timestamp, err := strconv.Atoi(raw["Start"])
	if err != nil {
		return err
	}

	u.Source = raw["Source"]
	u.Destination = raw["Destination"]
	u.Start = time.Unix(int64(timestamp), 0)

	return nil
}

func parseLogs(stdout string) []log {
	GinkgoHelper()

	var logs []log

	for _, line := range strings.Split(stdout, "\n") {
		var l log
		if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &l); err != nil {
			continue
		}
		logs = append(logs, l)
	}

	return logs
}
