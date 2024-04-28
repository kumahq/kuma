package context

import (
	"github.com/kumahq/kuma/pkg/util/data"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Override profile values", func() {

	It("should apply profile values and remove default values.yaml", func() {
		loadedFiles := createFiles("values.production.yaml", "values.yaml")
		files, err := useProfileValues(loadedFiles, "production")
		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(files[0].Name).To(Equal("values.yaml"))
		Expect(files[0].FullPath).To(Equal("values.yaml"))
		Expect(string(files[0].Data)).To(Equal("values.production.yaml"))
	})

	It("should use default values.yaml when use demo profile", func() {
		loadedFiles := createFiles("values.yaml", "templates/config.yaml")
		files, err := useProfileValues(loadedFiles, "demo")
		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(HaveLen(2))
		Expect(files[0].Name).To(Equal("values.yaml"))
		Expect(string(files[0].Data)).To(Equal("values.yaml"))
	})

	It("should keep values.yaml when use empty profile", func() {
		loadedFiles := createFiles("values.yaml")
		files, err := useProfileValues(loadedFiles, "demo")
		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(files[0].Name).To(Equal("values.yaml"))
		Expect(string(files[0].Data)).To(Equal("values.yaml"))
	})

	It("should keep default values.yaml when profile does not exist", func() {
		loadedFiles := createFiles("values.yaml")
		files, err := useProfileValues(loadedFiles, "default")
		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(files[0].Name).To(Equal("values.yaml"))
	})
})

func createFiles(relativePaths ...string) []data.File {
	var loaded []data.File
	for _, p := range relativePaths {
		loaded = append(loaded, createFile(p))
	}
	return loaded
}

func createFile(relativePath string) data.File {
	dirName := "/fake/path/"
	name := relativePath
	lastIdxOfSlash := strings.LastIndex(relativePath, "/")
	if lastIdxOfSlash > -1 {
		name = relativePath[lastIdxOfSlash+1:]
	}

	return data.File{
		Data:     []byte(name),
		Name:     name,
		FullPath: dirName + relativePath,
	}
}
