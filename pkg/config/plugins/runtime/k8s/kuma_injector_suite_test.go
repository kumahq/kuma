package k8s_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestKumaInjector(t *testing.T) {
	test.RunSpecs(t, "KumaInjector Suite")
}
