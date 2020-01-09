package kuma_prometheus_sd_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKumaPrometheusSD(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kuma Prometheus SD Suite")
}
