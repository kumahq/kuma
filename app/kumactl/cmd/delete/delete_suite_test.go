package delete_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDeleteCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Delete Cmd Suite")
}
