package connectivity

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
	"github.com/kumahq/kuma/test/framework/kumactl"
)

func AvailableServices() {
	statefulClusterName := "statefulCluster"
	meshName := "available-services"

	var statefulCluster *UniversalCluster

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())

		statefulCluster = NewUniversalCluster(NewTestingT(), statefulClusterName, Silent)
		Expect(postgres.Install(statefulClusterName)(statefulCluster)).To(Succeed())
		err = NewClusterSetup().
			Install(Kuma(
				core.Zone,
				WithGlobalAddress(multizone.Global.GetKuma().GetKDSServerAddress()),
				WithPostgres(postgres.From(statefulCluster, statefulClusterName).GetEnvVars()),
				WithEnv("KUMA_METRICS_DATAPLANE_IDLE_TIMEOUT", "10s"),
			)).
			Install(MultipleIngressUniversal(UniversalZoneIngressPort, multizone.Global.GetKuma().GenerateZoneIngressToken)).
			Install(MultipleIngressUniversal(UniversalZoneIngressPort+1, multizone.Global.GetKuma().GenerateZoneIngressToken)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"}))).
			Setup(statefulCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(statefulCluster, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(statefulCluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
		Expect(statefulCluster.DismissCluster()).To(Succeed())
	})

	getIngress := func(g Gomega, kumactl *kumactl.KumactlOptions, port int) *mesh_proto.ZoneIngress {
		out, err := kumactl.RunKumactlAndGetOutput("get", "zone-ingress", fmt.Sprintf("%s-%d", AppIngress, port), "-o", "json")
		g.Expect(err).ToNot(HaveOccurred())
		ingress, err := rest.JSON.Unmarshal([]byte(out), core_mesh.ZoneIngressResourceTypeDescriptor)
		g.Expect(err).ToNot(HaveOccurred())
		return ingress.GetSpec().(*mesh_proto.ZoneIngress)
	}

	It("is always updated on all ZoneIngresses, even if they are offline", func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(multizone.UniZone1, "demo-client", "http://test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("uni-test-server"))
		}, "30s", "1s").Should(Succeed())
		kumactl := statefulCluster.GetKumactlOptions()

		Consistently(func(g Gomega) {
			ingress := getIngress(g, kumactl, UniversalZoneIngressPort)
			g.Expect(ingress.GetAvailableServices()).To(HaveLen(1))
		}).Should(Succeed())

		Consistently(func(g Gomega) {
			ingress := getIngress(g, kumactl, UniversalZoneIngressPort+1)
			g.Expect(ingress.GetAvailableServices()).To(HaveLen(1))
		}).Should(Succeed())

		// Kill ingress
		Expect(statefulCluster.GetApp(AppModeCP).KillMainApp()).To(Succeed())
		// Kill CP
		Expect(statefulCluster.DeleteApp(fmt.Sprintf("%s-%d", AppIngress, UniversalZoneIngressPort))).To(Succeed())
		// Bring back CP
		Expect(statefulCluster.GetApp(AppModeCP).StartMainApp()).To(Succeed())
		// Kill app
		Expect(statefulCluster.DeleteApp("test-server")).To(Succeed())
		Eventually(func(g Gomega) {
			g.Expect(kumactl.KumactlDelete("dataplane", "test-server", meshName)).To(Succeed())
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			ingress := getIngress(g, kumactl, UniversalZoneIngressPort+1)
			g.Expect(ingress.GetAvailableServices()).To(BeEmpty())
		}, "30s", "3s").Should(Succeed())

		Eventually(func(g Gomega) {
			ingress := getIngress(g, kumactl, UniversalZoneIngressPort)
			g.Expect(ingress.GetAvailableServices()).To(BeEmpty())
		}, "30s", "3s").Should(Succeed())
	})
}
