package maps_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMaps(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Maps Suite")
}
