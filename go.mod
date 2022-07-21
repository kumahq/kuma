module github.com/kumahq/kuma

go 1.18

require (
	cirello.io/pglock v1.9.0
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Nordix/simple-ipam v1.0.0
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/cncf/xds/go v0.0.0-20211130200136-a8f946100490
	github.com/containernetworking/cni v0.8.1
	github.com/containernetworking/plugins v0.9.1
	github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/envoyproxy/go-control-plane v0.10.2-0.20220325020618-49ff273808a1
	github.com/envoyproxy/protoc-gen-validate v0.6.7
	github.com/evanphx/json-patch/v5 v5.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	github.com/golang-jwt/jwt/v4 v4.4.2
	github.com/golang-migrate/migrate/v4 v4.15.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/gruntwork-io/terratest v0.40.18
	github.com/hoisie/mustache v0.0.0-20160804235033-6375acf62c69
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kumahq/kuma-net v0.3.0
	github.com/kumahq/protoc-gen-kumadoc v0.2.0
	github.com/lib/pq v1.10.6
	github.com/miekg/dns v1.1.50
	github.com/natefinch/atomic v1.0.1
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.19.0
	github.com/operator-framework/operator-lib v0.11.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.2
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.37.0
	github.com/prometheus/prometheus v0.0.0-00010101000000-000000000000
	github.com/sethvargo/go-retry v0.2.3
	github.com/slok/go-http-metrics v0.10.0
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.12.0
	github.com/spiffe/go-spiffe v0.0.0-20190820222348-6adcf1eecbcc
	github.com/testcontainers/testcontainers-go v0.13.0
	go.uber.org/multierr v1.8.0
	go.uber.org/zap v1.21.0
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.8.1
	istio.io/pkg v0.0.0-20201202160453-b7f8c8c88ca3
	k8s.io/api v0.24.3
	k8s.io/apiextensions-apiserver v0.24.3
	k8s.io/apimachinery v0.24.3
	k8s.io/client-go v0.24.3
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	sigs.k8s.io/controller-runtime v0.12.3
	sigs.k8s.io/controller-tools v0.9.2
	sigs.k8s.io/gateway-api v0.5.0
	sigs.k8s.io/testing_frameworks v0.1.2
)

require (
	cloud.google.com/go/compute v1.6.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.21 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.16 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/BurntSushi/toml v0.4.1 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/aws/aws-sdk-go v1.40.56 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/cgroups v1.0.3 // indirect
	github.com/containerd/containerd v1.5.13 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.13+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-errors/errors v1.0.2-0.20180813162953-d98b870cc4e0 // indirect
	github.com/go-kit/kit v0.11.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gobuffalo/flect v0.2.5 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gruntwork-io/go-commons v0.8.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.1-0.20190430135223-99e2f22d1c94 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-zglob v0.0.2-0.20190814121620-e3c945676326 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.5.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pquerna/otp v1.2.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.1 // indirect
	github.com/subosito/gotenv v1.3.0 // indirect
	github.com/urfave/cli v1.22.2 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/exp v0.0.0-20220407100705-7b9b53b0aca4 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20220517211312-f3a8303e98df // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.1.0 // indirect
	k8s.io/component-base v0.24.3 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220328201542-3ee0da9b0b42 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace github.com/prometheus/prometheus => ./vendored/github.com/prometheus/prometheus

replace sigs.k8s.io/gateway-api => github.com/kumahq/gateway-api v0.0.0-20220714082056-fbb05ce01577

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
//	$ go mod edit -replace github.com/envoyproxy/go-control-plane=github.com/kumahq/go-control-plane@channel-increased-capacity
//	$ GOPRIVATE=github.com/kumahq/go-control-plane go mod tidy
replace github.com/envoyproxy/go-control-plane => github.com/kumahq/go-control-plane v0.9.10-0.20211022075049-d35edcf0813a
