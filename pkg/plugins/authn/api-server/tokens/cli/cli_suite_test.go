package cli_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test"
)

func TestCliCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunSpecs(t, "CLI Suite")
}
