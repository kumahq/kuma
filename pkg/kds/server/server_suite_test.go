package server_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKDS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KDS Suite")
}
