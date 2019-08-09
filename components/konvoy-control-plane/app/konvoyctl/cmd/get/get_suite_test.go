package get_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGetCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Get Cmd Suite")
}
