package deploy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	api_server "github.com/kumahq/kuma/pkg/api-server"

	"github.com/kumahq/kuma/pkg/config/core"

	"github.com/go-errors/errors"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

func RemoteAndGLobal() {
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
	var optsGlobal, optsRemote = KumaK8sDeployOpts, KumaRemoteK8sDeployOpts
	var originalKumaNamespace = KumaNamespace

	BeforeEach(func() {
		// set the new namespace
		KumaNamespace = "other-kuma-system"
		var err error
		clusters, err = NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		c1 = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())

		global = c1.GetKuma()
		Expect(global).ToNot(BeNil())

		c2 = clusters.GetCluster(Kuma2)
		optsRemote = append(optsRemote,
			WithIngress(),
			WithGlobalAddress(global.GetKDSServerAddress()))

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote...)).
			Install(KumaDNS()).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s("default")).
			Install(EchoServerK8s("default")).
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
		if ShouldSkipCleanup() {
			return
		}

		defer func() {
			// restore the original namespace
			KumaNamespace = originalKumaNamespace
		}()

		err := c2.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = c1.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())

		err = c2.DeleteKuma(optsRemote...)
		Expect(err).ToNot(HaveOccurred())

		Expect(clusters.DismissCluster()).To(Succeed())
	})

	It("should deploy Remote and Global on 2 clusters", func() {
		// when check if remote is online
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

		// then remote is online
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
		}, "5s", "500ms").Should(ContainSubstring("kuma-2-remote.demo-client"))
	})
}
