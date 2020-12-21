module github.com/kumahq/kuma/pkg/transparentproxy/istio

go 1.15

require (
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.6.3
	// fetched with go get istio.io/istio/tools/istio-iptables@1.7.6
	istio.io/istio v0.0.0-20201207124053-74a8d16a8006
)
