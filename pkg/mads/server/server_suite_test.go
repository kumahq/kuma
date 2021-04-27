package server_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMADSServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MADS Server Suite")
}
