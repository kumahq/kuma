package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"

	api_server "github.com/kumahq/kuma/pkg/api-server"

	"github.com/kumahq/kuma/pkg/config/core"

	"github.com/go-errors/errors"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Remote and Global with Helm chart", func() {

	namespaceWithSidecarInjection := func(namespace string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  annotations:
    kuma.io/sidecar-injection: "enabled"
`, namespace)
	}

	var clusters Clusters
	var c1, c2 Cluster
	var global, remote ControlPlane
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		var err error
		clusters, err = NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		c1 = clusters.GetCluster(Kuma1)
		c2 = clusters.GetCluster(Kuma2)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)
		deployOptsFuncs = []DeployOptionsFunc{
			WithInstallationMode(HelmInstallationMode),
			WithHelmReleaseName(releaseName),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Global, deployOptsFuncs...)).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())

		global = c1.GetKuma()
		Expect(global).ToNot(BeNil())

		deployOptsFuncs = append(deployOptsFuncs, WithGlobalAddress(global.GetKDSServerAddress()))
		err = NewClusterSetup().
			Install(Kuma(core.Remote, deployOptsFuncs...)).
			Install(KumaDNS()).
			Install(Ingress(nil)).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s()).
			Install(EchoServerK8s()).
			Setup(c2)
		Expect(err).ToNot(HaveOccurred())

		remote = c2.GetKuma()
		Expect(remote).ToNot(BeNil())

		// when
		err = c1.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c2.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		err = k8s.KubectlApplyFromStringE(c1.GetTesting(), c1.GetKubectlOptions(),
			fmt.Sprintf(ZoneTemplateK8s,
				remote.GetName(),
				remote.GetIngressAddress()))
		Expect(err).ToNot(HaveOccurred())

		// then
		logs1, err := global.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs1).To(ContainSubstring("\"mode\":\"global\""))

		// and
		logs2, err := remote.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs2).To(ContainSubstring("\"mode\":\"remote\""))

	})

	AfterEach(func() {
		// tear down apps
		Expect(c2.DeleteNamespace(TestNamespace)).To(Succeed())
		// tear down Kuma
		Expect(clusters.DeleteKuma(deployOptsFuncs...)).To(Succeed())
	})

	It("Should deploy Remote and Global on 2 clusters", func() {
		clustersStatus := api_server.Zones{}
		Eventually(func() (bool, error) {
			status, response := http_helper.HttpGet(c1.GetTesting(), global.GetGlobaStatusAPI(), nil)
			if status != http.StatusOK {
				return false, errors.Errorf("unable to contact server %s with status %d", global.GetGlobaStatusAPI(), status)
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
		found := false
		for _, cluster := range clustersStatus {
			if cluster.Address == remote.GetIngressAddress() {
				Expect(cluster.Active).To(BeTrue())
				found = true
				break
			}
		}
		Expect(found).To(BeTrue())
	})

	It("should deploy Remote and Global on 2 clusters and sync dataplanes", func() {
		// given
		namespace := func(namespace string) string {
			return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, namespace)
		}
		dp := func(cluster, namespace, name string) string {
			return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Dataplane
mesh: default
metadata:
  name: %s
  namespace: %s
spec:
  networking:
    address: 192.168.0.1
    inbound:
      - port: 12343
        tags:
          kuma.io/service: backend
          kuma.io/zone: %s
    outbound:
      - port: 1212
        tags:
          kuma.io/service: web
`, name, namespace, cluster)
		}

		// when
		err := YamlK8s(namespace("custom-ns"))(c2)
		Expect(err).ToNot(HaveOccurred())
		err = YamlK8s(dp("kuma-2-remote", "custom-ns", "dp-1"))(c2)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() string {
			output, err := k8s.RunKubectlAndGetOutputE(c1.GetTesting(), c1.GetKubectlOptions("default"), "get", "dataplanes")
			Expect(err).ToNot(HaveOccurred())
			return output
		}, "5s", "500ms").Should(ContainSubstring("kuma-2-remote.dp-1.custom-ns"))
	})
})
