package meshidentity

import (
	"fmt"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/kds/hash"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/util/channels"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
	"github.com/kumahq/kuma/test/framework/utils"
)

func Migration() {
	namespace := "meshidentity-migration"
	meshName := "meshidentity-migration"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(
				builders.Mesh().
					WithBuiltinMTLSBackend("ca-1").
					WithEnabledMTLSBackend("ca-1").
					WithName(meshName).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-zone-1"),
				),
				democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName)),
			)).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(meshName),
					testserver.WithEchoArgs("echo", "--instance", "kube-test-server-zone-2"),
				),
				democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName)),
			)).
			SetupInGroup(multizone.KubeZone2, &group)

		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})
	// identity-c2v4v6874cx8x6c8-cww8457w48b482c7
	// identity-c2v4v6874cx8x6c8-w54dw4d47449z9z8

	getMeshTrust := func(zone string) (*meshtrust_api.MeshTrust, error) {
		trust, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtrust", "-m", meshName, hash.HashedName(meshName, hash.HashedName(meshName, "identity"), zone, Config.KumaNamespace), "-ojson")
		if err != nil {
			return nil, err
		}
		r, err := rest.JSON.Unmarshal([]byte(trust), meshtrust_api.MeshTrustResourceTypeDescriptor)
		if err != nil {
			return nil, err
		}
		return r.GetSpec().(*meshtrust_api.MeshTrust), nil
	}

	It("should migrate from mesh.mTLS to MeshIdentity", FlakeAttempts(3), func() {
		// given
		// cross zone traffic works
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", "test-server.meshidentity-migration.svc.kuma-2.mesh.local",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-zone-2"))
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
					multizone.KubeZone1, "demo-client", "test-server.meshidentity-migration.svc.kuma-2.mesh.local",
					client.FromKubernetesPod(namespace, "demo-client"),
				)
				if err != nil {
					reqError.Store(err)
				}
				// the same zone request
				_, err = client.CollectEchoResponse(
					multizone.KubeZone1, "demo-client", "test-server",
					client.FromKubernetesPod(namespace, "demo-client"),
				)
				if err != nil {
					reqError.Store(err)
				}
				time.Sleep(200 * time.Millisecond)
			}
		}()
		// when
		// create only identity to propagate SpiffeID to all MeshServices
		onlyIdentity := fmt.Sprintf(`
type: MeshIdentity
name: only-identity
mesh: %s
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
    path: "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
`, meshName)
		Expect(NewClusterSetup().
			Install(YamlUniversal(onlyIdentity)).
			Setup(multizone.Global)).To(Succeed())

		hashedName := hash.HashedName(meshName, "only-identity")
		Expect(WaitForResource(meshidentity_api.MeshIdentityResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: fmt.Sprintf("%s.%s", hashedName, Config.KumaNamespace)}, multizone.KubeZone1, multizone.KubeZone2)).To(Succeed())

		// create identity without selecting any Dataplane to create builtin certificates
		yaml := fmt.Sprintf(`
type: MeshIdentity
name: identity
mesh: %s
spec:
  selector: {}
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
`, meshName)
		Expect(NewClusterSetup().
			Install(YamlUniversal(yaml)).
			Setup(multizone.Global)).To(Succeed())

		hashedName = hash.HashedName(meshName, "identity")
		Expect(WaitForResource(meshidentity_api.MeshIdentityResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: fmt.Sprintf("%s.%s", hashedName, Config.KumaNamespace)}, multizone.KubeZone1, multizone.KubeZone2)).To(Succeed())

		// when
		// added Trust from zone 1 to zone 2
		trustTmpl := `
apiVersion: kuma.io/v1alpha1
kind: MeshTrust
metadata:
  name: identity-trust-%s
  namespace: kuma-system
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
    kuma.io/zone: %s
spec:
  caBundles:
    - type: Pem
      pem:
        value: |-
%s
  trustDomain: %s
`
		var trust1 *meshtrust_api.MeshTrust
		Eventually(func(g Gomega) {
			var err error
			trust1, err = getMeshTrust(multizone.KubeZone1.Name())
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
		Expect(NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(trustTmpl, multizone.KubeZone1.Name(), meshName, multizone.KubeZone2.Name(), utils.Indent(trust1.CABundles[0].PEM.Value, 10), trust1.TrustDomain))).
			Setup(multizone.KubeZone2)).To(Succeed())

		// and Trust from zone 2 to zone 1
		var trust2 *meshtrust_api.MeshTrust
		Eventually(func(g Gomega) {
			var err error
			trust2, err = getMeshTrust(multizone.KubeZone2.Name())
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
		Expect(NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(trustTmpl, multizone.KubeZone2.Name(), meshName, multizone.KubeZone1.Name(), utils.Indent(trust2.CABundles[0].PEM.Value, 10), trust2.TrustDomain))).
			Setup(multizone.KubeZone1)).To(Succeed())

		// and
		// select all dataplanes
		yaml = fmt.Sprintf(`
type: MeshIdentity
name: identity
mesh: %s
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
`, meshName)
		Expect(NewClusterSetup().
			Install(YamlUniversal(yaml)).
			Setup(multizone.Global)).To(Succeed())
		hashedName = hash.HashedName(meshName, "identity")
		Expect(WaitForResource(meshidentity_api.MeshIdentityResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: fmt.Sprintf("%s.%s", hashedName, Config.KumaNamespace)}, multizone.KubeZone1, multizone.KubeZone2)).To(Succeed())

		// and disable tls on Mesh
		Expect(NewClusterSetup().
			Install(Yaml(
				builders.Mesh().
					WithName(meshName).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Setup(multizone.Global)).To(Succeed())

		// then
		time.Sleep(5 * time.Second) // let the goroutine execute more requests
		Expect(reqError.Load()).To(BeNil())
	})
}
