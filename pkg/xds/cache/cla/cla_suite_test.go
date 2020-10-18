package cla_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCLA(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLA Suite")
}
