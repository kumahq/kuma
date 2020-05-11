module github.com/Kong/kuma/pkg/plugins/resources/k8s/native

go 1.14

require (
	github.com/Kong/kuma/api v0.0.0-00010101000000-000000000000
	github.com/appscode/jsonpatch v1.0.1 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pkg/errors v0.8.1
	go.uber.org/multierr v1.1.0 // indirect
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	gomodules.xyz/jsonpatch/v2 v2.0.1 // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/controller-tools v0.2.1 // indirect
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
)

replace github.com/Kong/kuma/api => ../../../../../api
