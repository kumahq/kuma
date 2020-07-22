package mux_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMultiplexKDS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multiplex KDS Suite")
}
