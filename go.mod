module github.com/kumahq/kuma

go 1.16

require (
	cirello.io/pglock v1.8.0
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Nordix/simple-ipam v1.0.0
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/envoyproxy/go-control-plane v0.9.9-0.20210512163311-63b5d3c536b0
	github.com/envoyproxy/protoc-gen-validate v0.4.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.1.1
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/gruntwork-io/terratest v0.30.15
	github.com/hoisie/mustache v0.0.0-20160804235033-6375acf62c69
	github.com/iancoleman/orderedmap v0.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kumahq/kuma/pkg/transparentproxy/istio v0.0.0-00010101000000-000000000000
	github.com/kumahq/protoc-gen-kumadoc v0.1.7
	github.com/lib/pq v1.10.3
	github.com/miekg/dns v1.1.43
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.30.0
	github.com/prometheus/prometheus v0.0.0-00010101000000-000000000000
	github.com/sethvargo/go-retry v0.1.0
	github.com/slok/go-http-metrics v0.9.0
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.2.1
	github.com/spiffe/go-spiffe v0.0.0-20190820222348-6adcf1eecbcc
	github.com/spiffe/spire v0.12.3
	github.com/spiffe/spire/proto/spire v0.12.0 // indirect
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.3.4
	k8s.io/api v0.18.14
	k8s.io/apiextensions-apiserver v0.18.14
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.18.14
	k8s.io/utils v0.0.0-20210709001253-0e1f9d693477
	sigs.k8s.io/controller-runtime v0.6.4
	sigs.k8s.io/testing_frameworks v0.1.2
)

replace (
	github.com/kumahq/kuma/pkg/transparentproxy/istio => ./pkg/transparentproxy/istio
	github.com/prometheus/prometheus => ./vendored/github.com/prometheus/prometheus
)

// The following replacement refers to the kuma-release-1.3 branch.
//
// There are a few Go module traps to be aware of when dealing with
// this replacement:
//
//	https://github.com/golang/go/issues/32955
//	https://github.com/golang/go/issues/45413
//
// To force Go tooling to update the Git hash of the branch you need to
// work around the module caching system by doing this:
//
//	$ go mod edit -replace github.com/envoyproxy/go-control-plane=github.com/kumahq/go-control-plane@kuma-release-1.3
//	$ GOPRIVATE=github.com/kumahq/go-control-plane go mod tidy
replace github.com/envoyproxy/go-control-plane => github.com/kumahq/go-control-plane v0.9.9-0.20210914001841-ec3541a22836
