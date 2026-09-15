package config

import (
	"fmt"
	"io"
	"os"
)

type Deprecation struct {
	Env             string
	EnvMsg          string
	ConfigValuePath func(cfg Config) (string, bool)
	ConfigValueMsg  string
}

func PrintDeprecations(deprecations []Deprecation, cfg Config, out io.Writer) {
	for _, d := range deprecations {
		if _, ok := os.LookupEnv(d.Env); ok {
			_, _ = fmt.Fprintf(out, "Deprecated: %v. %v\n", d.Env, d.EnvMsg)
		}
		if path, exist := d.ConfigValuePath(cfg); exist {
			_, _ = fmt.Fprintf(out, "Deprecated: %v. %v\n", path, d.ConfigValueMsg)
		}
	}
}
