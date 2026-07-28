package meshservice

import (
	"context"
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/envs/multizone"
)

func Connectivity() {
	namespace := "msconnectivity"
	clientNamespace := "msconnectivity-client"
	meshName := "msconnectivity"
	autoGenerateUniversalClusterName := "autogenerate-universal"

	env := multizone.NewMeshEnv(meshName)

	var autoGenerateUniversalCluster *UniversalCluster

	var testServerPodNames []string
	BeforeAll(func() {
		autoGenerateUniversalCluster = NewUniversalCluster(NewTestingT(), autoGenerateUniversalClusterName, Silent)

		Expect(env.
			WithGlobal(
				YamlUniversal(fmt.Sprintf(`
type: HostnameGenerator
name: kube-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.svc.{{ .Zone }}.mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: kubernetes
        kuma.io/mesh: %s
        k8s.kuma.io/is-headless-service: "false"
`, meshName)),
				YamlUniversal(`
type: HostnameGenerator
name: kube-not-my-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.svc.{{ .Zone }}.not-my-mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: kubernetes
        kuma.io/mesh: non-existent-mesh
        k8s.kuma.io/is-headless-service: "false"
`),
				YamlUniversal(fmt.Sprintf(`
type: HostnameGenerator
name: uni-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.svc.{{ .Zone }}.mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: universal
        kuma.io/mesh: "%s"
`, meshName)),
				YamlUniversal(`
type: HostnameGenerator
name: uni-not-my-mesh-specific-msconnectivity
spec:
  template: '{{ .DisplayName }}.svc.{{ .Zone }}.not-my-mesh-specific.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: universal
        kuma.io/mesh: non-existent-mesh
`),
			).
			WithZoneSetup(multizone.KubeZone1, Namespace(clientNamespace)).
			WithZone(multizone.KubeZone1,
				testserver.Install(
					testserver.WithName("demo-client"),
					testserver.WithNamespace(clientNamespace),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-1"),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithName("statefulset-test-server"),
					testserver.WithStatefulSet(),
					testserver.WithHeadlessService(),
					testserver.WithEchoArgs("echo", "--instance", "kube-statefulset-test-server-1"),
				),
				democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName)),
			).
			WithZone(multizone.KubeZone2,
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
				),
			).
			WithZone(multizone.UniZone1,
				DemoClientUniversal("uni-demo-client", meshName,
					WithTransparentProxy(true),
					WithWorkload("uni-demo-client"),
				),
				TestServerUniversal("test-server", meshName,
					WithArgs([]string{"echo", "--instance", "uni-test-server-1"}),
					WithWorkload("test-server"),
				),
			).
			WithZone(multizone.UniZone2,
				TestServerUniversal("test-server", meshName,
					WithArgs([]string{"echo", "--instance", "uni-test-server"}),
					WithWorkload("test-server"),
				),
			).
			WithZoneSetup(autoGenerateUniversalCluster, Kuma(
				core.Zone,
				WithGlobalAddress(multizone.Global.GetKuma().GetKDSServerAddress()),
				WithEnv("KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY", "0s"), // we have only 1 Kuma CP instance so there is no risk setting this to 0
				WithEnv("KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF", "1s"),
			)).
			WithZone(autoGenerateUniversalCluster,
				TestServerUniversal("test-server", meshName,
					WithArgs([]string{"echo", "--instance", "auto-uni-test-server"}),
					WithWorkload("test-server"),
				),
			).
			Setup()).To(Succeed())

		Expect(multizone.KubeZone1.WaitApp("statefulset-test-server", namespace, 1)).To(Succeed())
		for _, pod := range k8s.ListPodsContext(multizone.KubeZone1.GetTesting(), context.Background(),
			multizone.KubeZone1.GetKubectlOptions(namespace),
			kube_meta.ListOptions{
				LabelSelector: "app=statefulset-test-server",
			},
		) {
			testServerPodNames = append(testServerPodNames, pod.Name)
		}
		Expect(testServerPodNames).To(HaveLen(1))
	})

	AfterEachFailure(func() {
		env.Debug()
	})

	E2EAfterAll(func() {
		Expect(env.Cleanup()).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(clientNamespace)).To(Succeed())
		Expect(autoGenerateUniversalCluster.DismissCluster()).To(Succeed())
	})

	type testCase struct {
		address          func() string
		expectedInstance string
		should           types.GomegaMatcher
	}

	DescribeTable("client from Kubernetes",
		func(given testCase) {
			if given.should == nil {
				given.should = Succeed()
			}
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(multizone.KubeZone1, "demo-client", given.address(),
					client.FromKubernetesPod(namespace, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal(given.expectedInstance))
			}, "30s", "1s").Should(given.should)
		},
		Entry("should access service in the same Kubernetes cluster", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.cluster.local:80" },
			expectedInstance: "kube-test-server-1",
		}),
		Entry("should access service in the same Kubernetes cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.kuma-1.mesh-specific.mesh.local:80" },
			expectedInstance: "kube-test-server-1",
		}),
		Entry("should access headless service in the same Kubernetes cluster", testCase{
			address: func() string {
				return fmt.Sprintf("http://%s.statefulset-test-server.msconnectivity.svc.cluster.local:80", testServerPodNames[0])
			},
			expectedInstance: "kube-statefulset-test-server-1",
		}),
		Entry("should access service in another Kubernetes cluster", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.kuma-2.mesh.local:80" },
			expectedInstance: "kube-test-server-2",
		}),
		Entry("should access service in a Universal cluster", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should access service in a Universal cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should not be able to access service in a Universal cluster if the hostname generator for that name doesn't match", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.not-my-mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server",
			should:           Not(Succeed()),
		}),
		Entry("should access service in a Universal cluster where MeshService is autogenerated", testCase{
			address:          func() string { return "http://test-server.svc.autogenerate-universal.mesh.local:80" },
			expectedInstance: "auto-uni-test-server",
		}),
	)

	DescribeTable("client from Universal",
		func(given testCase) {
			if given.should == nil {
				given.should = Succeed()
			}
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(multizone.UniZone1, "uni-demo-client", given.address())
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal(given.expectedInstance))
			}, "30s", "1s").Should(given.should)
		},
		Entry("should access the headless service in another Kubernetes cluster 1", testCase{
			address: func() string {
				return fmt.Sprintf("http://%s.statefulset-test-server.msconnectivity.svc.kuma-1.mesh.local:80", testServerPodNames[0])
			},
			expectedInstance: "kube-statefulset-test-server-1",
		}),
		Entry("should access service in another Kubernetes cluster 2", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.kuma-2.mesh.local:80" },
			expectedInstance: "kube-test-server-2",
		}),
		Entry("should access service in a Kubernetes cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.msconnectivity.svc.kuma-2.mesh-specific.mesh.local:80" },
			expectedInstance: "kube-test-server-2",
		}),
		Entry("should not be able to access service in a Kubernetes cluster if the hostname generator for that name doesn't match", testCase{
			address: func() string {
				return "http://test-server.msconnectivity.svc.kuma-2.not-my-mesh-specific.mesh.local:80"
			},
			expectedInstance: "kube-test-server-2",
			should:           Not(Succeed()),
		}),
		Entry("should access service in the same Universal cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.svc.kuma-4.mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server-1",
		}),
		Entry("should access service in another Universal cluster via a mesh-targeted generator name", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh-specific.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should access service in another Universal cluster", testCase{
			address:          func() string { return "http://test-server.svc.kuma-5.mesh.local:80" },
			expectedInstance: "uni-test-server",
		}),
		Entry("should access service in another Universal cluster where MeshService is autogenerated", testCase{
			address:          func() string { return "http://test-server.svc.autogenerate-universal.mesh.local:80" },
			expectedInstance: "auto-uni-test-server",
		}),
	)
}
