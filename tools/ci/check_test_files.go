package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var allDirs []string
	for _, dir := range []string{"app", "pkg"} {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				allDirs = append(allDirs, path)
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	var res []string
	for _, dir := range allDirs {
		testFiles := false
		suiteFile := false
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() && path != dir { // Don't go deeper
				return filepath.SkipDir
			}
			if strings.HasSuffix(d.Name(), "suite_test.go") {
				suiteFile = true
			} else if strings.HasSuffix(d.Name(), "_test.go") {
				testFiles = true
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
		if testFiles && !suiteFile {
			res = append(res, fmt.Sprintf("%s has test files but no suite file, this will not run the attached tests", dir))
		}
		if suiteFile && !testFiles {
			res = append(res, fmt.Sprintf("%s has a suite file but no test file, this will make tests fail", dir))
		}
	}
	for _, r := range res {
		fmt.Printf("ERROR: %s\n", r)
	}
	if len(res) > 0 {
		os.Exit(1)
	}
}
