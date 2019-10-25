package catalogue_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCatalogue(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Catalogue")
}
