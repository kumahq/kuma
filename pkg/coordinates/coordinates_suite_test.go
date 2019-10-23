package coordinates_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCoordinates(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Coordinates")
}
