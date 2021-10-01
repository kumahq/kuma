package ws_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test"
)

func TestWS(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunSpecs(t, "User Tokens WS Suite")
}
