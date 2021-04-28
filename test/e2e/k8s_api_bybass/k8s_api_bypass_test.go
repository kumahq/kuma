package k8s_api_bybass_test

import (
	"github.com/kumahq/kuma/test/e2e/k8s_api_bybass"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Test Kubernetes API Bypass", k8s_api_bybass.K8sApiBypass)
