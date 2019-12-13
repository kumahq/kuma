package rest_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCaProvidedRest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rest CA Provided Suite")
}
