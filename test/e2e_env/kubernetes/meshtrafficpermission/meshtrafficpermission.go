package meshtrafficpermission

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshTrafficPermissionRules() {
	mesh := "mtp-rules"
	namespace := "mtp-rules-ns"
	testServerURL := fmt.Sprintf("test-server.%s.svc:80", namespace)

	meshIdentity := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshIdentity
metadata:
  name: identity
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

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(Yaml(builders.Mesh().
				WithName(mesh).
				WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive))).
			Install(NamespaceWithSidecarInjection(namespace)).
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
		Expect(DeleteMeshResources(kubernetes.Cluster, mesh, v1alpha1.MeshTrafficPermissionResourceTypeDescriptor)).To(Succeed())
	})

	DescribeTable("should allow traffic based on spiffe id", func(mtpConfig string) {
		// should fail on rbac error since we don't have MTP installed
		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", testServerURL,
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.ResponseCode).To(Equal(403))
		}, "30s", "1s").Should(Succeed())

		// when
		Expect(YamlK8s(mtpConfig)(kubernetes.Cluster)).To(Succeed())

		// should allow traffic`
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", testServerURL,
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(ContainSubstring("test-server"))
		}, "30s", "1s").Should(Succeed())
	},
		Entry("exact match on spiffeId", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  rules:
    - default:
        allow:
          - spiffeId:
              type: Exact
              value: spiffe://%s.default.mesh.local/ns/%s/sa/default
`, namespace, mesh, mesh, namespace)),
		Entry("match on spiffeId prefix", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp-1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  rules:
    - default:
        allow:
          - spiffeId:
              type: Prefix
              value: spiffe://%s.default.mesh.local
`, namespace, mesh, mesh)),
	)
}
