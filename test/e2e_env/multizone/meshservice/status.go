package meshservice

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/kds/hash"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func MeshService() {
	meshName := "meshservice"
	namespace := "meshservice"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		Expect(NewClusterSetup().
			Install(YamlUniversal(`
type: MeshService
name: backend
mesh: meshservice
labels:
  kuma.io/origin: zone
spec:
  selector:
    dataplaneTags:
      kuma.io/service: test-server
  ports:
  - port: 80
    targetPort: 80
    protocol: http
`)).
			Install(TestServerUniversal("dp-echo-1", meshName,
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceVersion("v1"),
			)).
			Setup(multizone.UniZone1)).To(Succeed())

		Expect(NewClusterSetup().
			Install(DemoClientUniversal("uni-demo-client", meshName, WithTransparentProxy(true))).
			Setup(multizone.UniZone2)).To(Succeed())

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugUniversal(multizone.UniZone2, meshName)
		DebugKube(multizone.KubeZone1, meshName)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	msStatus := func(cluster Cluster, name string) (*v1alpha1.MeshService, *v1alpha1.MeshServiceStatus, error) {
		out, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshservice", "-m", meshName, name, "-ojson")
		if err != nil {
			return nil, nil, err
		}
		res, err := rest.JSON.Unmarshal([]byte(out), v1alpha1.MeshServiceResourceTypeDescriptor)
		if err != nil {
			return nil, nil, err
		}
		return res.GetSpec().(*v1alpha1.MeshService), res.GetStatus().(*v1alpha1.MeshServiceStatus), nil
	}

	It("should sync MeshService to global with VIP status", func() {
		// VIP and identities are assigned
		Eventually(func(g Gomega) {
			spec, status, err := msStatus(multizone.UniZone1, "backend")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.VIPs).To(HaveLen(1))
			g.Expect(status.VIPs[0].IP).To(ContainSubstring("241.0.0."))
			g.Expect(spec.Identities).To(Equal([]v1alpha1.MeshServiceIdentity{
				{
					Type:  v1alpha1.MeshServiceIdentityServiceTagType,
					Value: "test-server",
				},
			}))
		}, "30s", "1s").Should(Succeed())

		// and MeshService is synced to global with the original status
		Eventually(func(g Gomega) {
			spec, status, err := msStatus(multizone.Global, hash.HashedName(meshName, "backend", "kuma-4"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.VIPs).To(HaveLen(1))
			g.Expect(status.VIPs[0].IP).To(ContainSubstring("241.0.0."))
			g.Expect(spec.Identities).To(Equal([]v1alpha1.MeshServiceIdentity{
				{
					Type:  v1alpha1.MeshServiceIdentityServiceTagType,
					Value: "test-server",
				},
			}))
		}, "30s", "1s").Should(Succeed())

		// and MeshService is synced to other zone but VIP is generated by other zone
		Eventually(func(g Gomega) {
			spec, status, err := msStatus(multizone.UniZone2, hash.HashedName(meshName, "backend"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.VIPs).To(HaveLen(1))
			g.Expect(status.VIPs[0].IP).To(ContainSubstring("251.0.0."))
			g.Expect(spec.Identities).To(Equal([]v1alpha1.MeshServiceIdentity{
				{
					Type:  v1alpha1.MeshServiceIdentityServiceTagType,
					Value: "test-server",
				},
			}))
		}, "30s", "1s").Should(Succeed())
	})

	It("should connect cross-zone using MeshService from universal cluster", func() {
		err := multizone.UniZone2.Install(YamlUniversal(`
type: HostnameGenerator
mesh: meshservice
name: basic-uni
labels:
  kuma.io/origin: zone
spec:
  template: '{{ label "kuma.io/display-name" }}.{{ label "kuma.io/zone" }}.ms.mesh'
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: backend
        kuma.io/origin: global
`))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(multizone.UniZone2, "uni-demo-client", "backend.kuma-4.ms.mesh:80")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("echo-v1"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should connect cross-zone using MeshService from kubernetes cluster", func() {
		err := multizone.KubeZone2.Install(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: HostnameGenerator
metadata:
  name: basic-kube
  labels:
    kuma.io/mesh: meshservice
    kuma.io/origin: zone
spec:
  template: '{{ label "kuma.io/display-name" }}.{{ label "kuma.io/zone" }}.ms.mesh'
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: backend
        kuma.io/origin: global
`))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(multizone.KubeZone2, "demo-client", "backend.kuma-4.ms.mesh:80",
				client.FromKubernetesPod(meshName, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("echo-v1"))
		}, "30s", "1s").Should(Succeed())
	})
}
