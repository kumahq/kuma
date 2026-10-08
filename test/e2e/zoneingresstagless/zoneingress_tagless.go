package zoneingresstagless

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
)

const (
	meshName  = "zoneingress-tagless"
	namespace = "zoneingress-tagless"
	instance  = "kube-test-server"
)

var (
	global Cluster
	zone1  *K8sCluster
	zone4  *UniversalCluster

	backendURL = fmt.Sprintf("http://test-server.%s.svc.%s.mesh.local:80", namespace, Kuma1)
)

// ZoneIngressWithInboundTagsDisabled turns on experimental.inboundTagsDisabled on a
// Kubernetes zone that has a standalone ZoneIngress and a non-Exclusive mesh. The
// ZoneIngress must stay valid and keep routing MeshService traffic from other zones.
func ZoneIngressWithInboundTagsDisabled() {
	BeforeAll(func() {
		global = NewUniversalCluster(NewTestingT(), Kuma5, Silent)
		Expect(NewClusterSetup().
			Install(Kuma(config_core.Global)).
			Install(MTLSMeshWithMeshServicesUniversal(meshName, "Everywhere")).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(global)).To(Succeed())
		globalCP := global.GetKuma()

		group := errgroup.Group{}

		zone1 = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithIngress(),
				WithIngressEnvoyAdminTunnel(),
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
			)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("--instance", instance),
			)).
			SetupInGroup(zone1, &group)

		zone4 = NewUniversalCluster(NewTestingT(), Kuma4, Silent)
		NewClusterSetup().
			Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
			Install(DemoClientUniversal("zone4-demo-client", meshName, WithTransparentProxy(true))).
			SetupInGroup(zone4, &group)

		Expect(group.Wait()).To(Succeed())
		Expect(WaitForZoneOnline(global, Kuma1)).To(Succeed())
		Expect(WaitForZoneOnline(global, Kuma4)).To(Succeed())
		Expect(WaitForMesh(meshName, []Cluster{zone1, zone4})).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(global, meshName)
		DebugUniversal(zone4, meshName)
		DebugKube(zone1, meshName, namespace)
	})

	AfterAll(func() {
		Expect(zone1.DeleteNamespace(namespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
		Expect(zone4.DismissCluster()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("should reach the Kubernetes MeshService from another zone while inbound tags are enabled", func() {
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(zone4, "zone4-demo-client", backendURL)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal(instance))
		}, "2m", "1s").MustPassRepeatedly(5).Should(Succeed())

		// the tracker lists the backend while its inbound still has kuma.io/service
		Eventually(func(g Gomega) {
			services, err := storedAvailableServices()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(services).To(ContainElement(HaveKeyWithValue(mesh_proto.ServiceTag, fmt.Sprintf("test-server_%s_svc_80", namespace))))
		}, "1m", "1s").Should(Succeed())
	})

	It("should drop inbound tags when the zone control plane disables them", func() {
		// stop first: a rollout leaves the old pod terminating, and the new port-forward can attach to it
		Expect(zone1.StopControlPlane()).To(Succeed())
		Expect(k8s.RunKubectlContextE(zone1.GetTesting(), context.Background(), zone1.GetKubectlOptions(Config.KumaNamespace),
			"set", "env", "deployment/"+Config.KumaServiceName, "KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED=true",
		)).To(Succeed())
		Expect(zone1.RestartControlPlane()).To(Succeed())
		Expect(WaitForZoneOnline(global, Kuma1)).To(Succeed())

		Eventually(func(g Gomega) {
			names, err := dataplaneJSONPath("{.items[*].metadata.name}")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(names).ToNot(BeEmpty())
			tags, err := dataplaneJSONPath("{.items[*].spec.networking.inbound[*].tags}")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(tags).To(BeEmpty())
		}, "1m", "1s").Should(Succeed())
	})

	It("should keep the ZoneIngress valid", func() {
		Eventually(func(g Gomega) {
			logs := zone1.GetKumaCPLogs()["stdout"]
			g.Expect(logs).ToNot(ContainSubstring("couldn't update ZoneIngress"))
			g.Expect(logs).ToNot(ContainSubstring("either WithService() or WithName() should be called"))

			metrics, err := global.GetKuma().GetMetrics()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(zoneIngressNacks(metrics, Kuma1)).To(BeEmpty())

			// the backend inbound has no kuma.io/service now, so the ZoneIngress must not list it
			services, err := storedAvailableServices()
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(services).To(BeEmpty())
		}, "2m", "2s").Should(Succeed())
	})

	It("should reach the Kubernetes MeshService from another zone after the backend moves", func() {
		// a new pod has a new address, so the request succeeds only if the ingress config updates
		Expect(zone1.KillAppPod("test-server", namespace)).To(Succeed())

		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(zone4, "zone4-demo-client", backendURL)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal(instance))
		}, "2m", "1s").MustPassRepeatedly(5).Should(Succeed())

		ControlPlaneAssertions(zone1)
	})
}

// storedAvailableServices returns the tags of the available services that the
// Kubernetes store holds for the mesh of this test
func storedAvailableServices() ([]map[string]string, error) {
	out, err := k8s.RunKubectlAndGetOutputContextE(zone1.GetTesting(), context.Background(),
		zone1.GetKubectlOptions(Config.KumaNamespace), "get", "zoneingresses", "-o", "json")
	if err != nil {
		return nil, err
	}
	var list struct {
		Items []struct {
			Spec struct {
				AvailableServices []struct {
					Mesh string            `json:"mesh"`
					Tags map[string]string `json:"tags"`
				} `json:"availableServices"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no ZoneIngress in %s", Kuma1)
	}
	var tags []map[string]string
	for _, item := range list.Items {
		for _, svc := range item.Spec.AvailableServices {
			if svc.Mesh == meshName {
				tags = append(tags, svc.Tags)
			}
		}
	}
	return tags, nil
}

func dataplaneJSONPath(path string) (string, error) {
	return k8s.RunKubectlAndGetOutputContextE(zone1.GetTesting(), context.Background(),
		zone1.GetKubectlOptions(namespace), "get", "dataplanes", "-o", "jsonpath="+path)
}

// zoneIngressNacks returns the global kds_nack_total samples for ZoneIngress from the zone
func zoneIngressNacks(metrics, zone string) []string {
	var samples []string
	for line := range strings.SplitSeq(metrics, "\n") {
		if strings.HasPrefix(line, "kds_nack_total{") &&
			strings.Contains(line, `resource_type="ZoneIngress"`) &&
			strings.Contains(line, fmt.Sprintf("zone_name=%q", zone)) {
			samples = append(samples, line)
		}
	}
	return samples
}
