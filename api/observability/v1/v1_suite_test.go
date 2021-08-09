package v1_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestV1alpha1(t *testing.T) {
	RegisterFailHandlerWithT(t, Fail)
	RunSpecs(t, "v1 Suite")
}
