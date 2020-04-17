package kumacni

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKumaCNI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KumaCNI Suite")
}
