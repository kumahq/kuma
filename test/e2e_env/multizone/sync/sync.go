package sync

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Sync() {
	namespace := "sync"
	meshName := "sync"

	BeforeAll(func() {
		Expect(env.Global.Install(MTLSMeshUniversal(meshName))).To(Succeed())
		Expect(WaitForMesh(meshName, env.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(meshName, namespace)).
			Setup(env.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("test-server", meshName)).
			Setup(env.UniZone1)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		//Expect(env.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		//Expect(env.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		//Expect(env.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should show zones as online", func() {
		Eventually(func(g Gomega) {
			out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(strings.Count(out, "Online")).To(Equal(4))
		}, "30s", "1s").Should(Succeed())
	})

	Context("from Remote to Global", func() {
		It("should sync Zone Ingress", func() {
			Eventually(func(g Gomega) {
				out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zone-ingresses")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(strings.Count(out, "Online")).To(Equal(4))
			}, "30s", "1s").Should(Succeed())
		})

		It("should sync Zone Egresses", func() {
			Eventually(func(g Gomega) {
				out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneegresses")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(strings.Count(out, "Online")).To(Equal(4))
			}, "30s", "1s").Should(Succeed())
		})

		It("should sync Dataplane with insight", func() {
			Eventually(func(g Gomega) {
				out, err := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplanes", "--mesh", meshName)
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
				return k8s.RunKubectlAndGetOutputE(env.KubeZone1.GetTesting(), env.KubeZone1.GetKubectlOptions(), "get", "trafficroute")
			}, "30s", "1s").Should(ContainSubstring(name))
			Eventually(func() (string, error) {
				return env.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "traffic-routes", "-m", meshName)
			}, "30s", "1s").Should(ContainSubstring(name))
		}

		It("should sync policy creation", func() {
			// given
			name := "tr-synced"

			// when
			Expect(env.Global.Install(YamlUniversal(universalPolicyNamed(name, 100)))).To(Succeed())

			// then
			policySyncedToZones(name)
		})

		It("should sync policy update", func() {
			// given
			name := "tr-update"
			Expect(env.Global.Install(YamlUniversal(universalPolicyNamed(name, 100)))).To(Succeed())
			policySyncedToZones(name)

			// when
			Expect(env.Global.Install(YamlUniversal(universalPolicyNamed(name, 101)))).To(Succeed())

			// then
			Eventually(func() (string, error) {
				return k8s.RunKubectlAndGetOutputE(env.KubeZone1.GetTesting(), env.KubeZone1.GetKubectlOptions(), "get", "trafficroute", name, "-oyaml")
			}, "30s", "1s").Should(ContainSubstring(`weight: 101`))
			Eventually(func() (string, error) {
				return env.UniZone1.GetKumactlOptions().RunKumactlAndGetOutput("get", "traffic-route", name, "-m", meshName, "-o", "yaml")
			}, "30s", "1s").Should(ContainSubstring(`weight: 101`))
		})

		Context("Deny", func() {
			name := "tr-denied"
			BeforeAll(func() {
				Expect(env.Global.Install(YamlUniversal(universalPolicyNamed(name, 100)))).To(Succeed())
				policySyncedToZones(name)
			})

			It("should deny creating policy on Kube Zone CP", func() {
				err := k8s.KubectlApplyFromStringE(env.KubeZone1.GetTesting(), env.KubeZone1.GetKubectlOptions(), kubernetesPolicyNamed("denied", 100))
				Expect(err).To(HaveOccurred())
			})

			It("should deny creating policy on Universal Zone CP", func() {
				err := env.UniZone1.GetKumactlOptions().KumactlApplyFromString(universalPolicyNamed("denied", 100))
				Expect(err).To(HaveOccurred())
			})

			It("should deny update on Kube Zone CP", func() {
				policyUpdate := kubernetesPolicyNamed(name, 101)
				err := k8s.KubectlApplyFromStringE(env.KubeZone1.GetTesting(), env.KubeZone1.GetKubectlOptions(), policyUpdate)
				Expect(err).To(HaveOccurred())
			})

			It("should deny update on Universal Zone CP", func() {
				policyUpdate := universalPolicyNamed(name, 101)
				err := env.UniZone1.GetKumactlOptions().KumactlApplyFromString(policyUpdate)
				Expect(err).To(HaveOccurred())
			})

			It("should deny delete on Kube Zone CP", func() {
				err := k8s.RunKubectlE(env.KubeZone1.GetTesting(), env.KubeZone1.GetKubectlOptions(), "delete", "trafficroute", name)
				Expect(err).To(HaveOccurred())
			})

			It("should deny delete on Universal Zone CP", func() {
				err := env.UniZone1.GetKumactlOptions().RunKumactl("delete", "traffic-route", name, "-m", meshName)
				Expect(err).To(HaveOccurred())
			})
		})
	})
}
