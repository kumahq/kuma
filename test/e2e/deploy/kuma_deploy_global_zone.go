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

func ZoneAndGlobal() {
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

	trafficRoutePolicy := func(namespace string, policyname string, weight int) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
mesh: default
metadata:
  namespace: %s
  name: %s
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'
  conf:
    loadBalancer:
      roundRobin: {}
    split:
      - weight: %d
        destination:
          kuma.io/service: '*'
`, namespace, policyname, weight)
	}

	var clusters Clusters
	var c1, c2 Cluster
	var global, zone ControlPlane
	var optsGlobal, optsZone = KumaK8sDeployOpts, KumaZoneK8sDeployOpts
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
		optsZone = append(optsZone,
			WithIngress(),
			WithGlobalAddress(global.GetKDSServerAddress()))

		err = NewClusterSetup().
			Install(Kuma(core.Zone, optsZone...)).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s("default")).
			Setup(c2)
		Expect(err).ToNot(HaveOccurred())

		zone = c2.GetKuma()
		Expect(zone).ToNot(BeNil())

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
		logs2, err := zone.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs2).To(ContainSubstring("\"mode\":\"zone\""))
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

		err = c2.DeleteKuma(optsZone...)
		Expect(err).ToNot(HaveOccurred())

		Expect(clusters.DismissCluster()).To(Succeed())
	})

	It("should deploy Zone and Global on 2 clusters", func() {
		// when check if zone is online
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

		// then zone is online
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

		policy_create := trafficRoutePolicy(KumaNamespace, "traffic-default", 100)
		policy_update := trafficRoutePolicy(KumaNamespace, "traffic-default", 101)

		// Deny policy CREATE on zone
		err := k8s.KubectlApplyFromStringE(c2.GetTesting(), c2.GetKubectlOptions(), policy_update)
		Expect(err.Error()).To(ContainSubstring("should be only applied on global"))

		// Accept policy CREATE on global
		err = k8s.KubectlApplyFromStringE(c1.GetTesting(), c1.GetKubectlOptions(), policy_create)
		Expect(err).ToNot(HaveOccurred())

		// TrafficRoute synced to zone
		Eventually(func() string {
			output, err := k8s.RunKubectlAndGetOutputE(c2.GetTesting(), c2.GetKubectlOptions("default"), "get", "TrafficRoute")
			Expect(err).ToNot(HaveOccurred())
			return output
		}, "5s", "500ms").Should(ContainSubstring("traffic-default"))

		// Deny policy UPDATE on zone
		err = k8s.KubectlApplyFromStringE(c2.GetTesting(), c2.GetKubectlOptions(), policy_update)
		Expect(err.Error()).To(ContainSubstring("should be only applied on global"))

		// Accept policy UPDATE on global
		err = k8s.KubectlApplyFromStringE(c1.GetTesting(), c1.GetKubectlOptions(), policy_update)
		Expect(err).ToNot(HaveOccurred())
	})
}
