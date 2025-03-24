package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Spec struct {
	Description string   `json:"Description"`
	Paths       []string `json:"Paths"`
}

func parseInput(input *os.File) map[string][]Spec {
	result := map[string][]Spec{}

	var data []struct {
		SuiteDescription string
		SuitePath        string
		SpecReports      []struct {
			State                       string
			ContainerHierarchyTexts     []string
			LeafNodeText                string
			ContainerHierarchyLocations []struct {
				FileName   string
				LineNumber int
			}
			LeafNodeLocation struct {
				FileName   string
				LineNumber int
			}
		}
	}

	if err := json.NewDecoder(input).Decode(&data); err != nil {
		log.Fatal(err)
	}

	for _, d := range data {
		var specs []Spec
		for _, s := range d.SpecReports {
			if s.State == "passed" {
				continue
			}
			var paths []string
			for _, loc := range s.ContainerHierarchyLocations {
				paths = append(paths, fmt.Sprintf("%s:%d", trimPath(loc.FileName), loc.LineNumber))
			}
			paths = append(paths, fmt.Sprintf("%s:%d", trimPath(s.LeafNodeLocation.FileName), s.LeafNodeLocation.LineNumber))
			specs = append(specs, Spec{
				Description: strings.Join(append(s.ContainerHierarchyTexts, s.LeafNodeText), " "),
				Paths:       paths,
			})
		}
		if len(specs) > 0 {
			key := fmt.Sprintf("%s (%s)", d.SuiteDescription, trimPath(d.SuitePath))
			result[key] = specs
		}
	}
	return result
}

func trimPath(path string) string {
	parts := strings.SplitN(path, "github.com/", 2)
	if len(parts) == 2 && strings.Contains(parts[1], "/") {
		return strings.SplitN(parts[1], "/", 3)[2]
	}
	return path
}

func prettyPrint(data map[string][]Spec, noColor bool) {
	for suite, specs := range data {
		if noColor {
			fmt.Println(suite)
		} else {
			color.Cyan(suite)
		}
		fmt.Println()

		for _, spec := range specs {
			fmt.Println("-", spec.Description)
			for _, path := range spec.Paths {
				if noColor {
					fmt.Printf("  • %s\n", path)
				} else {
					color.Yellow("  • %s", path)
				}
			}
			fmt.Println()
		}
	}
}

func main() {
	inputFile := flag.String("input-file", "", "Input JSON file (default stdin)")
	noColor := flag.Bool("no-color", false, "Disable colored output")
	flag.Parse()

	if *noColor {
		color.NoColor = true
	}

	input := os.Stdin
	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		input = f
	}

	data := parseInput(input)
	prettyPrint(data, *noColor)
}
