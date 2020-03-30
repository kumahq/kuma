package tags_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTags(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Listener Suite")
}
