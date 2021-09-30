package ws_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWS(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunSpecs(t, "User Tokens WS Suite")
}
