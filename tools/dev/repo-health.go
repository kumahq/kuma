package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/exp/slices"
)

func main() {
	action := flag.String("action", "", "the action (MADR,README)")
	circleCI := flag.String("circleci-badge", "", "circle CI badge pattern")
	flag.Parse()
	switch *action {
	case "MADR":
		panic("Not implemented")
	case "README":
		f, err := os.Open("active-branches.json")
		defer f.Close()
		if err != nil {
			panic(err)
		}
		var versions []string
		err = json.NewDecoder(f).Decode(&versions)
		if err != nil {
			panic(err)
		}
		out := strings.Builder{}
		for i := range versions {
			out.WriteString(strings.ReplaceAll(*circleCI, "{{branch}}", versions[i]))
			out.WriteString("\n")
		}
		err = replaceInFile("README.md", "<!--CIBADGES-->", out.String())
		if err != nil {
			panic(err)
		}
	default:
		panic("Unknown action:" + *action)
	}
}

func replaceInFile(fileName string, marker string, content string) error {
	out := bytes.Buffer{}
	bytesRead, _ := ioutil.ReadFile(fileName)
	fileContent := string(bytesRead)
	lines := strings.Split(fileContent, "\n")
	startMarkerIdx := slices.Index(lines, marker)
	if startMarkerIdx == -1 {
		return fmt.Errorf("no %s marker", marker)
	}
	// Write start of file
	out.WriteString(strings.Join(lines[0:startMarkerIdx+1], "\n"))
	out.WriteString("\n")
	// Add new content
	out.WriteString(content)
	// Skip all existing lines
	endMarkerIdx := slices.Index(lines[startMarkerIdx+1:], marker)
	if endMarkerIdx == -1 {
		return fmt.Errorf("no %s marker", marker)
	}
	endMarkerIdx += startMarkerIdx + 1
	out.WriteString(strings.Join(lines[endMarkerIdx:], "\n"))
	out.WriteString("\n")

	return os.WriteFile(fileName, out.Bytes(), os.ModePerm)
}
