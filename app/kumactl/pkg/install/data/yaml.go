package data

import (
	"bytes"
	"regexp"
	"strings"
)

var sep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

func SplitYAML(file File) []File {
	files := make([]File, 0)

	bigFile := string(file)
	// Making sure that any extra whitespace in YAML stream doesn't interfere in splitting documents correctly.
	bigFileTmp := strings.TrimSpace(bigFile)
	docs := sep.Split(bigFileTmp, -1)
	for _, doc := range docs {
		if doc == "" {
			continue
		}

		doc = strings.TrimSpace(doc)
		files = append(files, File(doc))
	}

	return files
}

func JoinYAML(files []File) File {
	var buf bytes.Buffer
	for _, file := range files {
		docs := SplitYAML(file)
		for _, doc := range docs {
			if len(doc) == 0 {
				continue
			}
			buf.Write([]byte("\n---\n"))
			buf.Write(doc)
		}
	}
	return buf.Bytes()
}
