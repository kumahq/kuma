module github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native

go 1.12

require (
	github.com/Kong/konvoy/components/konvoy-control-plane/api v0.0.0-00010101000000-000000000000
	github.com/envoyproxy/go-control-plane v0.8.2 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.2.1
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
	sigs.k8s.io/controller-tools v0.2.0-beta.2 // indirect
)

replace github.com/Kong/konvoy/components/konvoy-control-plane/api => ../../../../../api
