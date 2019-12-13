package admin_server_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAdminServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Admin Server Suite")
}
