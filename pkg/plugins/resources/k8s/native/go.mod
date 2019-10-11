module github.com/Kong/kuma/pkg/plugins/resources/k8s/native

go 1.12

require (
	github.com/Kong/kuma/api v0.0.0-00010101000000-000000000000
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.2.2-0.20190730201129-28a6bbf47e48
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.2
	sigs.k8s.io/controller-tools v0.2.1 // indirect
)

replace github.com/Kong/kuma/api => ../../../../../api
