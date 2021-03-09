package insights_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInsights(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Insights Suite")
}
