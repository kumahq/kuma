module github.com/Kong/konvoy/components/konvoy-control-plane

go 1.12

require (
	github.com/envoyproxy/go-control-plane v0.8.0
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.2.1
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/spf13/cobra v0.0.4
	google.golang.org/grpc v1.19.1
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	sigs.k8s.io/controller-runtime v0.2.0-beta.1
)
