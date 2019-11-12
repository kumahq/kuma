package gui_test

import (
	"github.com/Kong/kuma/pkg/gui"
	"github.com/Kong/kuma/pkg/test/vfsgen"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gui Dir", func() {

	guiDir := filepath.Join("..", "..", "gui")
	testEntries := vfsgen.GenerateEntries(guiDir)

	DescribeTable("generated Go code must be in sync with the original GUI dir",
		func(given vfsgen.FileTestCase) {
			// given compiled file
			file, err := gui.GuiDir.Open(given.Filename)
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
