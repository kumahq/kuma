package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	kuma_cp "github.com/kumahq/kuma/app/kuma-cp/cmd"
	kuma_dp "github.com/kumahq/kuma/app/kuma-dp/cmd"
	kumactl "github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/pkg/version"
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

func disableAutogen(cmd *cobra.Command) *cobra.Command {
	// Not that we don't want to advertise cobra, but this is the only way to
	// suppress the timestamp.
	cmd.DisableAutoGenTag = true

	for _, c := range cmd.Commands() {
		disableAutogen(c)
	}

	return cmd
}

func markdown(path string, cmd *cobra.Command) {
	must(os.MkdirAll(path, 0o755))
	must(doc.GenMarkdownTree(cmd, path))
}

func man(path string, header *doc.GenManHeader, cmd *cobra.Command) {
	must(os.MkdirAll(path, 0o755))
	must(doc.GenManTree(cmd, header, path))
}

type command struct {
	command *cobra.Command
	header  *doc.GenManHeader
}

func main() {
	prefix := GetenvOr("DESTDIR", ".")
	format := GetenvOr("FORMAT", "markdown")

	apps := map[string]command{
		path.Join(prefix, "kuma-cp"): {
			command: kuma_cp.DefaultRootCmd(),
			header: &doc.GenManHeader{
				Title:   "KUMA-CP",
				Section: "8",
				Source:  fmt.Sprintf("%s %s", version.Product, version.Build.Version),
				Manual:  version.Product,
			},
		},
		path.Join(prefix, "kuma-dp"): {
			command: kuma_dp.DefaultRootCmd(),
			header: &doc.GenManHeader{
				Title:   "KUMA-DP",
				Section: "8",
				Source:  fmt.Sprintf("%s %s", version.Product, version.Build.Version),
				Manual:  version.Product,
			},
		},
		path.Join(prefix, "kumactl"): {
			command: kumactl.DefaultRootCmd(),
			header: &doc.GenManHeader{
				Title:   "KUMACTL",
				Section: "1",
				Source:  fmt.Sprintf("%s %s", version.Product, version.Build.Version),
				Manual:  version.Product,
			},
		},
	}

	switch format {
	case "markdown":
		for p, c := range apps {
			markdown(p, disableAutogen(c.command))
		}
	case "man":
		for p, c := range apps {
			man(p, c.header, disableAutogen(c.command))
		}
	default:
		log.Fatalf("unsupported reference format %q", format)
	}
}
