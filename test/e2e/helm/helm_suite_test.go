package helm_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/helm"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EHelm(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Helm Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Control Plane autoscaling with Helm chart", helm.ControlPlaneAutoscalingWithHelmChart)
var _ = Describe("Test Zone and Global with Helm chart", helm.ZoneAndGlobalWithHelmChart)
var _ = Describe("Test App deployment with Helm chart", helm.AppDeploymentWithHelmChart)
var _ = Describe("Test upgrading with Helm chart", helm.UpgradingWithHelmChart)
