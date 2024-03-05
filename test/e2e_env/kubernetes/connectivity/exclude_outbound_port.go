package connectivity

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func ExcludeOutboundPort() {
	meshName := "exclude-outbound-port"
	namespace := "exclude-outbound-port"
	namespaceExternal := "exclude-outbound-port-external"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceExternal)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithNamespace(namespaceExternal),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespaceExternal)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to use network from init container if we ignore ports for uid", func() {
		Expect(kubernetes.Cluster.Install(testserver.Install(
			testserver.WithName("test-server"),
			testserver.WithNamespace(namespace),
			testserver.WithPodAnnotations(map[string]string{
				metadata.KumaInitFirst:                             "true",
				metadata.KumaTrafficExcludeOutboundTCPPortsForUIDs: "80:1234",
				metadata.KumaTrafficExcludeOutboundUDPPortsForUIDs: "53:1234",
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
			},
			)))).To(Succeed())
	})
}
