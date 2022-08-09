package bootstrap

import (
	"time"

	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

type KumaDpBootstrap struct {
	AggregateMetricsConfig []AggregateMetricsConfig
}

type AggregateMetricsConfig struct {
	Name    string
	Path    string
	Address string
	Port    uint32
	IsIPv6  bool
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
	DataplaneTokenPath    string
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
	Features              []string
	IsGatewayDataplane    bool
	Resources             types.ProxyResources
}
