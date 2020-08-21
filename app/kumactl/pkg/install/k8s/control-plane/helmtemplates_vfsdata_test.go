// +build !dev

package controlplane_test

import (
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	controlplane "github.com/kumahq/kuma/app/kumactl/pkg/install/k8s/control-plane"
	"github.com/kumahq/kuma/pkg/test/vfsgen"
)

var _ = Describe("Templates", func() {

	rootSrcDir := filepath.Join("..", "..", "..", "..", "..", "..")
	controlplaneTemplatesDir := controlplane.HelmTemplatesDir(rootSrcDir)
	controlplaneTemplatesTestEntries := vfsgen.GenerateEntries(controlplaneTemplatesDir)

	DescribeTable("generated Go code must be in sync with the original template files",
		func(given vfsgen.FileTestCase) {
			// when
			file, err := controlplane.HelmTemplates.Open(given.Filename)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actualContents, err := ioutil.ReadAll(file)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(string(actualContents)).To(Equal(string(given.ExpectedContents)), "generated Go code is no longer in sync with the original template files. To re-generate it, run `make generate/kumactl/install/k8s/control-plane`")
		},
		controlplaneTemplatesTestEntries...,
	)
})
