package main

import (
    "github.com/kumahq/kuma/tools/ci/marshalcheck"
    "golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
    singlechecker.Main(marshalcheck.Analyzer)
}
