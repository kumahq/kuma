package gc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GC Suite")
}
