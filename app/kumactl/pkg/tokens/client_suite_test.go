package tokens_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTokensClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tokens Client Suite")
}
