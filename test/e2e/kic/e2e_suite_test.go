package kic_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/kic"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E KIC Suite")
}

var _ = Describe("Kong Ingress on Kubernetes", kic.KICKubernetes)
