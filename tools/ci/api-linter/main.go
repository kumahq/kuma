package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/kumahq/kuma/v2/tools/ci/api-linter/linter"
)

func main() {
	singlechecker.Main(linter.Analyzer)
}
