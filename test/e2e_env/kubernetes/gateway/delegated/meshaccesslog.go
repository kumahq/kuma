package delegated

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshAccessLog(config *Config) func() {
	GinkgoHelper()

	return func() {
		type log struct {
			Source      string
			Destination string
			Start       string
		}

		parseTimestamp := func(value string) (time.Time, error) {
			timestamp, err := strconv.Atoi(value)
			if err != nil {
				return time.Time{}, err
			}
			return time.Unix(int64(timestamp), 0), nil
		}

		parseLogs := func(stdout string) []log {
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

		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

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

			Eventually(func(g Gomega) {
				// when
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())

				// then
				demoClientPod, err := framework.PodOfApp(
					kubernetes.Cluster,
					"demo-client",
					config.NamespaceOutsideMesh,
				)
				g.Expect(err).ToNot(HaveOccurred())

				logs, err := kubernetes.Cluster.GetPodLogs(demoClientPod, v1.PodLogOptions{})
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(parseLogs(logs)).To(ContainElement(
					And(
						HaveField("Start", WithTransform(parseTimestamp, BeTemporally("~", time.Now(), time.Hour))),
						HaveField("Source", fmt.Sprintf("%s-gateway-admin_%s_svc_8444", config.Mesh, config.Namespace)),
						HaveField("Destination", fmt.Sprintf("test-server_%s_svc_80", config.Namespace)),
					),
				))
			}, "30s", "1s").Should(Succeed())
		})
	}
}
