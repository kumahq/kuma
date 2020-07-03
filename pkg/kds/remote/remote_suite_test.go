package remote_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRemoteSync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Remote Sync Suite")
}
