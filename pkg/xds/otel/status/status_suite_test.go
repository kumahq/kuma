package status_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OTEL Status Suite")
}
