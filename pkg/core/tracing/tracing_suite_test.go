package tracing_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestTracing(t *testing.T) {
	test.RunSpecs(t, "Tracing Suite")
}
