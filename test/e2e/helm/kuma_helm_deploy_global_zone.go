package helm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ZoneAndGlobalWithHelmChart() {
	var clusters Clusters
	var c1, c2 Cluster
	var global, zone ControlPlane

	BeforeAll(func() {
		var err error
		clusters, err = NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		c1 = clusters.GetCluster(Kuma1).
			WithTimeout(6 * time.Second).
			WithRetries(60)
		c2 = clusters.GetCluster(Kuma2).
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		err = NewClusterSetup().
			Install(Kuma(core.Global,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
			)).
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
		Expect(clusters.DismissCluster()).To(Succeed())
	})

	It("Should deploy Zone and Global on 2 clusters", func() {
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
		}, time.Minute, DefaultTimeout).Should(BeTrue())

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
			output, err := k8s.RunKubectlAndGetOutputE(c1.GetTesting(), c1.GetKubectlOptions("default"), "get", "dataplanes")
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
