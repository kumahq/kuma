module github.com/kumahq/kuma/api

go 1.14

require (
	github.com/envoyproxy/go-control-plane v0.9.5
	github.com/envoyproxy/protoc-gen-validate v0.3.0-java.0.20200311152155-ab56c3dd1cf9 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.3.5
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.8.1
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859 // indirect
	google.golang.org/grpc v1.30.0 // indirect
// When running `make generate` in this folder, one can get into errors of missing proto dependecies
// To solve the issue, uncomment the section below and run `go mod download`
//github.com/cncf/udpa latest
//github.com/envoyproxy/data-plane-api latest
//github.com/googleapis/googleapis latest
)
