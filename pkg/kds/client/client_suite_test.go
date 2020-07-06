package client_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKDSClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KDS Client Suite")
}
