package main

import (
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	kuma_cp "github.com/kumahq/kuma/app/kuma-cp/cmd"
	kuma_dp "github.com/kumahq/kuma/app/kuma-dp/cmd"
	kuma_prometheus_sd "github.com/kumahq/kuma/app/kuma-prometheus-sd/cmd"
	kumactl "github.com/kumahq/kuma/app/kumactl/cmd"
)

// GetenvOr returns the value of the environment variable named by env,
// or the value of alternate if the environment variable is not present.
//
// TODO(jpeach) move this to pkg/utils
func GetenvOr(env string, alternate string) string {
	if value, ok := os.LookupEnv(env); ok {
		return value
	}

	return alternate
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func markdown(path string, cmd *cobra.Command) {
	must(os.MkdirAll(path, 0755))
	must(doc.GenMarkdownTree(cmd, path))
}

func main() {
	prefix := GetenvOr("DESTDIR", ".")
	format := GetenvOr("FORMAT", "markdown")

	apps := map[string]*cobra.Command{
		path.Join(prefix, "kuma-cp"):            kuma_cp.DefaultRootCmd(),
		path.Join(prefix, "kumactl"):            kumactl.DefaultRootCmd(),
		path.Join(prefix, "kuma-dp"):            kuma_dp.DefaultRootCmd(),
		path.Join(prefix, "kuma-prometheus-sd"): kuma_prometheus_sd.DefaultRootCmd(),
	}

	switch format {
	case "markdown":
		for p, c := range apps {
			markdown(p, c)
		}
	default:
		log.Fatalf("unsupported reference format %q", format)
	}
}
