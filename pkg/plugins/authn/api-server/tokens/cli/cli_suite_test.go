package cli_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCliCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunSpecs(t, "CLI Suite")
}
