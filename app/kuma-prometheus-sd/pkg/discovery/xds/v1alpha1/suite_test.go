package v1alpha1_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestXds(t *testing.T) {
	test.RunSpecs(t, "Prometheus SD V1Alpha1 Suite")
}
