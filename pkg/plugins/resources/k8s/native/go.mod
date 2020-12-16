module github.com/kumahq/kuma/pkg/plugins/resources/k8s/native

go 1.15

require (
	github.com/go-logr/logr v0.2.0
	github.com/golang/protobuf v1.4.3
	github.com/kumahq/kuma/api v0.0.0-00010101000000-000000000000 // indirect
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.18.9
	sigs.k8s.io/controller-runtime v0.6.1
)

replace github.com/kumahq/kuma/api => ../../../../../api
