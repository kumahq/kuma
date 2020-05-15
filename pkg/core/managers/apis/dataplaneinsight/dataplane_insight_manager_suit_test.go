package dataplaneinsight_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestDataplainInsightManaget(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mesh Manager Suite")
}
