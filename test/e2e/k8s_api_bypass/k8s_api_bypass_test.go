package k8s_api_bypass_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/k8s_api_bypass"
)

var _ = Describe("Test Kubernetes API Bypass", k8s_api_bypass.K8sApiBypass)
