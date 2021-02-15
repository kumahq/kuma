module github.com/kumahq/kuma/pkg/plugins/resources/k8s/native

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/golang/protobuf v1.4.3
	github.com/kumahq/kuma/api v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.2
)

replace github.com/kumahq/kuma/api => ../../../../../api
