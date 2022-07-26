package cni_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e_k3d_only/cni"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E CNI Suite")
}

var _ = Describe("Taint controller", Label("arm-not-supported"), cni.AppDeploymentWithCniAndTaintController)
var _ = Describe("Old CNI", Label("arm-not-supported"), cni.AppDeploymentWithCniAndNoTaintController)
