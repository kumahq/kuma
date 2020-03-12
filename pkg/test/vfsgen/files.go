package vfsgen

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/onsi/ginkgo/extensions/table"
	"github.com/pkg/errors"
	"github.com/shurcooL/httpfs/vfsutil"
)

type FileTestCase struct {
	Filename         string
	ExpectedContents []byte
}

func GenerateEntries(dir string) []table.TableEntry {
	fs := http.Dir(dir)

	entries := make([]table.TableEntry, 0)

	walkFn := func(path string, fi os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			data, err := ioutil.ReadAll(r)
			if err != nil {
				return errors.Wrapf(err, "failed to read file: %s", path)
			}
			entries = append(entries, table.Entry(path, FileTestCase{
				Filename:         path,
				ExpectedContents: data,
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
