package version

import (
	"fmt"
	"strings"
)

var (
	Product   = "Kuma"
	version   = "unknown"
	gitTag    = "unknown"
	gitCommit = "unknown"
	buildDate = "unknown"
	Envoy     = "unknown"
)

type BuildInfo struct {
	Version   string
	GitTag    string
	GitCommit string
	BuildDate string
}

func FormatDetailedProductInfo() string {
	return strings.Join(
		[]string{
			fmt.Sprintf("Product:    %s", Product),
			fmt.Sprintf("Version:    %s", Build.Version),
			fmt.Sprintf("Git Tag:    %s", Build.GitTag),
			fmt.Sprintf("Git Commit: %s", Build.GitCommit),
			fmt.Sprintf("Build Date: %s", Build.BuildDate),
		},
		"\n",
	)
}

var (
	Build BuildInfo
)

func init() {
	Build = BuildInfo{
		Version:   version,
		GitTag:    gitTag,
		GitCommit: gitCommit,
		BuildDate: buildDate,
	}
}
