package test

import (
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
)

// CustomResourceDir is the path from the top of the Kuma repository to
// the directory containing the Kuma CRD YAML files.
var CustomResourceDir = filepath.Join("deployments", "charts", "kuma", "crds")

// LabelBuildCheck this is for tests that check that the build is correct (some tests rely on build flags to be set)
var LabelBuildCheck = ginkgo.Label("build_check")
