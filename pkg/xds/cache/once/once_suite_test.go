package once_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOnce(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Once Suite")
}
