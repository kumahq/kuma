module github.com/kumahq/kuma/pkg/transparentproxy/istio

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1 // indirect
	github.com/spf13/viper v1.8.1
	// fetched with go get istio.io/istio/tools/istio-iptables@1.7.8
	istio.io/istio v0.0.0-20210223230603-30e54dcb8a1c
)
