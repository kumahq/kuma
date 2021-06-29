package installer

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/coreos/etcd/pkg/fileutil"
)

func copyDir(srcDir string, dstDir string) error {
	if fileutil.IsDirWriteable(dstDir) != nil {
		return errors.New(fmt.Sprintf("Can't write to Directory %s", dstDir))
	}

	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		filename := f.Name()

		srcFilepath := filepath.Join(srcDir, filename)
		err := fileAtomicCopy(srcFilepath, dstDir, filename)
		if err != nil {
			return err
		}
		log.Printf("Successful copy %s to %s.", filename, dstDir)
	}

	return nil
}
