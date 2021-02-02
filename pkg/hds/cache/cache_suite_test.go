package cache_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHDSCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HDS Cache Suite")
}
