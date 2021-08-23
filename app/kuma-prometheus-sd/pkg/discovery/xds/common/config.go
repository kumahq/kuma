package common

type ApiVersion string

const (
	V1 = ApiVersion("v1")
)

type DiscoveryConfig struct {
	ServerURL  string
	ClientName string
	ApiVersion ApiVersion
}
