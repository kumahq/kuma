package helm_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/helm"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Helm Suite")
}

var _ = Describe("Test Control Plane autoscaling with Helm chart", helm.ControlPlaneAutoscalingWithHelmChart)
var _ = Describe("Test Zone and Global with Helm chart", helm.ZoneAndGlobalWithHelmChart)
var _ = Describe("Test App deployment with Helm chart", helm.AppDeploymentWithHelmChart)
var _ = Describe("Test upgrading with Helm chart", helm.UpgradingWithHelmChart)
