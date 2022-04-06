package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/universal/auth"
	"github.com/kumahq/kuma/test/e2e/universal/env"
	"github.com/kumahq/kuma/test/e2e/universal/healthcheck"
	. "github.com/kumahq/kuma/test/framework"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Universal Suite")
}

var _ = E2EBeforeSuite(func() {
	env.Cluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)
	E2EDeferCleanup(env.Cluster.DismissCluster)
	Expect(env.Cluster.Install(Kuma(core.Standalone))).To(Succeed())

})

var _ = Describe("User Auth", auth.UserAuth)
var _ = Describe("DP Auth", auth.DpAuth, Ordered)
var _ = Describe("HealthCheck panic threshold", healthcheck.HealthCheckPanicThreshold, Ordered)
