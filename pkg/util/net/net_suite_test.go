package net_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNet(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Net Suite")
}
