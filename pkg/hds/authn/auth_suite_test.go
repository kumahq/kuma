package authn_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAuthn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HDS Authn Suite")
}
