package cni

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func ExcludeOutboundPort() {
	meshName := "exclude-outbound-port"

	namespace := "exclude-outbound-port"
	namespaceExternal := "exclude-outbound-port-external"

	var cluster Cluster
	var k8sCluster *K8sCluster

	BeforeAll(func() {
		k8sCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		cluster = k8sCluster.
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName := fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueId()))

		Expect(NewClusterSetup().
			Install(Kuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithSkipDefaultMesh(true), // it's common case for HELM deployments that Mesh is also managed by HELM therefore it's not created by default
				WithHelmOpt("cni.logLevel", "debug"),
				WithCNI(),
			)).
			Install(MTLSMeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceExternal)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespaceExternal),
			)).
			Setup(cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(k8sCluster, meshName, namespace, namespaceExternal)
	})

	E2EAfterAll(func() {
		Expect(cluster.DeleteNamespace(namespace)).To(Succeed())
		Expect(cluster.DeleteNamespace(namespaceExternal)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should be able to use network from init container if we ignore ports for uid", func() {
		Expect(cluster.Install(testserver.Install(
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
		))).To(Succeed())
	})
}
