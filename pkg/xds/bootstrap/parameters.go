package bootstrap

import "time"

type configParameters struct {
	Id                 string
	Service            string
	AdminAddress       string
	AdminPort          uint32
	AdminAccessLogPath string
	XdsClusterType     string
	XdsHost            string
	XdsPort            uint32
	XdsUri             string
	XdsConnectTimeout  time.Duration
	AccessLogPipe      string
	DataplaneToken     string
	DataplaneResource  string
	CertBytes          string
	KumaDpVersion      string
	KumaDpGitTag       string
	KumaDpGitCommit    string
	KumaDpBuildDate    string
	EnvoyVersion       string
	EnvoyBuild         string
	HdsEnabled         bool
	DynamicMetadata    map[string]string
	DNSPort            uint32
	EmptyDNSPort       uint32
	ProxyType          string
}
