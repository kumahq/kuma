package metrics

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestMetrics(t *testing.T) {
	test.RunSpecs(t, "Metrics Hijacker Suite")
}
