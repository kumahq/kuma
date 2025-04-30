package inspect

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_types "github.com/kumahq/kuma/api/openapi/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Inspect() {
	nsName := "inspect"
	meshName := "inspect"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(nsName)).
			Install(MeshKubernetes(meshName)).
			Install(democlient.Install(democlient.WithNamespace(nsName), democlient.WithMesh(meshName))).
			Install(TimeoutKubernetes(meshName)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
		// remove default meshtimeout policies
		Expect(DeleteMeshPolicyOrError(
			kubernetes.Cluster,
			v1alpha1.MeshTimeoutResourceTypeDescriptor,
			fmt.Sprintf("mesh-timeout-all-%s", meshName),
		)).To(Succeed())
		Expect(DeleteMeshPolicyOrError(
			kubernetes.Cluster,
			v1alpha1.MeshTimeoutResourceTypeDescriptor,
			fmt.Sprintf("mesh-gateways-timeout-all-%s", meshName),
		)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, nsName)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(nsName)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should return bad request on invalid name(#4985)", func() {
		_, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", "-m", meshName, "dummy-name", "--type=config-dump")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`Bad Request (bad request: name "dummy-name" must include namespace after the dot, ex. "name.namespace")`))
	})

	It("should return envoy config_dump", func() {
		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("dataplanes", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			// Dataplane names are generated, so we check for a partial match.
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("demo-client")))
		}, "10s", "250ms").Should(Succeed())

		podName, err := PodNameOfApp(kubernetes.Cluster, "demo-client", nsName)
		Expect(err).ToNot(HaveOccurred())
		dataplaneName := fmt.Sprintf("%s.%s", podName, nsName)
		stdout, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", "-m", meshName, dataplaneName, "--type=config-dump")
		Expect(err).ToNot(HaveOccurred())

		Expect(stdout).To(ContainSubstring(`"name": "inbound:passthrough:ipv4"`))
		Expect(stdout).To(ContainSubstring(`"name": "inbound:passthrough:ipv6"`))
		Expect(stdout).To(ContainSubstring(`"name": "kuma:envoy:admin"`))
		Expect(stdout).To(ContainSubstring(`"name": "outbound:passthrough:ipv4"`))
		Expect(stdout).To(ContainSubstring(`"name": "outbound:passthrough:ipv6"`))
	})

	DescribeTable("should execute inspect of policies",
		func(policyType string, policyName string) {
			Eventually(func(g Gomega) {
				r, err := http.Get(kubernetes.Cluster.GetKuma().GetAPIServerAddress() + fmt.Sprintf("/meshes/%s/timeouts/timeout-all-%s/_resources/dataplanes", meshName, meshName))
				g.Expect(err).ToNot(HaveOccurred())
				defer r.Body.Close()
				g.Expect(r).To(HaveHTTPStatus(200))

				body, err := io.ReadAll(r.Body)
				g.Expect(err).ToNot(HaveOccurred())
				result := api_types.InspectDataplanesForPolicyResponse{}
				g.Expect(json.Unmarshal(body, &result)).To(Succeed())

				g.Expect(result.Items).To(HaveLen(1))
				g.Expect(result.Total).To(Equal(1))
				g.Expect(result.Items[0].Name).To(HavePrefix("demo-client-"))
			}, "30s", "1s").Should(Succeed())
		},
		Entry("of dataplanes", "timeouts", fmt.Sprintf("timeout-all-%s", meshName)),
	)

	It("should execute inspect rules of dataplane", func() {
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1
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
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, Config.KumaNamespace, meshName))(kubernetes.Cluster)).To(Succeed())
		Eventually(func(g Gomega) {
			dataplanes, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("dataplanes", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			clientDp := ""
			for _, dp := range dataplanes {
				if strings.Contains(dp, "demo-client") {
					clientDp = dp
					break
				}
			}
			Expect(clientDp).ToNot(BeEmpty(), "no demo-client dpp found")

			r, err := http.Get(kubernetes.Cluster.GetKuma().GetAPIServerAddress() + fmt.Sprintf("/meshes/%s/dataplanes/%s/_rules", meshName, clientDp))
			g.Expect(err).ToNot(HaveOccurred())
			defer r.Body.Close()
			g.Expect(r).To(HaveHTTPStatus(200))

			body, err := io.ReadAll(r.Body)
			g.Expect(err).ToNot(HaveOccurred())
			result := api_types.InspectRulesResponse{}
			g.Expect(json.Unmarshal(body, &result)).To(Succeed())

			g.Expect(result.Resource.Name).To(Equal(clientDp))
			g.Expect(result.Rules).ToNot(BeEmpty())
			for _, rule := range result.Rules {
				if rule.Type == "MeshTimeout" {
					g.Expect(rule.ToRules).ToNot(BeNil())
					g.Expect(*rule.ToRules).ToNot(BeEmpty())
					g.Expect((*rule.ToRules)[0].Origin[0].Name).To(Equal(fmt.Sprintf("mt1.%s", Config.KumaNamespace)))
				}
			}
		}, "30s", "1s").Should(Succeed())
	})
}
