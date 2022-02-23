package deploy

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ZoneAndGlobal() {
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

	errRegex := `Operation not allowed\. .* resources like TrafficRoute can be updated or deleted only from the GLOBAL control plane and not from a ZONE control plane\.`

	var clusters Clusters
	var c1, c2 Cluster
	var global, zone ControlPlane
	var originalKumaNamespace = Config.KumaNamespace

	BeforeEach(func() {
		// set the new namespace
		Config.KumaNamespace = "other-kuma-system"
		var err error
		clusters, err = NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		c1 = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())

		global = c1.GetKuma()
		Expect(global).ToNot(BeNil())

		c2 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithIngress(),
				WithGlobalAddress(global.GetKDSServerAddress()),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default")).
			Setup(c2)
		Expect(err).ToNot(HaveOccurred())

		zone = c2.GetKuma()
		Expect(zone).ToNot(BeNil())

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
			Config.KumaNamespace = originalKumaNamespace
		}()

		err := c2.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = c1.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		err = c2.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		Expect(clusters.DismissCluster()).To(Succeed())
	})

	It("should deploy Zone and Global on 2 clusters", func() {
		// check if zone is online and backend is marked as kubernetes
		Eventually(func() (string, error) {
			return c1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
		}, "30s", "1s").Should(And(ContainSubstring("Online"), ContainSubstring("kubernetes")))

		// and dataplanes are synced to global
		Eventually(func() string {
			output, err := k8s.RunKubectlAndGetOutputE(c1.GetTesting(), c1.GetKubectlOptions("default"), "get", "dataplanes")
			Expect(err).ToNot(HaveOccurred())
			return output
		}, "5s", "500ms").Should(ContainSubstring("kuma-2-zone.demo-client"))

		policy_create := trafficRoutePolicy(Config.KumaNamespace, "traffic-default", 100)
		policy_update := trafficRoutePolicy(Config.KumaNamespace, "traffic-default", 101)

		// Deny policy CREATE on zone
		err := k8s.KubectlApplyFromStringE(c2.GetTesting(), c2.GetKubectlOptions(), policy_update)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(MatchRegexp(errRegex))

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
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(MatchRegexp(errRegex))

		// Deny policy DELETE on zone
		err = k8s.KubectlDeleteFromStringE(c2.GetTesting(), c2.GetKubectlOptions(), policy_create)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(MatchRegexp(errRegex))

		// Accept policy UPDATE on global
		err = k8s.KubectlApplyFromStringE(c1.GetTesting(), c1.GetKubectlOptions(), policy_update)
		Expect(err).ToNot(HaveOccurred())
	})
}
