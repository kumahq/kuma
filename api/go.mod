module github.com/kumahq/kuma/api

go 1.16

require (
	github.com/cncf/udpa v0.0.2-0.20201211205326-cc1b757b3edd // indirect
	github.com/envoyproxy/data-plane-api v0.0.0-20210105195927-01fb099f5a86 // indirect
	github.com/envoyproxy/go-control-plane v0.9.9-0.20201210154907-fd9021fe5dad
	github.com/envoyproxy/protoc-gen-validate v0.4.1
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.3
	github.com/googleapis/googleapis v0.0.0-20210422004218-6b711581c9a6 // indirect
	github.com/kumahq/protoc-gen-kumadoc v0.1.7
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.36.0
	google.golang.org/protobuf v1.25.0
// When running `make generate` in this folder, one can get into errors of missing proto dependecies
// To solve the issue, uncomment the section below and run `go mod download`
//github.com/cncf/udpa latest
//github.com/envoyproxy/data-plane-api latest
//github.com/googleapis/googleapis latest
)
