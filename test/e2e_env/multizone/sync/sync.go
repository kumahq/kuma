package sync

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/kds/hash"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/api"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Sync() {
	namespace := "sync"
	meshName := "sync"

	BeforeAll(func() {
		Expect(multizone.Global.Install(MTLSMeshUniversal(meshName))).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		group.Go(func() error {
			err := NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
				Setup(multizone.KubeZone1)
			return errors.Wrap(err, multizone.KubeZone1.Name())
		})

		group.Go(func() error {
			err := NewClusterSetup().
				Install(TestServerUniversal("test-server", meshName)).
				Setup(multizone.UniZone1)
			return errors.Wrap(err, multizone.UniZone1.Name())
		})
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should show zones as online", func() {
		Eventually(func(g Gomega) {
			out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
			g.Expect(err).ToNot(HaveOccurred())
			// Some tests create their own Zones that may or may not
			// be run simultaneously
			g.Expect(strings.Count(out, "Online")).To(BeNumerically(">=", 4))
		}, "30s", "1s").Should(Succeed())
	})

	It("should have insights in global and in zone", func() {
		// Ensure each side of KDS has the respective values for Global and Zone instance info
		Eventually(func(g Gomega) {
			result := &system.ZoneInsightResource{}
			api.FetchResource(g, multizone.Global, result, "", multizone.KubeZone1.ZoneName())
			g.Expect(result.Spec.Subscriptions).ToNot(BeEmpty())
			globalSub := result.Spec.Subscriptions[0]
			g.Expect(globalSub.GlobalInstanceId).ToNot(BeEmpty())
			g.Expect(globalSub.ZoneInstanceId).ToNot(BeEmpty())

			zoneResult := &system.ZoneInsightResource{}
			api.FetchResource(g, multizone.KubeZone1, zoneResult, "", multizone.KubeZone1.ZoneName())
			g.Expect(zoneResult.Spec.Subscriptions).ToNot(BeEmpty())
			zoneSub := zoneResult.Spec.Subscriptions[0]
			// Check that this is the other side of the connection
			g.Expect(zoneSub.GlobalInstanceId).To(Equal(globalSub.GlobalInstanceId))
			g.Expect(zoneSub.ZoneInstanceId).To(Equal(globalSub.ZoneInstanceId))
		}, "1m", "1s").Should(Succeed())
	})

	Context("from Remote to Global", func() {
		It("should sync Zone Ingress", func() {
			Eventually(func(g Gomega) {
				out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zone-ingresses")
				g.Expect(err).ToNot(HaveOccurred())
				// Some tests create their own ZoneIngresses that may or may not
				// be run simultaneously
				g.Expect(strings.Count(out, "Online")).To(BeNumerically(">=", 4))
			}, "30s", "1s").Should(Succeed())
		})

		It("should sync Zone Egresses", func() {
			Eventually(func(g Gomega) {
				out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneegresses")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(strings.Count(out, "Online")).To(Equal(4))
			}, "30s", "1s").Should(Succeed())
		})

		It("should sync Dataplane with insight", func() {
			Eventually(func(g Gomega) {
				out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", meshName)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(strings.Count(out, "Online")).To(Equal(2))
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("from Global to Zone", func() {
		universalPolicyNamed := func(name string, weight int) string {
			return fmt.Sprintf(`
type: TrafficRoute
mesh: sync
name: %s
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
conf:
  split:
    - weight: %d
      destination:
        kuma.io/service: '*'`, name, weight)
		}

		kubernetesPolicyNamed := func(name string, weight int) string {
			return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
mesh: default
metadata:
  name: %s
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'
  conf:
    split:
      - weight: %d
        destination:
          kuma.io/service: '*'`, name, weight)
		}

		policySyncedToZones := func(name string) {
			Eventually(func() (string, error) {
				return k8s.RunKubectlAndGetOutputE(multizone.KubeZone1.GetTesting(), multizone.KubeZone1.GetKubectlOptions(), "get", "trafficroute")
			}, "30s", "1s").Should(ContainSubstring(name))
			Eventually(func() (string, error) {
				return multizone.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "traffic-routes", "-m", meshName)
			}, "30s", "1s").Should(ContainSubstring(name))
		}

		It("should sync policy creation", func() {
			// given
			name := "tr-synced"

			// when
			Expect(multizone.Global.Install(YamlUniversal(universalPolicyNamed(name, 100)))).To(Succeed())

			// then
			policySyncedToZones(name)
		})

		It("should sync policy update", func() {
			// given
			name := "tr-update"
			Expect(multizone.Global.Install(YamlUniversal(universalPolicyNamed(name, 100)))).To(Succeed())
			policySyncedToZones(name)

			// when
			Expect(multizone.Global.Install(YamlUniversal(universalPolicyNamed(name, 101)))).To(Succeed())

			// then
			hashedName := hash.HashedName(meshName, "tr-update")
			Eventually(func() (string, error) {
				return k8s.RunKubectlAndGetOutputE(multizone.KubeZone1.GetTesting(), multizone.KubeZone1.GetKubectlOptions(), "get", "trafficroute", hashedName, "-oyaml")
			}, "30s", "1s").Should(ContainSubstring(`weight: 101`))
			Eventually(func() (string, error) {
				return multizone.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "traffic-route", hashedName, "-m", meshName, "-o", "yaml")
			}, "30s", "1s").Should(ContainSubstring(`weight: 101`))
		})

		Context("Deny", func() {
			name := "tr-denied"
			BeforeAll(func() {
				Expect(multizone.Global.Install(YamlUniversal(universalPolicyNamed(name, 100)))).To(Succeed())
				policySyncedToZones(name)
			})

			It("should deny creating policy on Kube Zone CP", func() {
				err := k8s.KubectlApplyFromStringE(multizone.KubeZone1.GetTesting(), multizone.KubeZone1.GetKubectlOptions(), kubernetesPolicyNamed("denied", 100))
				Expect(err).To(HaveOccurred())
			})

			It("should deny creating policy on Universal Zone CP", func() {
				err := multizone.UniZone1.GetKumactlOptions().KumactlApplyFromString(universalPolicyNamed("denied", 100))
				Expect(err).To(HaveOccurred())
			})

			It("should deny update on Kube Zone CP", func() {
				policyUpdate := kubernetesPolicyNamed(name, 101)
				err := k8s.KubectlApplyFromStringE(multizone.KubeZone1.GetTesting(), multizone.KubeZone1.GetKubectlOptions(), policyUpdate)
				Expect(err).To(HaveOccurred())
			})

			It("should deny update on Universal Zone CP", func() {
				policyUpdate := universalPolicyNamed(name, 101)
				err := multizone.UniZone1.GetKumactlOptions().KumactlApplyFromString(policyUpdate)
				Expect(err).To(HaveOccurred())
			})

			It("should deny delete on Kube Zone CP", func() {
				err := k8s.RunKubectlE(multizone.KubeZone1.GetTesting(), multizone.KubeZone1.GetKubectlOptions(), "delete", "trafficroute", name)
				Expect(err).To(HaveOccurred())
			})

			It("should deny delete on Universal Zone CP", func() {
				err := multizone.UniZone1.GetKumactlOptions().RunKumactl("delete", "traffic-route", name, "-m", meshName)
				Expect(err).To(HaveOccurred())
			})
		})

		It("should sync policy with a long name and store it as display name", func() {
			// given
			name := ""
			for i := 0; i < 253; i++ {
				name += "x"
			}

			// when
			Expect(multizone.Global.Install(YamlUniversal(universalPolicyNamed(name, 100)))).To(Succeed())

			// then
			hashedName := hash.HashedName(meshName, name)
			for _, cluster := range multizone.Zones() {
				Eventually(func() (string, error) {
					return cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "traffic-route", hashedName, "-m", meshName, "-o", "yaml")
				}, "30s", "1s").Should(ContainSubstring(`kuma.io/display-name: ` + name))
			}
		})
	})
}
