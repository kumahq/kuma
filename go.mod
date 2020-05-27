module github.com/Kong/kuma

go 1.14

require (
	github.com/Kong/kuma/api v0.0.0-00010101000000-000000000000
	github.com/Kong/kuma/pkg/plugins/resources/k8s/native v0.0.0-00010101000000-000000000000
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/emicklei/go-restful v2.9.6+incompatible
	github.com/envoyproxy/go-control-plane v0.9.5
	github.com/envoyproxy/protoc-gen-validate v0.3.0-java.0.20200311152155-ab56c3dd1cf9
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/golang-migrate/migrate/v4 v4.8.0
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/mux v1.7.2 // indirect
	github.com/hoisie/mustache v0.0.0-20160804235033-6375acf62c69
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.1.1
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.9.1
	github.com/prometheus/prometheus v0.0.0-00010101000000-000000000000
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spiffe/go-spiffe v0.0.0-20190820222348-6adcf1eecbcc
	github.com/spiffe/spire v0.10.0
	go.uber.org/multierr v1.1.0
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20200406173513-056763e48d71 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4 // indirect
	golang.org/x/tools v0.0.0-20200520220537-cf2d1e09c845 // indirect
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.26.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/fsnotify.v1 v1.4.9 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/helm v2.14.3+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/testing_frameworks v0.1.1
)

replace (
	github.com/Kong/kuma/api => ./api
	github.com/Kong/kuma/pkg/plugins/resources/k8s/native => ./pkg/plugins/resources/k8s/native

	github.com/prometheus/prometheus => ./vendor/github.com/prometheus/prometheus
	github.com/spiffe/spire/proto/spire => github.com/spiffe/spire/proto/spire v0.10.0
	gopkg.in/fsnotify.v1 => gopkg.in/fsnotify.v1 v1.4.7
	k8s.io/client-go => k8s.io/client-go v0.18.2
)
