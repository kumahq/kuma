package zone_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestZoneSync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zone Sync Suite")
}
