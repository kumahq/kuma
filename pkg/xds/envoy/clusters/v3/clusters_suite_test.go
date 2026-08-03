package clusters_test

import (
	"testing"
	"time"

	"github.com/kumahq/kuma/v3/pkg/test"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

func DefaultTimeout() envoy_common.Timeouts {
	return envoy_common.Timeouts{
		Connect: 5 * time.Second,
	}
}

func TestClusters(t *testing.T) {
	test.RunSpecs(t, "Clusters V3 Suite")
}
