package test

import "path/filepath"

// CustomResourceDir is the path from the top of the Kuma repository to
// the directory containing the Kuma CRD YAML files.
var CustomResourceDir = filepath.Join("deployments", "charts", "kuma", "crds")
