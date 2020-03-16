package resources_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/app/kuma-ui/pkg/resources"
	"github.com/Kong/kuma/pkg/test/vfsgen"
)

var _ = Describe("Gui Dir", func() {

	guiDir := filepath.Join("..", "..", "data", "resources")
	testEntries := vfsgen.GenerateEntries(guiDir)

	DescribeTable("generated Go code must be in sync with the original GUI dir",
		func(given vfsgen.FileTestCase) {
			// given compiled file
			file, err := resources.GuiDir.Open(given.Filename)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actualContents, err := ioutil.ReadAll(file)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(string(actualContents)).To(Equal(string(given.ExpectedContents)), "generated Go code is no longer in sync with the original template files. To re-generate it, run `make generate/gui`")
		},
		testEntries...,
	)

})
