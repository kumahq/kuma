module github.com/Kong/konvoy/components/konvoy-control-plane

go 1.12

require (
	github.com/Kong/konvoy/components/konvoy-control-plane/model v0.0.0-00010101000000-000000000000
	github.com/envoyproxy/go-control-plane v0.8.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/spf13/cobra v0.0.4
	golang.org/x/crypto v0.0.0-20190617133340-57b3e21c3d56 // indirect
	golang.org/x/net v0.0.0-20190613194153-d28f0bde5980 // indirect
	golang.org/x/sys v0.0.0-20190618155005-516e3c20635f // indirect
	golang.org/x/tools v0.0.0-20190618163018-fdf1049a943a // indirect
	google.golang.org/grpc v1.19.1
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
	sigs.k8s.io/testing_frameworks v0.1.1
)

replace github.com/Kong/konvoy/components/konvoy-control-plane/model => ./model
