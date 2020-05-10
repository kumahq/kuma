package completion_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCompletionCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Completion Cmd Suite")
}
