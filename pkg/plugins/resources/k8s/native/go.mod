module github.com/kumahq/kuma/pkg/plugins/resources/k8s/native

go 1.16

require (
	github.com/go-logr/logr v0.1.0
	github.com/golang/protobuf v1.5.2
	github.com/kumahq/kuma/api v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	k8s.io/apimachinery v0.18.14
	k8s.io/client-go v0.18.14
	sigs.k8s.io/controller-runtime v0.6.4
)

replace github.com/kumahq/kuma/api => ../../../../../api
