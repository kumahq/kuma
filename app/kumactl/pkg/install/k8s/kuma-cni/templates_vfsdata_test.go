// +build !dev

package kumacni_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	kumacni "github.com/Kong/kuma/app/kumactl/pkg/install/k8s/kuma-cni"
	"github.com/Kong/kuma/pkg/test/vfsgen"
)

var _ = Describe("Templates", func() {

	kumactlSrcDir := filepath.Join("..", "..", "..", "..")
	kumacniTemplatesDir := kumacni.TemplatesDir(kumactlSrcDir)
	kumacniTemplatesTestEntries := vfsgen.GenerateEntries(kumacniTemplatesDir)

	DescribeTable("generated Go code must be in sync with the original template files",
		func(given vfsgen.FileTestCase) {
			// when
			file, err := kumacni.Templates.Open(given.Filename)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actualContents, err := ioutil.ReadAll(file)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(string(actualContents)).To(Equal(string(given.ExpectedContents)), "generated Go code is no longer in sync with the original template files. To re-generate it, run `make generate/kumactl/install/k8s/kuma-cni`")
		},
		kumacniTemplatesTestEntries...,
	)
})
