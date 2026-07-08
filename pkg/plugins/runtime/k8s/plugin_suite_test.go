package k8s

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestK8s(t *testing.T) {
	test.RunSpecs(t, "K8s Plugin Suite")
}
