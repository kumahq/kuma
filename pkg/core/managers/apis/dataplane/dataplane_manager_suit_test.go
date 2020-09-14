package dataplane_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDataplaneManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataplane Manager Suite")
}
