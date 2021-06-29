package passivehealth_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPassiveHealth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Passive Health Suite")
}
