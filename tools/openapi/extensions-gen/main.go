// Command extensions-gen documents the registered resource extensions in an
// OpenAPI document.
//
// Every product keeps its own copy of this command, because what it generates is
// decided by which extension packages the binary imports. Kuma ships no extensions
// of its own, so this build registers nothing and leaves the document as it found
// it — importing a package here is what puts an extension in the spec.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kumahq/kuma/v3/tools/openapi/extensions"
)

func main() {
	var opts extensions.Options
	flag.StringVar(&opts.Spec, "spec", "", "path to the OpenAPI document to patch in place")
	flag.StringVar(&opts.WorkDir, "work-dir", "build/openapi-extensions",
		"directory, relative to the module root, for the throwaway package controller-gen reads")
	flag.StringVar(&opts.ControllerGenBin, "controller-gen-bin", "controller-gen", "path to a controller-gen binary")
	flag.StringVar(&opts.YqBin, "yq-bin", "yq", "path to a yq binary")
	flag.Parse()

	opts.Stderr = os.Stderr

	if opts.Spec == "" {
		fmt.Fprintln(os.Stderr, "--spec is required")
		os.Exit(1)
	}
	if err := extensions.Generate(context.Background(), opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
