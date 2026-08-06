package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_types "github.com/kumahq/kuma/v3/api/openapi/types"
	"github.com/kumahq/kuma/v3/pkg/kds/hash"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func Inspect() {
	const meshName = "inspect"
	const identityName = "inspect-identity"

	BeforeAll(func() {
		Expect(multizone.Global.Install(MeshUniversal(meshName))).To(Succeed())
		Expect(multizone.Global.Install(MeshIdentityBundled(meshName, identityName))).To(Succeed())
		Expect(multizone.Global.Install(MeshTrafficPermissionAllowAllUniversalWorkloadIdentity(
			meshName,
			MeshIdentityTrustDomain(meshName, multizone.UniZone1),
		))).To(Succeed())
		Expect(multizone.Global.Install(YamlUniversal(fmt.Sprintf(`
type: MeshTimeout
name: inspect-timeout
mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        idleTimeout: 1h
        http:
          requestTimeout: 15s
          maxStreamDuration: 0s`, meshName)))).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(Parallel(
				TestServerUniversal("test-server", meshName,
					WithArgs([]string{"echo", "--instance", "echo"}),
				),
				zoneproxy.Install(zoneproxy.WithMesh(meshName)),
			)).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())
		// remove default
		Eventually(func() error {
			return multizone.Global.GetKumactlOptions().RunKumactl("delete", "meshtimeout", "--mesh", meshName, "mesh-timeout-all-"+meshName)
		}).Should(Succeed())
	})

	BeforeEach(func() {
		Expect(multizone.Global.Install(MeshUniversal(meshName))).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	type testCase struct {
		cluster       func() Cluster
		args          []string
		expectedOut   string
		reinstallMesh bool
	}
	GlobalCluster := func() Cluster {
		return multizone.Global
	}
	UniZone1Cluster := func() Cluster {
		return multizone.UniZone1
	}

	testServerDPPName := hash.HashedName(meshName, "test-server", "kuma-4")
	ingressName := zoneproxy.IngressName(meshName)
	egressName := zoneproxy.EgressName(meshName)
	ingressDPPName := hash.HashedName(meshName, ingressName, "kuma-4")
	egressDPPName := hash.HashedName(meshName, egressName, "kuma-4")

	Context("Dataplane", func() {
		DescribeTable("should execute envoy inspection",
			func(given testCase) {
				if given.reinstallMesh {
					Expect(multizone.Global.Install(MeshUniversal(meshName))).To(Succeed())
				}
				Eventually(func(g Gomega) {
					args := append([]string{"inspect"}, given.args...)
					out, err := given.cluster().GetKumactlOptions().RunKumactlAndGetOutput(args...)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(out).Should(ContainSubstring(given.expectedOut))
				}, "30s", "1s").Should(Succeed())
			},
			Entry("of config dump for a dataplane using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", testServerDPPName, "--type", "config-dump", "--mesh", meshName},
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("of stats for a dataplane using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", testServerDPPName, "--type", "stats", "--mesh", meshName},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a dataplane using Global CP", testCase{
				cluster:       GlobalCluster,
				args:          []string{"dataplane", testServerDPPName, "--type", "clusters", "--mesh", meshName},
				reinstallMesh: true,
				expectedOut:   `system_envoy_admin::`,
			}),
			Entry("of config dump for a dataplane using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", "test-server", "--type", "config-dump", "--mesh", meshName},
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("of stats for a dataplane using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", "test-server", "--type", "stats", "--mesh", meshName},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a dataplane using Zone CP", testCase{
				cluster:       UniZone1Cluster,
				args:          []string{"dataplane", "test-server", "--type", "clusters", "--mesh", meshName},
				reinstallMesh: true,
				expectedOut:   `system_envoy_admin::`,
			}),
			Entry("of config dump for a zone ingress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", ingressDPPName, "--type", "config-dump", "--mesh", meshName},
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("of stats for a zone ingress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", ingressDPPName, "--type", "stats", "--mesh", meshName},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a zone ingress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", ingressDPPName, "--type", "clusters", "--mesh", meshName},
				expectedOut: `system_envoy_admin::`,
			}),
			Entry("of config dump for a zone ingress using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", ingressName, "--type", "config-dump", "--mesh", meshName},
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("of stats for a zone ingress using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", ingressName, "--type", "stats", "--mesh", meshName},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a zone ingress using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", ingressName, "--type", "clusters", "--mesh", meshName},
				expectedOut: `system_envoy_admin::`,
			}),
			Entry("of config dump for a zone egress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", egressDPPName, "--type", "config-dump", "--mesh", meshName},
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("of stats for a zone egress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", egressDPPName, "--type", "stats", "--mesh", meshName},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a zone egress using Global CP", testCase{
				cluster:     GlobalCluster,
				args:        []string{"dataplane", egressDPPName, "--type", "clusters", "--mesh", meshName},
				expectedOut: `system_envoy_admin::`,
			}),
			Entry("of config dump for a zone egress using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", egressName, "--type", "config-dump", "--mesh", meshName},
				expectedOut: `"dataplane.proxyType": "dataplane"`,
			}),
			Entry("of stats for a zone egress using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", egressName, "--type", "stats", "--mesh", meshName},
				expectedOut: `server.live: 1`,
			}),
			Entry("of clusters for a zone egress using Zone CP", testCase{
				cluster:     UniZone1Cluster,
				args:        []string{"dataplane", egressName, "--type", "clusters", "--mesh", meshName},
				expectedOut: `system_envoy_admin::`,
			}),
		)

		It("match dataplanes of policy", func() {
			Eventually(func(g Gomega) {
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, multizone.Global.GetKuma().GetAPIServerAddress()+fmt.Sprintf("/meshes/%s/meshtimeouts/inspect-timeout/_resources/dataplanes", meshName), http.NoBody)
				g.Expect(err).ToNot(HaveOccurred())
				r, err := http.DefaultClient.Do(req)
				g.Expect(err).ToNot(HaveOccurred())
				defer r.Body.Close()
				g.Expect(r).To(HaveHTTPStatus(200))

				body, err := io.ReadAll(r.Body)
				g.Expect(err).ToNot(HaveOccurred())
				result := api_types.InspectDataplanesForPolicyResponse{}
				g.Expect(json.Unmarshal(body, &result)).To(Succeed())

				// The policy targets the whole Mesh, and zone proxies are
				// mesh-scoped Dataplanes, so they match it as well.
				names := []string{}
				for _, item := range result.Items {
					names = append(names, item.Name)
				}
				g.Expect(names).To(ConsistOf(testServerDPPName, ingressDPPName, egressDPPName))
				g.Expect(result.Total).To(Equal(3))
			}, "30s", "1s").Should(Succeed())
		})

		It("should execute inspect rules of dataplane", func() {
			Expect(multizone.Global.Install(MeshUniversal(meshName))).To(Succeed())
			Expect(YamlUniversal(fmt.Sprintf(`
type: MeshTimeout
name: mt1
mesh: %s
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
          maxStreamDuration: 20s`, meshName))(multizone.Global)).To(Succeed())
			Eventually(func(g Gomega) {
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, multizone.Global.GetKuma().GetAPIServerAddress()+fmt.Sprintf("/meshes/%s/dataplanes/%s/_rules", meshName, testServerDPPName), http.NoBody)
				g.Expect(err).ToNot(HaveOccurred())
				r, err := http.DefaultClient.Do(req)
				g.Expect(err).ToNot(HaveOccurred())
				defer r.Body.Close()
				g.Expect(r).To(HaveHTTPStatus(200))

				body, err := io.ReadAll(r.Body)
				g.Expect(err).ToNot(HaveOccurred())
				result := api_types.InspectRulesResponse{}
				g.Expect(json.Unmarshal(body, &result)).To(Succeed())

				g.Expect(result.Resource.Name).To(Equal(testServerDPPName))
				g.Expect(result.Rules).ToNot(BeEmpty())
				for _, rule := range result.Rules {
					if rule.Type == "MeshTimeout" {
						if rule.ToRules != nil {
							g.Expect(*rule.ToRules).To(BeEmpty())
						}
					}
				}
			}, "30s", "1s").Should(Succeed())
		})
	})
}
