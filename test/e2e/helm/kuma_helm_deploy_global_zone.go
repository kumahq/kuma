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
	"github.com/kumahq/kuma/pkg/intercp/catalog"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ZoneAndGlobalWithHelmChart() {
	var c1, c2 Cluster
	var global, zone ControlPlane

	BeforeAll(func() {
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

		err := NewClusterSetup().
			Install(Kuma(core.Global,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithCPReplicas(2),
				WithHelmOpt("controlPlane.config", `
interCp:
  catalog:
    heartbeatInterval: 1s
    writerInterval: 3s
`),
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
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(TestNamespace), democlient.WithMesh("default")),
				testserver.Install(),
			)).
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
			output, err := k8s.RunKubectlAndGetOutputE(c1.GetTesting(), c1.GetKubectlOptions(Config.KumaNamespace), "get", "dataplanes")
			Expect(err).ToNot(HaveOccurred())
			return output
		}, "5s", "500ms").Should(ContainSubstring("demo-client"))
	})

	It("communication in between apps in zone works", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(c2, "demo-client", "http://test-server_kuma-test_svc_80.mesh",
				client.FromKubernetesPod(TestNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})

	Context("Intercommunication CP server catalog on Global CP", func() {
		fetchInstances := func() (map[string]struct{}, error) {
			out, err := k8s.RunKubectlAndGetOutputE(c1.GetTesting(), c1.GetKubectlOptions(Config.KumaNamespace), "get", "configmap", "cp-catalog", "-o", "jsonpath={.data.config}")
			if err != nil {
				return nil, err
			}
			instances := catalog.ConfigInstances{}
			if err := json.Unmarshal([]byte(out), &instances); err != nil {
				return nil, err
			}
			m := map[string]struct{}{}
			for _, instance := range instances.Instances {
				m[instance.Id] = struct{}{}
			}
			return m, nil
		}

		It("should update instances in catalog when we scale CP", func() {
			// given
			var instances map[string]struct{}
			Eventually(func(g Gomega) {
				ins, err := fetchInstances()
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(ins).To(HaveLen(2))
				instances = ins
			}, "30s", "1s").Should(Succeed())

			// when
			_, err := k8s.RunKubectlAndGetOutputE(c1.GetTesting(), c1.GetKubectlOptions(Config.KumaNamespace), "rollout", "restart", "deployment", Config.KumaServiceName)

			// then
			Expect(err).ToNot(HaveOccurred())
			Eventually(func(g Gomega) {
				ins, err := fetchInstances()
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(ins).To(HaveLen(2))
				for instanceID := range instances { // there are no old instances
					g.Expect(ins).ToNot(ContainElement(instanceID))
				}
			}, "30s", "1s").Should(Succeed())
		})
	})

	It("should execute admin operations on Global CP", func() {
		// given DP available on Global CP
		Eventually(func(g Gomega) {
			dataplanes, err := c1.GetKumactlOptions().KumactlList("dataplanes", "default")
			g.Expect(err).ToNot(HaveOccurred())
			// Dataplane names are generated, so we check for a partial match.
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("demo-client")))
			for _, dpName := range dataplanes {
				if strings.Contains(dpName, "demo-client") {
					_, err = c1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", dpName, "--type", "config-dump")
					Expect(err).ToNot(HaveOccurred())
				}
			}
		}, "30s", "250ms").Should(Succeed())
	})
}
