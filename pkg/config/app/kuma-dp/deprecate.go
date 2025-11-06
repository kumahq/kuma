package kumadp

import (
	"io"

	"github.com/kumahq/kuma/v2/pkg/config"
)

var deprecations = []config.Deprecation{}

func PrintDeprecations(cfg *Config, out io.Writer) {
	config.PrintDeprecations(deprecations, cfg, out)
}
