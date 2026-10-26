package opentelemetry

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestOpenTelemetry(t *testing.T) {
	test.RunSpecs(t, "OpenTelemetry Plugin Suite")
}
