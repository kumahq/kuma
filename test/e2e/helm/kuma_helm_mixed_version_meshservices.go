package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
)

// ZonesStayExclusiveBehindNewGlobal reproduces
// https://github.com/kumahq/kuma/issues/18868: a 3.0 global control plane
// reserves the `meshServices` field on Mesh, so it syncs Meshes without it to
// zones that still run 2.14. A 2.14 zone reads the missing field as
// `meshServices.mode: Disabled` and tears down all MeshService traffic: it
// deletes every generated MeshService, skips mesh-scoped zone proxy listeners
// and stops resolving `*.svc.mesh.local`.
//
// The spec pins both zones to the last 2.14.x release while Global runs the
// current build, applies a Mesh without `meshServices` - the state a normal
// 3.0 upgrade leaves behind, so only the KDS mapper can put a mode on the
// synced Mesh - and then asserts the zone keeps behaving as Exclusive: the
// synced Mesh carries the mode, MeshServices keep existing, and in-zone and
// cross-zone MeshService traffic keeps flowing.
//
// Global itself is not upgraded inside this spec: the framework has no
// mechanism to run an old Universal kuma-cp, so Global starts on the current
// build. That is equivalent for what is being tested - KDS does not care how
// the mixed state was reached, only how Global serves 2.14 zones while it
// lasts. Both zones are Kubernetes because only Kubernetes zones can be
// version-pinned, via old Helm charts; a Universal zone would need the same
// missing old-binary mechanism.
func ZonesStayExclusiveBehindNewGlobal() {
	meshName := "ms-exclusive"
	identityName := "ms-exclusive-identity"
	namespace := "ms-exclusive-ns"
	// <chart name>-<mesh>-ingress, see kuma.mesh.zoneproxy.name in the chart
	zoneIngressApp := fmt.Sprintf("kuma-%s-ingress", meshName)

	var global, zoneK8s1, zoneK8s2 Cluster
	var globalCP ControlPlane

	BeforeEach(func() {
		global = NewUniversalCluster(NewTestingT(), Kuma1, Silent)
		zoneK8s1 = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		zoneK8s2 = NewK8sCluster(NewTestingT(), Kuma3, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		Expect(NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(global)).To(Succeed())
		globalCP = global.GetKuma()
		Expect(globalCP).ToNot(BeNil())
	})

	AfterEachFailure(func() {
		DebugUniversal(global, "default")
		DebugKube(zoneK8s1, "default", namespace)
		DebugKube(zoneK8s2, "default", namespace)
	})

	E2EAfterEach(func() {
		ControlPlaneAssertions(global)
		ControlPlaneAssertions(zoneK8s1)
		ControlPlaneAssertions(zoneK8s2)
		grp := sync.WaitGroup{}
		grp.Add(3)
		go func() {
			defer GinkgoRecover()
			defer grp.Done()
			Expect(zoneK8s1.DeleteNamespace(namespace)).To(Succeed())
			Expect(zoneK8s1.DeleteKuma()).To(Succeed())
			Expect(zoneK8s1.DismissCluster()).To(Succeed())
		}()
		go func() {
			defer GinkgoRecover()
			defer grp.Done()
			Expect(zoneK8s2.DeleteNamespace(namespace)).To(Succeed())
			Expect(zoneK8s2.DeleteKuma()).To(Succeed())
			Expect(zoneK8s2.DismissCluster()).To(Succeed())
		}()
		go func() {
			defer GinkgoRecover()
			defer grp.Done()
			Expect(global.DeleteKuma()).To(Succeed())
			Expect(global.DismissCluster()).To(Succeed())
		}()
		grp.Wait()
	})

	// getFromZone reads straight off the zone's API, because kumactl would
	// unmarshal the response into its own Mesh model and drop the field this
	// spec is about before the assertion could see it.
	getFromZone := func(cluster Cluster, path string) (string, error) {
		req, err := http.NewRequestWithContext(
			context.Background(), http.MethodGet,
			cluster.GetKuma().GetAPIServerAddress()+path, http.NoBody,
		)
		if err != nil {
			return "", err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("GET %s returned %d: %s", path, resp.StatusCode, string(body))
		}
		return string(body), nil
	}

	getMeshFromZone := func(cluster Cluster) (string, error) {
		return getFromZone(cluster, "/meshes/"+meshName)
	}

	meshServicesInZone := func(cluster Cluster) (int, error) {
		body, err := getFromZone(cluster, "/meshes/"+meshName+"/meshservices")
		if err != nil {
			return 0, err
		}
		list := struct {
			Total int `json:"total"`
		}{}
		if err := json.Unmarshal([]byte(body), &list); err != nil {
			return 0, err
		}
		return list.Total, nil
	}

	DescribeTable("zone on an older minor keeps MeshService Exclusive mode",
		func(version string) {
			By("Apply a Mesh without meshServices on the 3.0 global")
			err := NewClusterSetup().
				Install(YamlUniversal(fmt.Sprintf(`
type: Mesh
name: %s
`, meshName))).
				Install(MeshIdentityBundled(meshName, identityName)).
				Install(MeshTrafficPermissionAllowAllUniversalWorkloadIdentity(
					meshName,
					MeshIdentityTrustDomains(meshName, zoneK8s1, zoneK8s2)...,
				)).
				// The default generators only name synced (cross-zone)
				// MeshServices on Kubernetes zones. This one names local ones,
				// the way `ledger.svc.mesh.local` resolved in the issue.
				Install(YamlUniversal(fmt.Sprintf(`
type: HostnameGenerator
name: ms-exclusive-local
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.svc.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/mesh: %s
        kuma.io/env: kubernetes
        k8s.kuma.io/is-headless-service: "false"
`, meshName))).
				Setup(global)
			Expect(err).ToNot(HaveOccurred())

			installZone := func(zone Cluster, releaseName string) {
				err := NewClusterSetup().
					Install(Kuma(core.Zone,
						WithInstallationMode(HelmInstallationMode),
						WithHelmChartPath(Config.HelmChartName),
						WithHelmReleaseName(releaseName),
						WithHelmChartVersion(version),
						WithGlobalAddress(globalCP.GetKDSServerAddress()),
						WithHelmOpt("meshes[0].name", meshName),
						WithHelmOpt("meshes[0].ingress.enabled", "true"),
						// The chart defaults the ingress Service to LoadBalancer,
						// which never gets an address on k3d.
						WithHelmOpt("meshes[0].ingress.service.type", "NodePort"),
						WithoutHelmOpt("global.image.tag"),
					)).
					// The mesh zone proxy is a Dataplane in the test's mesh, so
					// the mesh has to reach this zone over KDS before its
					// readiness can pass.
					Install(func(c Cluster) error {
						return WaitForMesh(meshName, []Cluster{c})
					}).
					Install(WaitNumPods(Config.KumaNamespace, 1, zoneIngressApp)).
					Install(WaitPodsAvailable(Config.KumaNamespace, zoneIngressApp)).
					Setup(zone)
				Expect(err).ToNot(HaveOccurred())
			}

			By("Install both zones on 2.14: " + version)
			installZone(zoneK8s1, fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueID())))
			installZone(zoneK8s2, fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueID())))

			By("Deploy workloads: test-server in kuma-2, clients in both zones")
			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
				)).
				Install(democlient.Install(
					democlient.WithNamespace(namespace),
					democlient.WithMesh(meshName),
				)).
				Setup(zoneK8s1)
			Expect(err).ToNot(HaveOccurred())

			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(democlient.Install(
					democlient.WithNamespace(namespace),
					democlient.WithMesh(meshName),
				)).
				Setup(zoneK8s2)
			Expect(err).ToNot(HaveOccurred())

			By("Synced Mesh on each 2.14 zone still carries meshServices.mode: Exclusive")
			for _, zone := range []Cluster{zoneK8s1, zoneK8s2} {
				Eventually(func(g Gomega) {
					mesh, err := getMeshFromZone(zone)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(mesh).To(ContainSubstring(`"meshServices"`),
						"zone %s lost the meshServices field, so it falls back to Disabled", zone.Name())
					g.Expect(mesh).To(ContainSubstring(`"Exclusive"`),
						"zone %s does not see the mesh as Exclusive", zone.Name())
				}, "60s", "1s").Should(Succeed())
			}

			By("MeshServices keep being served in both zones")
			for _, zone := range []Cluster{zoneK8s1, zoneK8s2} {
				Eventually(func(g Gomega) {
					total, err := meshServicesInZone(zone)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(total).To(BeNumerically(">", 0),
						"zone %s deleted its MeshServices", zone.Name())
				}, "60s", "1s").Should(Succeed())
			}

			By("MeshService DNS resolves and serves in-zone traffic")
			Eventually(func(g Gomega) {
				g.Expect(client.CollectEchoResponse(
					zoneK8s1, "demo-client",
					fmt.Sprintf("http://test-server.%s.svc.mesh.local:80", namespace),
					client.FromKubernetesPod(namespace, "demo-client"),
				)).To(HaveField("Instance", ContainSubstring("test-server")))
			}, "60s", "1s").Should(Succeed())

			By("Distribute MeshTrusts for cross-zone mTLS")
			Expect(DistributeMeshTrusts(global, meshName, identityName, zoneK8s1, zoneK8s2)).To(Succeed())

			By("Cross-zone MeshService traffic keeps working")
			Eventually(func(g Gomega) {
				g.Expect(client.CollectEchoResponse(
					zoneK8s2, "demo-client",
					fmt.Sprintf("http://test-server.%s.svc.kuma-2.mesh.local:80", namespace),
					client.FromKubernetesPod(namespace, "demo-client"),
				)).To(HaveField("Instance", ContainSubstring("test-server")))
			}, "3m", "1s").Should(Succeed())

			By("Zones do not tear MeshServices down over time")
			Consistently(func(g Gomega) {
				for _, zone := range []Cluster{zoneK8s1, zoneK8s2} {
					total, err := meshServicesInZone(zone)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(total).To(BeNumerically(">", 0),
						"zone %s deleted its MeshServices", zone.Name())
				}
			}, "30s", "5s").Should(Succeed())
		},
		EntryDescription("from version: %s"),
		// Only the last 2.14.x release is a supported upgrade path onto this
		// (3.0-line) master, same as UpgradingZoneWithHelmChart.
		SupportedVersionEntriesAtLeast("2.14.0"),
	)
}
