package cni

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
)

func ExcludeOutboundPort() {
	meshName := "exclude-outbound-port"

	namespace := "exclude-outbound-port"
	namespaceExternal := "exclude-outbound-port-external"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())

		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceExternal)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespaceExternal),
			)).
			Setup(multizone.KubeZone2)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(multizone.KubeZone2, meshName, namespace, namespaceExternal)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespaceExternal)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to use network from init container if we ignore ports for uid", func() {
		Expect(NewClusterSetup().Install(testserver.Install(
			testserver.WithName("test-server"),
			testserver.WithMesh(meshName),
			testserver.WithNamespace(namespace),
			testserver.WithPodAnnotations(map[string]string{
				metadata.KumaTrafficExcludeOutboundPortsForUIDs: "tcp:80:1234;udp:53:1234",
			}),
			testserver.AddInitContainer(corev1.Container{
				Name:            "init-test-server",
				Image:           Config.GetUniversalImage(),
				ImagePullPolicy: "IfNotPresent",
				Command:         []string{"curl"},
				Args:            []string{"-v", "-m", "3", "--fail", "test-server.exclude-outbound-port-external.svc.cluster.local:80"},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu":    resource.MustParse("50m"),
						"memory": resource.MustParse("64Mi"),
					},
				},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: pointer.To(int64(1234)),
				},
			}),
		)).Setup(multizone.KubeZone2)).To(Succeed())
	})
}
