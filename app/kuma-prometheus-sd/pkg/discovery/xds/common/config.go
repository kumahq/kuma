package common

type ApiVersion string

const (
	V1Alpha1 = ApiVersion("v1alpha1")
	V1       = ApiVersion("v1")
)

type DiscoveryConfig struct {
	ServerURL  string
	ClientName string
	ApiVersion ApiVersion
}
