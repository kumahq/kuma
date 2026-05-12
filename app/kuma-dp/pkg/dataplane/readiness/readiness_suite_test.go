package readiness_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestReadiness(t *testing.T) {
	test.RunSpecs(t, "Readiness Reporter Suite")
}
