package helm

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ZoneAndGlobalInUniversalModeWithHelmChart() {
	var globalCluster, zoneCluster Cluster
	var global, zone ControlPlane

	BeforeAll(func() {
		var err error
		globalCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		zoneCluster = NewK8sCluster(NewTestingT(), Kuma2, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		err = NewClusterSetup().
			Install(Namespace(Config.KumaNamespace)).
			Install(postgres.Install(Kuma1,
				postgres.WithK8sNamespace(Config.KumaNamespace),
				postgres.WithUsername("mesh"),
				postgres.WithPassword("mesh"),
				postgres.WithDatabase("mesh"),
				postgres.WithPrimaryName("postgres"),
			)).
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: postgres
  namespace: %s
type: Opaque
stringData:
  password: "mesh"
`, Config.KumaNamespace))).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitPodsAvailableWithLabel(Config.KumaNamespace, "app.kubernetes.io/name", "postgresql")(globalCluster)).To(Succeed())

		err = NewClusterSetup().
			Install(Kuma(core.Global,
				// it's required because we check if Kuma is ready,
				// and we use "kubectl get mesh" which is not available in universal mode
				WithSkipDefaultMesh(true),
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithCPReplicas(2),
				WithHelmOpt("controlPlane.environment", "universal"),
				WithHelmOpt("controlPlane.envVars.KUMA_STORE_POSTGRES_HOST", "postgres-release-postgresql"),
				WithHelmOpt("controlPlane.envVars.KUMA_STORE_POSTGRES_PORT", "5432"),
				WithHelmOpt("controlPlane.envVars.KUMA_STORE_POSTGRES_USER", "mesh"),
				WithHelmOpt("controlPlane.envVars.KUMA_STORE_POSTGRES_DB_NAME", "mesh"),
				WithHelmOpt("controlPlane.secrets.postgresPassword.Secret", "postgres"),
				WithHelmOpt("controlPlane.secrets.postgresPassword.Key", "password"),
				WithHelmOpt("controlPlane.secrets.postgresPassword.Env", "KUMA_STORE_POSTGRES_PASSWORD"),
			)).
			Install(MeshUniversal("default")).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())

		global = globalCluster.GetKuma()
		Expect(global).ToNot(BeNil())

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithGlobalAddress(global.GetKDSServerAddress()),
				WithHelmOpt("ingress.enabled", "true"),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default", TestNamespace)).
			Install(testserver.Install()).
			Setup(zoneCluster)
		Expect(err).ToNot(HaveOccurred())

		zone = zoneCluster.GetKuma()
		Expect(zone).ToNot(BeNil())
	})

	E2EAfterAll(func() {
		Expect(zoneCluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(globalCluster.DeleteKuma()).To(Succeed())
		Expect(zoneCluster.DeleteKuma()).To(Succeed())
		Expect(globalCluster.DismissCluster()).To(Succeed())
		Expect(zoneCluster.DismissCluster()).To(Succeed())
	})

	It("should deploy Zone and Global on 2 clusters", func() {
		// mesh is synced to zone
		Eventually(func() string {
			output, err := zoneCluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			Expect(err).ToNot(HaveOccurred())
			return output
		}, "5s", "500ms").Should(ContainSubstring("default"))

		// and dataplanes are synced to global
		Eventually(func() string {
			output, err := globalCluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			Expect(err).ToNot(HaveOccurred())
			return output
		}, "5s", "500ms").Should(ContainSubstring("kuma-2-zone.demo-client"))
	})

	It("communication in between apps in zone works", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectResponse(zoneCluster, "demo-client", "http://test-server_kuma-test_svc_80.mesh",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
