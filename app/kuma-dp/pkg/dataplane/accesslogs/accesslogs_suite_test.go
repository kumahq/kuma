package accesslogs_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAccesslogs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Accesslogs Suite")
}
