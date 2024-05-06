package data

import (
	"bytes"
	"regexp"
	"strings"
)

var sep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

func SplitYAML(file File) []File {
	files := make([]File, 0)

	bigFile := string(file.Data)
	// Making sure that any extra whitespace in YAML stream doesn't interfere in splitting documents correctly.
	bigFileTmp := strings.TrimSpace(bigFile)
	docs := sep.Split(bigFileTmp, -1)
	for _, doc := range docs {
		if doc == "" {
			continue
		}

		doc = strings.TrimSpace(doc)
		file := File{
			Data: []byte(doc),
			Name: file.Name,
		}
		files = append(files, file)
	}

	return files
}

func JoinYAML(files []File) File {
	var buf bytes.Buffer
	for _, file := range files {
		docs := SplitYAML(file)
		for _, doc := range docs {
			if len(doc.Data) == 0 {
				continue
			}
			buf.Write([]byte("\n---\n"))
			buf.Write(doc.Data)
		}
	}
	file := File{
		Data: buf.Bytes(),
	}
	return file
}
