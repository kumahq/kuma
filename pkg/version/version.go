package version

import (
	"fmt"
	"runtime"
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

<<<<<<< HEAD
var (
	Build BuildInfo
)
=======
func (b BuildInfo) UserAgent(component string) string {
	return fmt.Sprintf("%s/%s (%s; %s; %s/%s)",
		component,
		b.Version,
		runtime.GOOS,
		runtime.GOARCH,
		b.Product,
		b.GitCommit[:7])
}

var Build BuildInfo
>>>>>>> 01b999035 (feat(kds): add user-agent with useful version info (#7886))

func init() {
	Build = BuildInfo{
		Version:   version,
		GitTag:    gitTag,
		GitCommit: gitCommit,
		BuildDate: buildDate,
	}
}
