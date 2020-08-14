package dataplaneinsight_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDataplaneInsightManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataplane Insights Manager Suite")
}
