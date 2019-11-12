package errors_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Errors")
}
