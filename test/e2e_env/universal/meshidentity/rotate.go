package meshidentity

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v2/pkg/util/channels"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func Rotate() {
	meshName := "meshidentity-rotate"

	BeforeEach(func() {
		Expect(NewClusterSetup().
			Install(ResourceUniversal(samples.MeshDefaultBuilder().WithName(meshName).WithMeshServicesEnabled(v1alpha1.Mesh_MeshServices_Exclusive).Build())).
			Install(DemoClientUniversal("rotate-demo-client", meshName,
				WithTransparentProxy(true),
				WithWorkload("demo-client"),
			)).
			Install(TestServerUniversal("rotate-test-server", meshName,
				WithArgs([]string{"echo", "--instance", "rotate-test-server"}),
				WithWorkload("test-server"),
			)).
			Setup(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	isMeshIdentityReady := func(cluster *UniversalCluster, name string, hasSelector bool) (bool, error) {
		output, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshidentity", "-m", meshName, name, "-o", "json")
		if err != nil {
			return false, err
		}
		if hasSelector {
			return strings.Contains(output, "matchLabels") && (strings.Contains(output, "PartiallyReady") || strings.Contains(output, "Successfully initialized")), nil
		}
		return strings.Contains(output, "Successfully initialized"), nil
	}

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should rotate a bundled CA without any downtime", func() {
		trustDomain := fmt.Sprintf("%s.mesh.local", meshName)
		Expect(NewClusterSetup().
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshIdentity
name: identity-rotate
mesh: %s
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: %s
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
`, meshName, trustDomain))).
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshTrafficPermission
name: demo-client-to-test-server
mesh: %s
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/workload: test-server
  rules:
  - default:
      allow:
      - spiffeID:
          type: Exact
          value: spiffe://%s/workload/demo-client
`, meshName, trustDomain))).
			Setup(universal.Cluster)).To(Succeed())

		// given
		// communication works
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				universal.Cluster, "rotate-demo-client", "test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("rotate-test-server"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// and
		// start constant requests
		reqError := atomic.Value{}
		stopCh := make(chan struct{})
		defer close(stopCh)
		go func() {
			for {
				if channels.IsClosed(stopCh) {
					return
				}
				// cross zone request
				_, err := client.CollectEchoResponse(
					universal.Cluster, "rotate-demo-client", "test-server.svc.mesh.local",
				)
				println(fmt.Sprintf("%v", err))
				if err != nil {
					reqError.Store(err)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()

		// when
		// create a new MeshIdentity to initialize MeshTrust
		Expect(NewClusterSetup().
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshIdentity
name: a-identity-rotate
mesh: %s
spec:
  selector: {}
  spiffeID:
    trustDomain: %s
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
`, meshName, trustDomain))).
			Setup(universal.Cluster)).To(Succeed())

		// then
		// wait to be reconcile
		Eventually(func(g Gomega) {
			ready, err := isMeshIdentityReady(universal.Cluster, "a-identity-rotate", false)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(ready).To(BeTrue())
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		Expect(NewClusterSetup().
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshIdentity
name: a-identity-rotate
mesh: %s
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: %s
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
`, meshName, trustDomain))).
			Setup(universal.Cluster)).To(Succeed())

		// then
		// wait to be reconcile
		Eventually(func(g Gomega) {
			ready, err := isMeshIdentityReady(universal.Cluster, "a-identity-rotate", true)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(ready).To(BeTrue())
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// when
		// old identity is removed
		Expect(DeleteMeshPolicyOrError(universal.Cluster, meshidentity_api.MeshIdentityResourceTypeDescriptor, "identity-rotate", meshName)).To(Succeed())

		// then
		// we shouldn't observe any downtime
		Consistently(func(g Gomega) {
			g.Expect(reqError.Load()).To(BeNil())
		}, "5s", "1s").Should(Succeed())
	})
}
