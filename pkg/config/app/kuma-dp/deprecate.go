package kumadp

import (
	"io"

	"github.com/kumahq/kuma/pkg/config"
)

var deprecations = []config.Deprecation{}

func PrintDeprecations(cfg *Config, out io.Writer) {
	config.PrintDeprecations(deprecations, cfg, out)
}
