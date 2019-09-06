// +build !dev

package controlplane_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/shurcooL/httpfs/vfsutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	controlplane "github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/install/k8s/control-plane"
)

var _ = Describe("Templates", func() {

	type testCase struct {
		filename         string
		expectedContents []byte
	}

	kumactlSrcDir := filepath.Join("..", "..", "..", "..")

	controlplaneTemplatesDir := controlplane.TemplatesDir(kumactlSrcDir)

	generateTestEntries := func(dir string) []TableEntry {
		fs := http.Dir(dir)

		entries := make([]TableEntry, 0)

		walkFn := func(path string, fi os.FileInfo, r io.ReadSeeker, err error) error {
			if err != nil {
				return err
			}
			if !fi.IsDir() {
				data, err := ioutil.ReadAll(r)
				if err != nil {
					return errors.Wrapf(err, "failed to read file: %s", path)
				}
				entries = append(entries, Entry(path, testCase{
					filename:         path,
					expectedContents: data,
				}))
			}
			return nil
		}

		err := vfsutil.WalkFiles(fs, "/", walkFn)
		if err != nil {
			panic(err) // Gomega assertions are not available outside of `It()` block
		}

		return entries
	}

	controlplaneTemplatesTestEntries := generateTestEntries(controlplaneTemplatesDir)

	DescribeTable("generated Go code must be in sync with the original template files",
		func(given testCase) {
			// when
			file, err := controlplane.Templates.Open(given.filename)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actualContents, err := ioutil.ReadAll(file)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(string(actualContents)).To(Equal(string(given.expectedContents)), "generated Go code is no longer in sync with the original template files. To re-generate it, run `make generate/kumactl/install/control-plane`")
		},
		controlplaneTemplatesTestEntries...,
	)
})
