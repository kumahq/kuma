package helm_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/helm"
)

var _ = Describe("Test Control Plane autoscaling with Helm chart", helm.ControlPlaneAutoscalingWithHelmChart)
