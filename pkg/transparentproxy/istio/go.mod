module github.com/kumahq/kuma/pkg/transparentproxy/istio

go 1.15

require (
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.6.3
	// fetched with go get istio.io/istio/tools/istio-iptables@1.7.8
	istio.io/istio v0.0.0-20210223230603-30e54dcb8a1c
)
