package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	Product     = "Kuma"
	basedOnKuma = ""
	version     = "unknown"
	gitTag      = "unknown"
	gitCommit   = "unknown"
	buildDate   = "unknown"
	Envoy       = "unknown"
)

type BuildInfo struct {
	Product     string
	Version     string
	GitTag      string
	GitCommit   string
	BuildDate   string
	BasedOnKuma string
}

func (b BuildInfo) FormatDetailedProductInfo() string {
	base := []string{
		fmt.Sprintf("Product:       %s", b.Product),
		fmt.Sprintf("Version:       %s", b.Version),
		fmt.Sprintf("Git Tag:       %s", b.GitTag),
		fmt.Sprintf("Git Commit:    %s", b.GitCommit),
		fmt.Sprintf("Build Date:    %s", b.BuildDate),
	}
	if b.BasedOnKuma != "" {
		base = append(base, fmt.Sprintf("Based on Kuma: %s", b.BasedOnKuma))
	}
	return strings.Join(
		base,
		"\n",
	)
}

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

func init() {
	Build = BuildInfo{
		Product:     Product,
		Version:     version,
		GitTag:      gitTag,
		GitCommit:   gitCommit,
		BuildDate:   buildDate,
		BasedOnKuma: basedOnKuma,
	}
}
