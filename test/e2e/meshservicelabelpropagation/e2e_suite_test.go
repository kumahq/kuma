package meshservicelabelpropagation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/meshservicelabelpropagation"
	"github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E MeshService Label Propagation Universal Suite")
}

var _ = framework.E2ESynchronizedBeforeSuite(
	func() []byte {
		meshservicelabelpropagation.Cluster = framework.NewUniversalCluster(framework.NewTestingT(), framework.Kuma3, framework.Silent)
		err := framework.NewClusterSetup().
			Install(framework.Kuma(config_core.Zone,
				framework.WithEnv("KUMA_MESH_SERVICE_LABEL_PROPAGATION_ENABLED", "true"),
			)).
			Setup(meshservicelabelpropagation.Cluster)
		Expect(err).ToNot(HaveOccurred())
		return nil
	},
	func(_ []byte) {},
)

var _ = framework.E2EAfterSuite(func() {
	if meshservicelabelpropagation.Cluster == nil {
		return
	}
	Expect(meshservicelabelpropagation.Cluster.DismissCluster()).To(Succeed())
})

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("MeshService Label Propagation on Universal", Label("job-2"), Ordered, meshservicelabelpropagation.LabelPropagation, Serial)
)
