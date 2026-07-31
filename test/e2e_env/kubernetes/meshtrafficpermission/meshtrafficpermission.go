package meshtrafficpermission

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func MeshTrafficPermissionRules() {
	mesh := "mtp-rules"
	namespace := "mtp-rules-ns"
	testServerURL := fmt.Sprintf("test-server.%s.svc:80", namespace)

	meshIdentity := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshIdentity
metadata:
  name: identity-mtp
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

	meshWideAllowPolicy := func(spiffePrefix string) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp-allow-all
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        allow:
          - spiffeID:
              type: Prefix
              value: %s
`, namespace, mesh, spiffePrefix)
	}

	testServerDenyPolicy := func(spiffePrefix string) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp-deny-demo-client
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: test-server
  rules:
    - default:
        deny:
          - spiffeID:
              type: Prefix
              value: %s
`, namespace, mesh, spiffePrefix)
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(Yaml(builders.Mesh().
				WithName(mesh))).
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
		}, "2m", "3s").Should(Succeed())

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
		}, "2m", "3s").Should(Succeed())
	},
		Entry("exact match on spiffeID", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp-identity
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  rules:
    - default:
        allow:
          - spiffeID:
              type: Exact
              value: spiffe://%s.default.mesh.local/ns/%s/sa/default
`, namespace, mesh, mesh, namespace)),
		Entry("match on spiffeID prefix", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp-identity
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  rules:
    - default:
        allow:
          - spiffeID:
              type: Prefix
              value: spiffe://%s.default.mesh.local
`, namespace, mesh, mesh)),
	)

	It("should prefer a more specific dataplane deny over a mesh-wide allow", func() {
		// given
		allowSpiffePrefix := fmt.Sprintf("spiffe://%s.default.mesh.local", mesh)
		denySpiffePrefix := fmt.Sprintf("%s/ns/%s", allowSpiffePrefix, namespace)

		// when
		Expect(YamlK8s(meshWideAllowPolicy(allowSpiffePrefix))(kubernetes.Cluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				kubernetes.Cluster, "demo-client", testServerURL,
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(ContainSubstring("test-server"))
		}, "2m", "3s").Should(Succeed())

		// when
		Expect(YamlK8s(testServerDenyPolicy(denySpiffePrefix))(kubernetes.Cluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			resp, err := client.CollectFailure(
				kubernetes.Cluster, "demo-client", testServerURL,
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.ResponseCode).To(Equal(403))
		}, "2m", "3s").Should(Succeed())
	})
}
