package callbacks_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCallbacks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "XDS Callbacks Suite")
}
