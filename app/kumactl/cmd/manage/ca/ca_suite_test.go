package ca_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CA Suite")
}
