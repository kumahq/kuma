package kuma_cp

import (
	"io"

	"github.com/kumahq/kuma/v3/pkg/config"
)

var deprecations = []config.Deprecation{
	{
		Env:    "KUMA_STORE_CACHE_ENABLED",
		EnvMsg: "the resource store cache can no longer be disabled, the setting is ignored and will be removed. Only the expiration time (KUMA_STORE_CACHE_EXPIRATION_TIME) is configurable.",
	},
}

func PrintDeprecations(cfg *Config, out io.Writer) {
	config.PrintDeprecations(deprecations, cfg, out)
}
