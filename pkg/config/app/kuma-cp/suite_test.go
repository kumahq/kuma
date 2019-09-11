package kuma_cp

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCpConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Control Plane Configuration Suite")
}
