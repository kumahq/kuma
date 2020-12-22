package zone_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestZoneManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zone Manager Suite")
}
