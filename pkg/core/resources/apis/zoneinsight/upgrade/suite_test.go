package upgrade_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestZoneInsightUpgrade(t *testing.T) {
	test.RunSpecs(t, "ZoneInsight Kubernetes Upgrade Suite")
}
