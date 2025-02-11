package main

import (
    "github.com/kumahq/kuma/tools/ci/api-linter/linter"
    "golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
    singlechecker.Main(linter.Analyzer)
}
