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
	github.com/envoyproxy/go-control-plane v0.9.9-0.20210310174751-14ce79b5761d
	github.com/envoyproxy/protoc-gen-validate v0.4.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-errors/errors v1.0.2-0.20180813162953-d98b870cc4e0
	github.com/go-kit/kit v0.10.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/gruntwork-io/terratest v0.30.15
	github.com/hoisie/mustache v0.0.0-20160804235033-6375acf62c69
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kumahq/kuma/api v0.0.0-00010101000000-000000000000
	github.com/kumahq/kuma/pkg/plugins/resources/k8s/native v0.0.0-00010101000000-000000000000
	github.com/kumahq/kuma/pkg/transparentproxy/istio v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.2
	github.com/miekg/dns v1.1.42
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.29.0
	github.com/prometheus/prometheus v0.0.0-00010101000000-000000000000
	github.com/sethvargo/go-retry v0.1.0
	github.com/slok/go-http-metrics v0.9.0
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.1.3
	github.com/spiffe/go-spiffe v0.0.0-20190820222348-6adcf1eecbcc
	github.com/spiffe/spire v0.12.1
	github.com/spiffe/spire/proto/spire v0.12.0 // indirect
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.13.0
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/genproto v0.0.0-20201030142918-24207fddd1c3
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.27.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.3.4
	k8s.io/api v0.18.14
	k8s.io/apiextensions-apiserver v0.18.14
	k8s.io/apimachinery v0.18.14
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451
	sigs.k8s.io/controller-runtime v0.6.4
	sigs.k8s.io/testing_frameworks v0.1.2
)

replace (
	github.com/kumahq/kuma/api => ./api
	github.com/kumahq/kuma/pkg/plugins/resources/k8s/native => ./pkg/plugins/resources/k8s/native
	github.com/kumahq/kuma/pkg/transparentproxy/istio => ./pkg/transparentproxy/istio

	github.com/prometheus/prometheus => ./vendored/github.com/prometheus/prometheus
	k8s.io/client-go => k8s.io/client-go v0.18.14
)
