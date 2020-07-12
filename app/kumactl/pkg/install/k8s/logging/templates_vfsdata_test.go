// +build !dev

package logging_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/app/kumactl/pkg/install/k8s/logging"
	"github.com/Kong/kuma/pkg/test/vfsgen"
)

var _ = Describe("Templates", func() {

	kumactlSrcDir := filepath.Join("..", "..", "..", "..")
	loggingTemplatesDir := logging.TemplatesDir(kumactlSrcDir)
	loggingTemplatesTestEntries := vfsgen.GenerateEntries(loggingTemplatesDir)

	DescribeTable("generated Go code must be in sync with the original template files",
		func(given vfsgen.FileTestCase) {
			// when
			file, err := logging.Templates.Open(given.Filename)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actualContents, err := ioutil.ReadAll(file)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(string(actualContents)).To(Equal(string(given.ExpectedContents)), "generated Go code is no longer in sync with the original template files. To re-generate it, run `make generate/kumactl/install/k8s/logging`")
		},
		loggingTemplatesTestEntries...,
	)
})
