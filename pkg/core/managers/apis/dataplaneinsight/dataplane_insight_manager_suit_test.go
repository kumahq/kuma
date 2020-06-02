package dataplaneinsight_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDataplainInsightManaget(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mesh Manager Suite")
}
