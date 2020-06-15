// +build !dev

package ingress_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	ingress "github.com/Kong/kuma/app/kumactl/pkg/install/k8s/ingress"
	"github.com/Kong/kuma/pkg/test/vfsgen"
)

var _ = Describe("Templates", func() {

	kumactlSrcDir := filepath.Join("..", "..", "..", "..")
	ingressTemplatesDir := ingress.TemplatesDir(kumactlSrcDir)
	ingressTemplatesTestEntries := vfsgen.GenerateEntries(ingressTemplatesDir)

	DescribeTable("generated Go code must be in sync with the original template files",
		func(given vfsgen.FileTestCase) {
			// when
			file, err := ingress.Templates.Open(given.Filename)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actualContents, err := ioutil.ReadAll(file)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(string(actualContents)).To(Equal(string(given.ExpectedContents)), "generated Go code is no longer in sync with the original template files. To re-generate it, run `make generate/kumactl/install/k8s/ingress`")
		},
		ingressTemplatesTestEntries...,
	)
})
