module github.com/Kong/kuma/pkg/test/xds/client

go 1.15

require (
	github.com/envoyproxy/go-control-plane v0.9.7
	github.com/golang/protobuf v1.4.2
	github.com/kumahq/kuma v0.0.0-20201012122130-3a28be165906
	github.com/kumahq/kuma/api v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	go.uber.org/multierr v1.6.0
	google.golang.org/genproto v0.0.0-20201009135657-4d944d34d83c
	google.golang.org/grpc v1.32.0
)

replace (
	github.com/kumahq/kuma/api => ../../../../api
	github.com/prometheus/prometheus => ../../../../vendored/github.com/prometheus/prometheus
	k8s.io/client-go => k8s.io/client-go v0.18.14
	github.com/kumahq/kuma/pkg/plugins/resources/k8s/native => ../../../plugins/resources/k8s/native
)
