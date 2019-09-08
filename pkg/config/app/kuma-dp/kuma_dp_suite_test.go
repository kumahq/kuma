package kumadp_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKumaDp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KumaDp Suite")
}
