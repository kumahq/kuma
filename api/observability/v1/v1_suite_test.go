package v1_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestV1alpha1(t *testing.T) {
	test.RunSpecs(t, "Observability v1 Suite")
}
