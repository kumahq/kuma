package datasource_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDataSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DataSource Suite")
}
