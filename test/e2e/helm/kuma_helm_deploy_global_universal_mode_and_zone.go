package helm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ZoneAndGlobalInUniversalModeWithHelmChart() {
	var c1, c2 Cluster
	var global, zone ControlPlane

	BeforeAll(func() {
		var err error
		c1 = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		c2 = NewK8sCluster(NewTestingT(), Kuma2, Silent).
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
				// we deploy postgres in Config.KumaNamespace so it's deleted with CP teardown
				// otherwise this fails because there is no namespace
				postgres.WithSkipNamespaceCleanup(true),
			)).
			// Install(WaitService(Config.KumaNamespace, "postgres-release-postgresql")). // this does not seem to work
			// control plane crashes twice waiting for postgres to come up
			Install(YamlK8s(`
apiVersion: v1
kind: Secret
metadata:
  name: postgres
  namespace: kuma-system
type: Opaque
stringData:
  password: "mesh"
`)).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Global,
				WithSkipDefaultMesh(true),
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithCPReplicas(2),
				WithHelmOpt("controlPlane.environment", "universal"),
				WithHelmOpt("postgres.host", "postgres-release-postgresql"),
				WithHelmOpt("postgres.port", "5432"),
				WithHelmOpt("postgres.user", "mesh"),
				WithHelmOpt("postgres.db", "mesh"),
				WithHelmOpt("controlPlane.secrets[0].Secret", "postgres"),
				WithHelmOpt("controlPlane.secrets[0].Key", "password"),
				WithHelmOpt("controlPlane.secrets[0].Env", "KUMA_STORE_POSTGRES_PASSWORD"),
			)).
			Install(MeshUniversal("default")).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())

		global = c1.GetKuma()
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
			Setup(c2)
		Expect(err).ToNot(HaveOccurred())

		zone = c2.GetKuma()
		Expect(zone).ToNot(BeNil())
	})

	E2EAfterAll(func() {
		Expect(c2.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(c1.DeleteKuma()).To(Succeed())
		Expect(c2.DeleteKuma()).To(Succeed())
		Expect(c1.DismissCluster()).To(Succeed())
		Expect(c2.DismissCluster()).To(Succeed())
	})

	It("should deploy Zone and Global on 2 clusters", func() {
		clustersStatus := api_server.Zones{}
		Eventually(func() (bool, error) {
			status, response := http_helper.HttpGet(c1.GetTesting(), global.GetGlobalStatusAPI(), nil)
			if status != http.StatusOK {
				return false, errors.Errorf("unable to contact server %s with status %d", global.GetGlobalStatusAPI(), status)
			}
			err := json.Unmarshal([]byte(response), &clustersStatus)
			if err != nil {
				return false, errors.Errorf("unable to parse response [%s] with error: %v", response, err)
			}
			if len(clustersStatus) != 1 {
				return false, nil
			}
			return clustersStatus[0].Active, nil
		}, "1m", "1s").Should(BeTrue())

		// then
		active := true
		for _, cluster := range clustersStatus {
			if !cluster.Active {
				active = false
			}
		}
		Expect(active).To(BeTrue())

		// and dataplanes are synced to global
		Eventually(func() string {
			output, err := c1.GetKumactlOptions().RunKumactlAndGetOutput("get", "dataplanes")
			Expect(err).ToNot(HaveOccurred())
			return output
		}, "5s", "500ms").Should(ContainSubstring("kuma-2-zone.demo-client"))
	})

	It("communication in between apps in zone works", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectResponse(c2, "demo-client", "http://test-server_kuma-test_svc_80.mesh",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
