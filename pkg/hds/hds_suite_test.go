package hds_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHDS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HDS Suite")
}
