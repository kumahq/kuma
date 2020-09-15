module github.com/kumahq/kuma/pkg/plugins/resources/k8s/native

go 1.14

require (
	github.com/go-logr/logr v0.1.0
	github.com/golang/protobuf v1.3.5
	github.com/kumahq/kuma/api v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.8.1
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.0
)

replace github.com/kumahq/kuma/api => ../../../../../api
