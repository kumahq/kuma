package bootstrap

import "time"

type KumaDpBootstrap struct {
	AggregateMetricsConfig map[string]AggregateMetricsConfig
}

type AggregateMetricsConfig struct {
	Path string
	Port uint32
}

type configParameters struct {
	Id                    string
	Service               string
	AdminAddress          string
	AdminPort             uint32
	AdminAccessLogPath    string
	XdsHost               string
	XdsPort               uint32
	XdsConnectTimeout     time.Duration
	AccessLogPipe         string
	DataplaneToken        string
	DataplaneResource     string
	CertBytes             []byte
	KumaDpVersion         string
	KumaDpGitTag          string
	KumaDpGitCommit       string
	KumaDpBuildDate       string
	EnvoyVersion          string
	EnvoyBuild            string
	EnvoyKumaDpCompatible bool
	HdsEnabled            bool
	DynamicMetadata       map[string]string
	DNSPort               uint32
	EmptyDNSPort          uint32
	ProxyType             string
}
