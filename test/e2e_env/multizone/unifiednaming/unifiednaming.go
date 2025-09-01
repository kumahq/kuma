package unifiednaming

import (
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
	"github.com/kumahq/kuma/test/framework/portforward"
	"github.com/kumahq/kuma/test/server/types"
)

const (
	meshName           = "unified-naming"
	namespace          = "unified-naming"
	containerPatchName = "enable-feature-unified-resource-naming"
)

func UnifiedNaming() {
	containerPatch := fmt.Sprintf(`apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  name: %s
  namespace: %s
spec:
  sidecarPatch:
  - op: add
    path: /env/-
    value: '{
      "name": "KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED",
      "value": "true"
    }'`, containerPatchName, Config.KumaNamespace)

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(samples.MeshMTLSBuilder().
				WithName(meshName).
				WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())

		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		wg := &errgroup.Group{}

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			SetupInGroup(multizone.KubeZone1, wg)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(YamlK8s(containerPatch)).
			SetupInGroup(multizone.KubeZone2, wg)

		Expect(wg.Wait()).To(Succeed())
	})

	BeforeEach(func() {
		wg := &errgroup.Group{}

		NewClusterSetup().
			Install(appsKube(multizone.KubeZone1, false, "test-server")).
			SetupInGroup(multizone.KubeZone1, wg)

		NewClusterSetup().
			Install(appsKube(multizone.KubeZone2, false)).
			SetupInGroup(multizone.KubeZone2, wg)

		NewClusterSetup().
			Install(appsUni(multizone.UniZone1)).
			SetupInGroup(multizone.UniZone1, wg)

		Expect(wg.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterEach(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("enables unified resource naming per app without breaking cross-zone connectivity", func() {
		By("verifying unified naming is disabled by default in Kube Zone 2")
		assertUnifiedNamingFlag(multizone.KubeZone2, false)

		By("verifying connectivity across zones and environments (Kube<->Kube and Kube<->Universal) with unified naming disabled")
		assertConnectivity()

		By("enabling unified naming on test-server via ContainerPatch in Kube Zone 2")
		Expect(NewClusterSetup().
			Install(appsKube(multizone.KubeZone2, true, "test-server")).
			Setup(multizone.KubeZone2),
		).To(Succeed())

		Consistently(multizone.KubeZone2.GetKumaCPLogs(), "5s", "1s").
			ShouldNot(ContainElement(ContainSubstring("loading container patches failed")))

		assertUnifiedNamingFlag(multizone.KubeZone2, true, "test-server")

		By("verifying connectivity across zones and environments remains healthy when only test-server has unified naming enabled")
		assertConnectivity()

		By("enabling unified naming on demo-client via ContainerPatch in Kube Zone 2")
		Expect(NewClusterSetup().
			Install(appsKube(multizone.KubeZone2, true, "demo-client")).
			Setup(multizone.KubeZone2),
		).To(Succeed())

		Consistently(multizone.KubeZone2.GetKumaCPLogs(), "5s", "1s").
			ShouldNot(ContainElement(ContainSubstring("loading container patches failed")))

		By("verifying unified naming is enabled on both apps in Kube Zone 2")
		assertUnifiedNamingFlag(multizone.KubeZone2, true)

		By("verifying connectivity across zones and environments remains healthy when both apps have unified naming enabled")
		assertConnectivity()

		By("disabling unified naming on test-server in Kube Zone 2 (redeploy without ContainerPatch)")
		Expect(NewClusterSetup().
			Install(appsKube(multizone.KubeZone2, false, "test-server")).
			Setup(multizone.KubeZone2),
		).To(Succeed())

		By("verifying unified naming is enabled on demo-client and disabled on test-server in Kube Zone 2")
		assertUnifiedNamingFlag(multizone.KubeZone2, true, "demo-client")
		assertUnifiedNamingFlag(multizone.KubeZone2, false, "test-server")

		By("verifying connectivity across zones and environments remains healthy with mixed settings (client enabled, server disabled)")
		assertConnectivity()
	})
}

func assertUnifiedNamingFlag(zone *K8sCluster, unifiedNaming bool, apps ...string) {
	GinkgoHelper()

	flagMatcher := BeFalse()
	statMatcher := BeEmpty()
	if unifiedNaming {
		flagMatcher = BeTrue()
		statMatcher = Not(BeEmpty())
	}

	if len(apps) == 0 {
		apps = []string{"test-server", "demo-client"}
	}

	wg := &sync.WaitGroup{}

	for i := range apps {
		app := apps[i]

		spec := portforward.Spec{
			AppName:    app,
			Namespace:  namespace,
			RemotePort: 9901,
		}

		wg.Go(func() {
			defer GinkgoRecover()

			Expect(zone.WaitApp(app, namespace, 1)).To(Succeed())

			admin, err := zone.GetOrCreateAdminTunnel(spec)
			Expect(err).ToNot(HaveOccurred())

			stats, err := admin.GetStats("cluster.kri_msvc_unified-naming")
			Expect(err).ToNot(HaveOccurred())
			Expect(stats.Stats).To(statMatcher)

			xds, err := admin.GetConfigDump()
			Expect(err).ToNot(HaveOccurred())

			Expect(xds.Metadata().HasFeature(xds_types.FeatureUnifiedResourceNaming)).
				To(flagMatcher)

			zone.ClosePortForwards(spec)
		})
	}

	wg.Wait()
}

func buildAssertRequestFn(from Cluster, to Cluster, url string) func() {
	GinkgoHelper()

	args := []any{from, "demo-client", url}
	if _, ok := from.(*K8sCluster); ok {
		args = append(args, client.FromKubernetesPod(namespace, "demo-client"))
	}

	getInstance := func(r types.EchoResponse) string { return r.Instance }
	expectedIns := fmt.Sprintf("test-server-%s", to.ZoneName())
	description := fmt.Sprintf("demo-client (%s) should reach %s (%s)", from.ZoneName(), url, to.ZoneName())

	return func() {
		defer GinkgoRecover()

		Eventually(client.CollectEchoResponse, "30s", "1s").
			WithArguments(args...).
			Should(WithTransform(getInstance, Equal(expectedIns)), description)
	}
}

func assertConnectivity() {
	GinkgoHelper()

	urlKubeL := fmt.Sprintf("http://test-server.%s.svc.cluster.local", namespace)
	urlKube1 := fmt.Sprintf("http://test-server.%s.svc.%s.mesh.local", namespace, multizone.KubeZone1.ZoneName())
	urlKube2 := fmt.Sprintf("http://test-server.%s.svc.%s.mesh.local", namespace, multizone.KubeZone2.ZoneName())
	urlUniv1 := fmt.Sprintf("http://test-server.svc.%s.mesh.local", multizone.UniZone1.ZoneName())

	wg := &sync.WaitGroup{}

	wg.Go(buildAssertRequestFn(multizone.KubeZone2, multizone.KubeZone2, urlKubeL))
	wg.Go(buildAssertRequestFn(multizone.KubeZone2, multizone.KubeZone1, urlKube1))
	wg.Go(buildAssertRequestFn(multizone.KubeZone2, multizone.UniZone1, urlUniv1))
	wg.Go(buildAssertRequestFn(multizone.UniZone1, multizone.KubeZone2, urlKube2))

	wg.Wait()
}

func appsKube(c *K8sCluster, unifiedNaming bool, apps ...string) InstallFunc {
	var install []InstallFunc

	if len(apps) == 0 {
		apps = []string{"test-server", "demo-client"}
	}

	annotations := map[string]string{}
	if unifiedNaming {
		annotations[metadata.KumaContainerPatches] = containerPatchName
	}

	for _, app := range apps {
		switch app {
		case "demo-client":
			install = append(install, democlient.Install(
				democlient.WithName(app),
				democlient.WithNamespace(namespace),
				democlient.WithMesh(meshName),
				democlient.WithPodAnnotations(annotations),
			))
		case "test-server":
			install = append(install, testserver.Install(
				testserver.WithName(app),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", fmt.Sprintf("%s-%s", app, c.ZoneName())),
				testserver.WithPodAnnotations(annotations),
			))
		}

	}

	return Parallel(install...)
}

func appsUni(c *UniversalCluster, apps ...string) InstallFunc {
	var fns []InstallFunc

	if len(apps) == 0 {
		apps = []string{"test-server", "demo-client"}
	}

	for _, app := range apps {
		switch app {
		case "demo-client":
			fns = append(fns, DemoClientUniversal(app, meshName, WithTransparentProxy(true)))
		case "test-server":
			args := []string{"echo", "--instance", fmt.Sprintf("%s-%s", app, c.ZoneName())}
			fns = append(fns, TestServerUniversal(app, meshName, WithArgs(args)))
		}
	}

	return Parallel(fns...)
}
