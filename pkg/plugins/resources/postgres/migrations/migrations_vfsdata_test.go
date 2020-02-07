// +build !dev

package migrations_test

import (
	"io/ioutil"

	"github.com/Kong/kuma/pkg/plugins/resources/postgres/migrations"
	"github.com/Kong/kuma/pkg/test/vfsgen"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migration files", func() {

	migrationsTestEntries := vfsgen.GenerateEntries(migrations.MigrationsDir())

	DescribeTable("generated Go code must be in sync with the original template files",
		func(given vfsgen.FileTestCase) {
			// when
			file, err := migrations.Migrations.Open(given.Filename)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actualContents, err := ioutil.ReadAll(file)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(string(actualContents)).To(Equal(string(given.ExpectedContents)), "generated Go code is no longer in sync with the original template files. To re-generate it, run `make generate/kuma-cp/migrations`")
		},
		migrationsTestEntries...,
	)
})
