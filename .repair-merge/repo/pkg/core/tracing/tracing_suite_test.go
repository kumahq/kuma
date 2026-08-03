package tracing_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestTracing(t *testing.T) {
	test.RunSpecs(t, "Tracing Suite")
}
