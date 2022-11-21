package cni_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/cni"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E CNI Suite")
}

var _ = Describe("Taint controller", Label("job-0"), Label("kind-not-supported"), Label("legacy-k3s-not-supported"), cni.AppDeploymentWithCniAndTaintController)
var _ = Describe("Old CNI", Label("job-0"), Label("arm-not-supported"), Label("legacy-k3s-not-supported"), Label("kind-not-supported"), cni.AppDeploymentWithCniAndNoTaintController)
