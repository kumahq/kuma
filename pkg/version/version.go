package version

var (
	version   = "unknown"
	gitTag    = "unknown"
	gitCommit = "unknown"
	buildDate = "unknown"
)

type BuildInfo struct {
	Version   string
	GitTag    string
	GitCommit string
	BuildDate string
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
