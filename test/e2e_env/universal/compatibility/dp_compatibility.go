package compatibility

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func UniversalCompatibility() {
	meshName := "compatibility"

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("demo-client-latest", meshName,
				WithTransparentProxy(true)),
			).
			Install(TestServerUniversal("test-server-latest", meshName,
				WithArgs([]string{"echo", "--instance", "test-server-latest"}),
				WithServiceName("test-server-latest")),
			).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	DescribeTable("connection between old and new DPP",
		func(version string) {
			serverName := fmt.Sprintf("test-server-%s", strings.ReplaceAll(version, ".", "-"))
			Expect(universal.Cluster.Install(TestServerUniversal(serverName, meshName,
				WithServiceName(serverName),
				WithArgs([]string{"echo", "--instance", serverName}),
				WithDPVersion(version),
			))).To(Succeed())

			clientName := fmt.Sprintf("demo-client-%s", strings.ReplaceAll(version, ".", "-"))
			Expect(universal.Cluster.Install(DemoClientUniversal(clientName, meshName,
				WithTransparentProxy(true),
				WithDPVersion(version),
			))).To(Succeed())

			Eventually(func(g Gomega) {
				// New client can reach new server
				_, err := client.CollectEchoResponse(universal.Cluster, "demo-client-latest", "test-server-latest.mesh")
				g.Expect(err).ToNot(HaveOccurred())
				// Old client can reach new server
				_, err = client.CollectEchoResponse(universal.Cluster, clientName, "test-server-latest.mesh")
				g.Expect(err).ToNot(HaveOccurred())
				// New client can reach old server
				_, err = client.CollectEchoResponse(universal.Cluster, "demo-client-latest", serverName+".mesh")
				g.Expect(err).ToNot(HaveOccurred())
			}, "20s", "250ms").Should(Succeed())
		},
		EntryDescription("from version: %s"),
		SupportedVersionEntries(universal.Cluster.GetTesting()))
}
