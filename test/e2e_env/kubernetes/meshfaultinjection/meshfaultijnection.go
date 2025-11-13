package meshfaultinjection

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/envs/kubernetes"
)

func MeshFaultInjection() {
	mesh := "mfi-rules"
	namespace := "mfi-rules-ns"
	testServerURL := fmt.Sprintf("test-server.%s.svc:80", namespace)

	meshIdentity := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshIdentity
metadata:
  name: identity-mfi
  namespace: %s  
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
`, Config.KumaNamespace, mesh)

	mtp := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp-mfi
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  rules:
    - default:
        allow:
          - spiffeID:
              type: Prefix
              value: spiffe://%s.default.mesh.local`, namespace, mesh, mesh)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(Yaml(builders.Mesh().
				WithName(mesh).
				WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive))).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(YamlK8s(mtp)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(mesh)),
				testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(namespace)),
			)).
			Install(YamlK8s(meshIdentity)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, mesh, v1alpha1.MeshFaultInjectionResourceTypeDescriptor)).To(Succeed())
	})

	DescribeTable("should fail based on spiffeID", func(mfiConfig string) {
		// traffic should work normally
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", testServerURL,
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(ContainSubstring("test-server"))
		}, "30s", "1s").Should(Succeed())

		// when
		Expect(YamlK8s(mfiConfig)(kubernetes.Cluster)).To(Succeed())

		// should fail on fault injection
		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", testServerURL,
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.ResponseCode).To(Equal(501))
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())
	},
		Entry("exact match on spiffeID", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshFaultInjection
metadata:
  name: mfi-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  rules:
    - matches:
        - spiffeID:
            type: Exact
            value: spiffe://%s.default.mesh.local/ns/%s/sa/default
      default:
        http:
          - abort: 
              httpStatus: 501
              percentage: 100
`, namespace, mesh, mesh, namespace)),
		Entry("prefix match on spiffeID", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshFaultInjection
metadata:
  name: mfi-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  rules:
    - matches:
        - spiffeID:
            type: Prefix
            value: spiffe://%s.default.mesh.local
      default:
        http:
          - abort: 
              httpStatus: 501
              percentage: 100
`, namespace, mesh, mesh)))
}
