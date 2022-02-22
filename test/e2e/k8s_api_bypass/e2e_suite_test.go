package k8s_api_bypass_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/k8s_api_bypass"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Kubernetes API Bypass")
}

var _ = Describe("Test Kubernetes API Bypass", k8s_api_bypass.K8sApiBypass)
