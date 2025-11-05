package meshidentity

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	meshtrust_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshtrust/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v2/pkg/kds/hash"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/util/channels"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
	"github.com/kumahq/kuma/v2/test/framework/utils"
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

	trustForLegacyMTLS := func() string {
		trustTmplGlobal := `
type: MeshTrust
mesh: %s
name: identity-trust-global
spec:
  caBundles:
    - type: Pem
      pem:
        value: |-
%s
  trustDomain: %s
`
		type secret struct {
			Data string `json:"data"`
		}

		var oldMeshCA string
		Eventually(func(g Gomega) {
			out, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("get", "secret", "-m", meshName, fmt.Sprintf("%s.ca-builtin-cert-ca-1", meshName), "-o", "json")
			g.Expect(err).ToNot(HaveOccurred())
			oldMeshCA = out
		}, "30s", "1s").Should(Succeed())

		var sec secret
		Expect(json.Unmarshal([]byte(oldMeshCA), &sec)).To(Succeed())

		// Decode from Base64
		decoded, err := base64.StdEncoding.DecodeString(sec.Data)
		Expect(err).ToNot(HaveOccurred())

		return fmt.Sprintf(trustTmplGlobal, meshName, utils.Indent(string(decoded), 10), meshName)
	}

	getMeshTrust := func(zone string) (*meshtrust_api.MeshTrust, error) {
		trust, err := multizone.Global.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtrust", "-m", meshName, hash.HashedName(meshName, hash.HashedName(meshName, "identity-migration"), zone, Config.KumaNamespace), "-ojson")
		if err != nil {
			return nil, err
		}
		r, err := rest.JSON.Unmarshal([]byte(trust), meshtrust_api.MeshTrustResourceTypeDescriptor)
		if err != nil {
			return nil, err
		}
		return r.GetSpec().(*meshtrust_api.MeshTrust), nil
	}

	isMeshIdentityReady := func(cluster *K8sCluster, name string) (bool, error) {
		output, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), cluster.GetKubectlOptions(Config.KumaNamespace), "get", "meshidentity", name, "-ojson")
		if err != nil {
			return false, err
		}
		return strings.Contains(output, "PartiallyReady") || strings.Contains(output, "Successfully initialized"), nil
	}

	It("should migrate from mesh.mTLS to MeshIdentity", func() {
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
name: only-identity-migration
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

		hashedName := hash.HashedName(meshName, "only-identity-migration")
		Expect(WaitForResource(meshidentity_api.MeshIdentityResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: fmt.Sprintf("%s.%s", hashedName, Config.KumaNamespace)}, multizone.KubeZone1, multizone.KubeZone2)).To(Succeed())

		// wait for MeshIdentity to be reconcile
		Eventually(func(g Gomega) {
			isReady, err := isMeshIdentityReady(multizone.KubeZone1, hashedName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(isReady).To(BeTrue())

			isReady, err = isMeshIdentityReady(multizone.KubeZone2, hashedName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(isReady).To(BeTrue())
		}, "30s", "1s").Should(Succeed())

		// create identity without selecting any Dataplane to create builtin certificates
		yaml := fmt.Sprintf(`
type: MeshIdentity
name: identity-migration
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

		hashedName = hash.HashedName(meshName, "identity-migration")
		Expect(WaitForResource(meshidentity_api.MeshIdentityResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: fmt.Sprintf("%s.%s", hashedName, Config.KumaNamespace)}, multizone.KubeZone1, multizone.KubeZone2)).To(Succeed())

		// wait for MeshIdentity to be reconcile
		Eventually(func(g Gomega) {
			isReady, err := isMeshIdentityReady(multizone.KubeZone1, hashedName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(isReady).To(BeTrue())

			isReady, err = isMeshIdentityReady(multizone.KubeZone2, hashedName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(isReady).To(BeTrue())
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// when
		// added Trust from zone 1 to zone 2
		trustTmpl := `
apiVersion: kuma.io/v1alpha1
kind: MeshTrust
metadata:
  name: identity-migration-trust-%s
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
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())
		Expect(NewClusterSetup().
			Install(YamlK8s(fmt.Sprintf(trustTmpl, multizone.KubeZone2.Name(), meshName, multizone.KubeZone1.Name(), utils.Indent(trust2.CABundles[0].PEM.Value, 10), trust2.TrustDomain))).
			Setup(multizone.KubeZone1)).To(Succeed())

		Expect(NewClusterSetup().
			Install(YamlUniversal(trustForLegacyMTLS())).
			Setup(multizone.Global)).To(Succeed())

		// and
		// select all dataplanes
		yaml = fmt.Sprintf(`
type: MeshIdentity
name: identity-migration
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
		hashedName = hash.HashedName(meshName, "identity-migration")
		Expect(WaitForResource(meshidentity_api.MeshIdentityResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: fmt.Sprintf("%s.%s", hashedName, Config.KumaNamespace)}, multizone.KubeZone1, multizone.KubeZone2)).To(Succeed())

		// wait for MeshIdentity to be reconcile
		Eventually(func(g Gomega) {
			isReady, err := isMeshIdentityReady(multizone.KubeZone1, hashedName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(isReady).To(BeTrue())

			isReady, err = isMeshIdentityReady(multizone.KubeZone2, hashedName)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(isReady).To(BeTrue())
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// and disable tls on Mesh
		Expect(NewClusterSetup().
			Install(Yaml(
				builders.Mesh().
					WithName(meshName).
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Setup(multizone.Global)).To(Succeed())

		// then
		Consistently(func(g Gomega) {
			g.Expect(reqError.Load()).To(BeNil())
		}, "5s", "1s").Should(Succeed())
	})
}
