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
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
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
				// WithSkipDefaultMesh is required because we check if Kuma is ready by using "kubectl get mesh"
				// here in the framework https://github.com/kumahq/kuma/blob/1633d34ad116dd1e618f4a27dd1526f5ff7d8bde/test/framework/k8s_cluster.go#L564
				// but on universal mode we use postgres to manage resources so without this it will fail making the test suite fail
				WithSkipDefaultMesh(true),
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithCPReplicas(2),
				WithHelmOpt("controlPlane.environment", "universal"),
				WithHelmOpt("controlPlane.envVars.KUMA_MULTIZONE_GLOBAL_KDS_TLS_ENABLED", "false"),
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
				WithGlobalAddress(global.GetKDSInsecureServerAddress()),
				WithHelmOpt("ingress.enabled", "true"),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(TestNamespace), democlient.WithMesh("default")),
				testserver.Install(),
			)).
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
		Eventually(func(g Gomega) {
			output, err := zoneCluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).To(ContainSubstring("default"))
		}, "5s", "500ms").Should(Succeed())

		// and dataplanes are synced to global
		Eventually(func(g Gomega) {
			output, err := globalCluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(output).To(ContainSubstring("demo-client"))
		}, "5s", "500ms").Should(Succeed())
	})

	It("communication in between apps in zone works", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(zoneCluster, "demo-client", "http://test-server_kuma-test_svc_80.mesh",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
