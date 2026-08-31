package zoneegress

import (
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	"github.com/kumahq/kuma/v2/pkg/util/channels"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/envs/multizone"
)

func Scaling() {
	const meshName = "ze-scaling"
	const namespace = "ze-scaling"

	mesh := fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  zoneEgress: true
`, meshName)

	destination := fmt.Sprintf("test-server_%s_svc_80.mesh", namespace)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(YamlUniversal(mesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
			)).
			SetupInGroup(multizone.KubeZone2, &group)
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	AfterAll(func() {
		Expect(ScaleApp(multizone.KubeZone1, Config.ZoneEgressApp, Config.KumaNamespace, 1)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should not drop requests while the zone egress scales up and down", func() {
		// given cross-zone traffic flowing through the local zone egress
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponsesAndFailures(
				multizone.KubeZone1, "demo-client", destination,
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithNumberOfRequests(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(HaveEach(HaveField("ResponseCode", 200)))
		}, "2m", "1s").Should(Succeed())

		// and a client sending requests without interruption. A request that never reaches the server
		// is reported with an exit code and no response code, so anything other than 200 is a drop.
		var mtx sync.Mutex
		var drops []client.FailureResponse
		var errs []error
		stopCh := make(chan struct{})
		doneCh := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(doneCh)
			for !channels.IsClosed(stopCh) {
				responses, err := client.CollectResponsesAndFailures(
					multizone.KubeZone1, "demo-client", destination,
					client.FromKubernetesPod(namespace, "demo-client"),
					client.WithNumberOfRequests(10),
				)
				mtx.Lock()
				if err != nil {
					errs = append(errs, err)
				}
				for _, response := range responses {
					if response.ResponseCode != 200 {
						drops = append(drops, response)
					}
				}
				mtx.Unlock()
			}
		}()
		defer func() {
			close(stopCh)
			Eventually(doneCh, "1m").Should(BeClosed())
		}()

		noDrops := func(g Gomega) {
			mtx.Lock()
			defer mtx.Unlock()
			g.Expect(errs).To(BeEmpty())
			g.Expect(drops).To(BeEmpty())
		}

		// when a second egress instance joins
		Expect(ScaleApp(multizone.KubeZone1, Config.ZoneEgressApp, Config.KumaNamespace, 2)).To(Succeed())
		Expect(multizone.KubeZone1.Install(WaitPodsAvailable(Config.KumaNamespace, Config.ZoneEgressApp))).To(Succeed())

		// then nothing is dropped, and the clients have had time to pick the new instance up
		Consistently(noDrops, "20s", "1s").Should(Succeed())

		// when one of the two instances is terminated. ScaleApp returns once the Pod is actually gone,
		// so the whole termination window - from SIGTERM to the process exiting - is covered by traffic.
		Expect(ScaleApp(multizone.KubeZone1, Config.ZoneEgressApp, Config.KumaNamespace, 1)).To(Succeed())

		// then nothing is dropped. Keep sending for a while after the Pod is gone: without the filter the
		// endpoint outlives the process by a control plane reconcile, and that is where the resets land.
		Consistently(noDrops, "20s", "1s").Should(Succeed())
	})
}
