package k8s_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestK8s(t *testing.T) {
	test.RunSpecs(t, "K8s Suite")
}
