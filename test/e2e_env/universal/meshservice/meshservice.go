package meshservice

import (
	"fmt"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/channels"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func MeshService() {
	meshName := "mesh-service"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TestServerUniversal(
				"first-test-server",
				meshName,
				WithArgs([]string{"echo"}),
				WithServiceName("first-test-server"),
				WithAdditionalTags(map[string]string{
					"app": "test-server",
				}))).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should be able to create and use MeshService with HostnameGenerator", func() {
		otherTLDGenerator := `
type: HostnameGenerator
name: basic-other-tld
labels:
  kuma.io/origin: zone
spec:
  template: "{{ .Name }}.meshservice.othertld"
  selector:
    meshService:
      matchLabels:
        app: backend
`
		service := fmt.Sprintf(`
type: MeshService
name: backend
mesh: "%s"
labels:
  app: backend
  kuma.io/origin: zone
spec:
  ports:
  - port: 80 
    targetPort: 80
    appProtocol: http
  selector:
    dataplaneTags:
      app: test-server
`, meshName)
		Expect(universal.Cluster.Install(YamlUniversal(otherTLDGenerator))).To(Succeed())
		Expect(universal.Cluster.Install(YamlUniversal(service))).To(Succeed())

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "backend.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			_, err = client.CollectEchoResponse(
				universal.Cluster, "demo-client", "backend.meshservice.othertld",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "500ms", MustPassRepeatedly(10)).Should(Succeed())
	})

	It("should switch to permissive mTLS without drop of the traffic", func() {
		// given constant requests to the service
		reqError := atomic.Value{}
		stopCh := make(chan struct{})
		defer close(stopCh)
		go func() {
			for {
				if channels.IsClosed(stopCh) {
					return
				}
				_, err := client.CollectEchoResponse(
					universal.Cluster, "demo-client", "backend.meshservice.othertld",
				)
				if err != nil {
					reqError.Store(err)
				}
				time.Sleep(50 * time.Millisecond)
			}
		}()

		// when
		yaml := `
type: Mesh
name: mesh-service
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
    mode: PERMISSIVE
`
		err := universal.Cluster.Install(YamlUniversal(yaml))

		// then traffic went over mTLS with no errors
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				universal.Cluster, "demo-client", "http://localhost:9901/stats",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring(fmt.Sprintf("cluster.%s_backend__kuma-3_msvc_80.ssl.handshake", meshName)))
			g.Expect(stdout).ToNot(ContainSubstring(fmt.Sprintf("cluster.%s_backend__kuma-3_msvc_80.ssl.handshake: 0", meshName)))
		}, "30s", "1s").Should(Succeed())
		Expect(reqError.Load()).To(BeNil())

		// when switch to strict mTLS
		err = universal.Cluster.Install(MTLSMeshUniversal(meshName))

		// then
		Expect(err).ToNot(HaveOccurred())
		time.Sleep(5 * time.Second) // let the goroutine execute more requests
		Expect(reqError.Load()).To(BeNil())
	})
}
