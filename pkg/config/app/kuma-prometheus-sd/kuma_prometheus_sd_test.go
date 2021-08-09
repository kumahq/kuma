package kuma_prometheus_sd_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestKumaPrometheusSD(t *testing.T) {
	test.RunSpecs(t, "Kuma Prometheus SD Suite")
}
