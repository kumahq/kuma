package bootstrap

import "time"

type configParameters struct {
	Id                 string
	Service            string
	AdminAddress       string
	AdminPort          uint32
	AdminAccessLogPath string
	XdsHost            string
	XdsPort            uint32
	XdsConnectTimeout  time.Duration
	AccessLogPipe      string
	DataplaneTokenPath string
	DataplaneResource  string
	CertBytes          string
	KumaDpVersion      string
	KumaDpGitTag       string
	KumaDpGitCommit    string
	KumaDpBuildDate    string
	EnvoyVersion       string
	EnvoyBuild         string
}
