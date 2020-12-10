module github.com/kumahq/kuma

go 1.15

require (
	cirello.io/pglock v1.8.0
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/Nordix/simple-ipam v1.0.0
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/emicklei/go-restful v2.14.2+incompatible
	github.com/envoyproxy/go-control-plane v0.9.5
	github.com/envoyproxy/protoc-gen-validate v0.4.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-errors/errors v1.0.2-0.20180813162953-d98b870cc4e0
	github.com/go-kit/kit v0.10.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang-migrate/migrate/v4 v4.8.0
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/gruntwork-io/terratest v0.27.5
	github.com/hoisie/mustache v0.0.0-20160804235033-6375acf62c69
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kumahq/kuma/api v0.0.0-00010101000000-000000000000
	github.com/kumahq/kuma/pkg/plugins/resources/k8s/native v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.7.0
	github.com/miekg/dns v1.1.29
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.10.0
	github.com/prometheus/prometheus v0.0.0-00010101000000-000000000000
	github.com/sethvargo/go-retry v0.1.0
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/slok/go-http-metrics v0.9.0
	github.com/spf13/cobra v1.0.0
	github.com/spiffe/go-spiffe v0.0.0-20190820222348-6adcf1eecbcc
	github.com/spiffe/spire v0.10.0
	go.uber.org/multierr v1.3.0
	go.uber.org/zap v1.13.0
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.0.0-20201208233053-a543418bbed2 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.3.4
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v10.0.0+incompatible
	// migrating to v0.6.2 fails integration tests with:
	// failed to convert core list model of type SampleTrafficRoute into k8s counterpart
	sigs.k8s.io/controller-runtime v0.6.1
	sigs.k8s.io/testing_frameworks v0.1.1
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
	github.com/kumahq/kuma/api => ./api
	github.com/kumahq/kuma/pkg/plugins/resources/k8s/native => ./pkg/plugins/resources/k8s/native

	github.com/prometheus/prometheus => ./vendored/github.com/prometheus/prometheus
	github.com/spiffe/spire/proto/spire => github.com/spiffe/spire/proto/spire v0.10.0
	k8s.io/client-go => k8s.io/client-go v0.18.9
)
