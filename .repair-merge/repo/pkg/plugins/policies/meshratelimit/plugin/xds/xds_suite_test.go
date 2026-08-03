package xds

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestXDS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MeshRateLimit XDS Suite")
}
