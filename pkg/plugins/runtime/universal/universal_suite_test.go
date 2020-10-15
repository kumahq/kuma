package universal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUniversal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Universal Runtime Plugin")
}
