package zoneinsight_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestZoneInsightManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zone Insights Manager Suite")
}
