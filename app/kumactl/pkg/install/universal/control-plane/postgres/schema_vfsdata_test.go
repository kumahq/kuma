// +build !dev

package postgres_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/app/kumactl/pkg/install/universal/control-plane/postgres"
)

var _ = Describe("Schema file", func() {

	It("generated Go code must be in sync with the original schema file", func() {
		// given compiled file
		file, err := postgres.Schema.Open("resource.sql")
		Expect(err).ToNot(HaveOccurred())
		expectedContents, err := ioutil.ReadAll(file)
		Expect(err).ToNot(HaveOccurred())

		// and actual file
		kumactlSrcDir := filepath.Join("..", "..", "..", "..", "..")
		schemaDir := postgres.SchemaDir(kumactlSrcDir)
		actualContents, err := ioutil.ReadFile(filepath.Join(schemaDir, "resource.sql"))

		// then both files are identical
		Expect(err).ToNot(HaveOccurred())
		Expect(string(actualContents)).To(Equal(string(expectedContents)), "generated Go code is no longer in sync with the original schema file. To re-generate it, run `make generate/kumactl/install/universal/control-plane/postgres`")
	})

})
