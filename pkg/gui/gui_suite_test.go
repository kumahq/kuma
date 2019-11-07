package gui_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GUI Suite")
}
