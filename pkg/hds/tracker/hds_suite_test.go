package tracker_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHDSTracker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HDS Tracker Suite")
}
