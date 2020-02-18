package accesslog_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAccesslog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Accesslog Suite")
}
