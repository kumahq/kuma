package skipinboundtags_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/skipinboundtags"
	"github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/report"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Skip Inbound Tags Kubernetes Suite")
}

var _ = framework.E2EBeforeSuite(func() {
	skipinboundtags.KubeCluster = framework.NewK8sCluster(framework.NewTestingT(), framework.Kuma1, framework.Silent)

	err := framework.NewClusterSetup().
		Install(framework.Kuma(config_core.Zone,
			framework.WithEnv("KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED", "true"),
		)).
		Setup(skipinboundtags.KubeCluster)

	Expect(err).ToNot(HaveOccurred())
})

var _ = framework.E2EAfterSuite(func() {
	Expect(skipinboundtags.KubeCluster.DeleteKuma()).To(Succeed())
	Expect(skipinboundtags.KubeCluster.DismissCluster()).To(Succeed())
})

var (
	_ = ReportAfterSuite("report suite", report.DumpReport)
	_ = Describe("Skip Inbound Tags on Kubernetes", Label("job-3"), skipinboundtags.SkipInboundTags, Ordered)
)
