# CHANGELOG

## [1.5.0]
> Released on 2022/02/21

Changes:
* feat(*): zone egress [#3809](https://github.com//kumahq/kuma/pull/3809) [#3757](https://github.com//kumahq/kuma/pull/3757)
* feat(kuma-cp) data plane proxy membership [#3619](https://github.com//kumahq/kuma/pull/3619)
* feat(kuma-cp): reachable services in transparent proxying [#3791](https://github.com//kumahq/kuma/pull/3791)
* feat(inspect-api): retrieve full XDS config [#3768](https://github.com//kumahq/kuma/pull/3768)
* feat(*): inspect api support [#3805](https://github.com//kumahq/kuma/pull/3805) [#3568](https://github.com//kumahq/kuma/pull/3568) [#3462](https://github.com//kumahq/kuma/pull/3462)
* feat(kuma-cp): add proxytemplate to matched policies for inspect poli… [#3786](https://github.com//kumahq/kuma/pull/3786) 👍contributed by @tharun208
* feat(kuma-cp): enable traffic route for inspect endpoints [#3735](https://github.com//kumahq/kuma/pull/3735) 👍contributed by @tharun208
* feat(*): move adminPort to DPP resource [#3739](https://github.com//kumahq/kuma/pull/3739)
* feat(helm): add imagePullSecrets support [#3755](https://github.com//kumahq/kuma/pull/3755) 👍contributed by @johnharris85
* feat(*): enable Gateway with runtime flag [#3736](https://github.com//kumahq/kuma/pull/3736)
* feat(kumactl): add --api-timeout flag [#3723](https://github.com//kumahq/kuma/pull/3723)
* feat: allow for ca/identity secrets for every mesh [#3696](https://github.com//kumahq/kuma/pull/3696)
* feat(kuma-cp): allow extra cm in kuma cp chart [#3671](https://github.com//kumahq/kuma/pull/3671) 👍contributed by @wjrbetts
* feat(kuma-cp): add gui link in index api response [#3675](https://github.com//kumahq/kuma/pull/3675) 👍contributed by @tharun208
* feat(*): allow ca.crt to be in separate k8s secret [#3638](https://github.com//kumahq/kuma/pull/3638)
* feat(kumactl): add type of logging and tracing backends with name in table output [#3636](https://github.com//kumahq/kuma/pull/3636) 👍contributed by @tharun208
* feat(kuma-cp): enable client side gRPC keepalive [#3574](https://github.com//kumahq/kuma/pull/3574)
* feat(gui): new onboarding view [kumahq/kuma-gui#194](https://github.com/kumahq/kuma-gui/pull/194)
* feat(gui): link to documentation from policy view [kumahq/kuma-gui#289](https://github.com/kumahq/kuma-gui/pull/289)

* fix(kuma-cp): do not update unchanged insights [#3819](https://github.com//kumahq/kuma/pull/3819)
* fix(*): do not annotate gateway services with ingress upstream [#3816](https://github.com//kumahq/kuma/pull/3816)
* fix(*): properly escape DB password when creating postgres connection string [#3804](https://github.com//kumahq/kuma/pull/3804)
* fix(kuma-cp): fix missing label sidecar injection [#3740](https://github.com//kumahq/kuma/pull/3740)
* fix(kuma-dp): fix conntrack collisions  [#3459](https://github.com//kumahq/kuma/pull/3459) 👍contributed by @johnharris85
* fix(conf): remove invalid health check fields from example [#3697](https://github.com//kumahq/kuma/pull/3697) 👍contributed by @tharun208
* fix(kuma-dp): binary lookup function skips not available directories [#3667](https://github.com//kumahq/kuma/pull/3667)
* fix(k8s): make sure controllers start after leader election [#3666](https://github.com//kumahq/kuma/pull/3666)
* fix(build): fix gomega matchers for inspect resources command test [#3660](https://github.com//kumahq/kuma/pull/3660) [#3651](https://github.com//kumahq/kuma/pull/3651) 👍contributed by @tharun208
* fix(kumactl): ignore any unregistered CRDs, not only from the root chart [#3643](https://github.com//kumahq/kuma/pull/3643)
* fix(kumactl): print meta before spec for Kuma resources [#3637](https://github.com//kumahq/kuma/pull/3637)
* fix(kuma-cp): add cp selector to global sync service [#3579](https://github.com//kumahq/kuma/pull/3579)
* fix(kuma-cp) do not override other dataplane with dp lifecycle [#3507](https://github.com//kumahq/kuma/pull/3507)
* fix(helm) Add support to customize nodeport [#1944](https://github.com//kumahq/kuma/pull/1944) 👍contributed by @bhiravabhatla

* perf(kuma-cp): use mesh snapshot in proxy builder [#3700](https://github.com//kumahq/kuma/pull/3700)
* perf(kuma-cp): use mesh snapshot in gateway [#3710](https://github.com//kumahq/kuma/pull/3710)
* perf(kuma-cp): share mesh context [#3659](https://github.com//kumahq/kuma/pull/3659)

* improvement(metadata): include name of annotation to parse error message [#3677](https://github.com//kumahq/kuma/pull/3677) 👍contributed by @ChinYing-Li
* refactor(insights): delete method GetLatestSubscription for insights [#3656](https://github.com//kumahq/kuma/pull/3656) 👍contributed by @tharun208
* refactor(kuma-cp): unify mesh determination for k8s objects [#3708](https://github.com//kumahq/kuma/pull/3708)
* refactor(*): replace ensureDefaultXXX functions with a single generic function [#3662](https://github.com//kumahq/kuma/pull/3662) 👍contributed by @tharun208
* chore(zone-ingress): delete deprecated env KUMA_DATAPLANE_ADMIN_PORT [#3766](https://github.com//kumahq/kuma/pull/3766)
* chore(k8s): remove GetBool method and use GetEnabled [#3698](https://github.com//kumahq/kuma/pull/3698) 👍contributed by @tharun208
* chore(*): generate CRD types [#3453](https://github.com//kumahq/kuma/pull/3453)
* chore(dataplane)!: disallow using 0.0.0.0 in networking.address for dp [#3691](https://github.com//kumahq/kuma/pull/3691)
* chore(kuma-cp): consolidate mesh defaults creation [#3678](https://github.com//kumahq/kuma/pull/3678)
* chore(config): remove ability to disable insights [#3501](https://github.com//kumahq/kuma/pull/3501)
* chore(*): remove old Ingress [#3435](https://github.com//kumahq/kuma/pull/3435)
* chore(*): upgrade Envoy to v1.21.1 [#3909](https://github.com//kumahq/kuma/pull/3909)
* chore(grafana): update to latest grafana plugin version [#3812](https://github.com//kumahq/kuma/pull/3812)
* ci(*): release on every commit in master and release branches [#3712](https://github.com//kumahq/kuma/pull/3712)

## [1.4.1]
> Released on 2021/12/15

Changes:
* feat: add kubernetes tags automatically [#3439](https://github.com//kumahq/kuma/pull/3439)
* perf: update Mesh and ServiceInsights only when really needed [#3463](https://github.com//kumahq/kuma/pull/3463)
* perf: eliminate uneccessary JSON marshalling [#3483](https://github.com//kumahq/kuma/pull/3483)
* feat: sidecar injection webhook based on labels [#3417](https://github.com//kumahq/kuma/pull/3417)
* chore: upgrade gui to new version [#3454](https://github.com//kumahq/kuma/pull/3454)
* test: fix postgress tests permissions [#3443](https://github.com//kumahq/kuma/pull/3443)
* feat: add affinity to CP and Ingress pods [#3036](https://github.com//kumahq/kuma/pull/3036)
  👍contributed by @andrey-dubnik
* chore: bump github.com/golang-jwt/jwt/v4 from 4.1.0 to 4.2.0 [#3432](https://github.com//kumahq/kuma/pull/3432)
* feat: consolidate tokens logic to support expiration, rotation, revocation and RSA256 [#3376](https://github.com/kumahq/kuma/pull/3376)
* fix: simplify cluster creation with endpoints [#3403](https://github.com//kumahq/kuma/pull/3403)
* fix: enable metrics hijacker for current version of Kuma [#3405](https://github.com//kumahq/kuma/pull/3405)
* fix: switch to mTLS when CP communicates with Envoy Admin [#3353](https://github.com//kumahq/kuma/pull/3353)
* chore: bump github.com/spiffe/spire from 0.12.3 to 1.1.1 [#3388](https://github.com//kumahq/kuma/pull/3388)
* chore: bump github.com/spf13/viper from 1.8.1 to 1.9.0 [#3389](https://github.com//kumahq/kuma/pull/3389)
* fix: validate cp url in dp conf [#3357](https://github.com//kumahq/kuma/pull/3357)
* chore: send reports to tls endpoint [#3361](https://github.com//kumahq/kuma/pull/3361)
* chore: check explicit service account name [#3228](https://github.com//kumahq/kuma/pull/3228)
* feat: inspect other dependencies versions [#3352](https://github.com//kumahq/kuma/pull/3352)
* chore: add area/gateway label [#3263](https://github.com//kumahq/kuma/pull/3263)
* chore: remove dp token from xds metadata [#3282](https://github.com//kumahq/kuma/pull/3282)
* refactor: move from io/ioutil to io and os packages [#3265](https://github.com//kumahq/kuma/pull/3265)
  👍contributed by @Juneezee
* fix: validate newly generated xDS snapshots [#3195](https://github.com//kumahq/kuma/pull/3195)
* chore: bump k8s.io/apiextensions-apiserver from 0.22.3 to 0.22.4 [#3218](https://github.com//kumahq/kuma/pull/3218)
* chore: bump helm chart version to 0.8 [#3202](https://github.com//kumahq/kuma/pull/3202)

## [1.4.0]
> Released on 2021/11/22

Changes:
* chore(*) scripts for build, publish and fetch Envoy binaries [#3110](https://github.com//kumahq/kuma/pull/3110) [#3182](https://github.com//kumahq/kuma/pull/3182)
* chore(kuma-cp) upgrade gui to new version [#3178](https://github.com//kumahq/kuma/pull/3178) [#3179](https://github.com//kumahq/kuma/pull/3179)
* chore(kuma-cp) Use go structs instead of gotemplate for bootstrap [#3156](https://github.com//kumahq/kuma/pull/3156) [#3173](https://github.com//kumahq/kuma/pull/3173)
* chore(deps): bump github.com/slok/go-http-metrics from 0.9.0 to 0.10.0 [#3170](https://github.com//kumahq/kuma/pull/3170)
* Disable reporting by default [#3070](https://github.com//kumahq/kuma/pull/3070) [#3159](https://github.com//kumahq/kuma/pull/3159)
* chore(kumactl) remove install CRDs filter function [#3139](https://github.com//kumahq/kuma/pull/3139)
* feat(kuma-dp) Add conf to disable service vip [#3143](https://github.com//kumahq/kuma/pull/3143)
* chore(kuma-cp) update some TODO comments [#3141](https://github.com//kumahq/kuma/pull/3141)
* feat(kuma-cp) Add kuma.io/ignore annotation [#3142](https://github.com//kumahq/kuma/pull/3142)
* fix(kuma-dp) match gateway cluster names in the hijacker [#3106](https://github.com//kumahq/kuma/pull/3106)
* feat: add ECDSA certificate generator support [#3093](https://github.com//kumahq/kuma/pull/3093)
* feat: add more global resources to GlobalInsights [#3094](https://github.com//kumahq/kuma/pull/3094)
* feat: allow creating secrets for the not yet existing mesh [#3076](https://github.com//kumahq/kuma/pull/3076)
  👍contributed by cloudwiz
* feat: don't add v6 in DNS when v6 is disabled [#3089](https://github.com//kumahq/kuma/pull/3089)
* fix: explicitly disable dns in env when disabled in injector [#3077](https://github.com//kumahq/kuma/pull/3077)
* feat: added support for https tracing endpoint [#3057](https://github.com//kumahq/kuma/pull/3057)
  👍contributed by sudeeptoroy
* fix: normalize generating TLS certificates [#3027](https://github.com//kumahq/kuma/pull/3027)
* fix: zero downtime when enabling permissive mTLS [#3019](https://github.com//kumahq/kuma/pull/3019)
* feat: add deprecation notice for kuma-prometheus-sd [#2994](https://github.com//kumahq/kuma/pull/2994)
* feat: add GlobalInsights api endpoint [#3018](https://github.com//kumahq/kuma/pull/3018)
* fix: duplicate TLS certificate usage [#3008](https://github.com//kumahq/kuma/pull/3008)
* chore: add command argument count parameters [#3010](https://github.com//kumahq/kuma/pull/3010)
* feat: aggregate dp stats by type in MeshInsight [#2999](https://github.com//kumahq/kuma/pull/2999)
* chore: delete CLI flag '--bootstrap-version' [#2965](https://github.com//kumahq/kuma/pull/2965)
* feat: show the effective Dataplane address [#2977](https://github.com//kumahq/kuma/pull/2977)
* feat: aggregate services in MeshInsight [#2974](https://github.com//kumahq/kuma/pull/2974)
* fix: allow only one healthcheck [#2972](https://github.com//kumahq/kuma/pull/2972)
* feat: give CA managers all backends at once [#2956](https://github.com//kumahq/kuma/pull/2956)
* chore: normalize timeout configurer API [#2934](https://github.com//kumahq/kuma/pull/2934)
* fix: locality-aware lb for external-services [#2903](https://github.com//kumahq/kuma/pull/2903)
* feat: add install control-plane --version flag for all components [#2904](https://github.com//kumahq/kuma/pull/2904)
* feat: add zone selector to Kuma Mesh dashboard [#2860](https://github.com//kumahq/kuma/pull/2860)
* fix: possible to delete resources on Zone CP [#2665](https://github.com//kumahq/kuma/pull/2665)
* fix: make cluster names contextually unique [#3098](https://github.com//kumahq/kuma/pull/3098)
* feat: automatically enable gzip content on gateways [#3104](https://github.com//kumahq/kuma/pull/3104)
* feat: add Gateway TLS termination support [#3044](https://github.com//kumahq/kuma/pull/3044)
* feat: add gateway support for external services [#2990](https://github.com//kumahq/kuma/pull/2990)
* fix: enable secrets support for Gateway resources [#2953](https://github.com//kumahq/kuma/pull/2953)
* feat: initial connection policy support for Gateway [#2933](https://github.com//kumahq/kuma/pull/2933)
* feat: add access to generate zone ingress token [#3075](https://github.com//kumahq/kuma/pull/3075)
* feat: user token with RSA256 [#2992](https://github.com//kumahq/kuma/pull/2992)
* feat: prefix system users and groups with mesh-system [#3013](https://github.com//kumahq/kuma/pull/3013)
* feat: localhost is not an admin on kubernetes [#3003](https://github.com//kumahq/kuma/pull/3003)
* feat: user token enabled by default [#2941](https://github.com//kumahq/kuma/pull/2941)
* feat: Admin User Token bootstrap [#2923](https://github.com//kumahq/kuma/pull/2923)
* chore: refactor access control for individual access [#2983](https://github.com//kumahq/kuma/pull/2983)
* feat: support plugin based authentication including user tokens [#2895](https://github.com//kumahq/kuma/pull/2895)
* feat: User Token for API Server authentication [#2892](https://github.com//kumahq/kuma/pull/2892)
* chore: refactor authz and authn to plugins [#2837](https://github.com//kumahq/kuma/pull/2837)
* chore(kuma-cp) upgrade gui to new version [#3148](https://github.com//kumahq/kuma/pull/3148)
* chore(*) upgrade to Go 1.17.3 [#3147](https://github.com//kumahq/kuma/pull/3147)
* chore(deps): bump github.com/operator-framework/operator-lib [#3158](https://github.com//kumahq/kuma/pull/3158)
* chore(deps): bump github.com/gruntwork-io/terratest [#3130](https://github.com//kumahq/kuma/pull/3130)
* chore: update helm and controller-runtime [#2764](https://github.com//kumahq/kuma/pull/2764)
* chore: bump github.com/lib/pq from 1.10.3 to 1.10.4 [#3131](https://github.com//kumahq/kuma/pull/3131)
* chore: bump google.golang.org/grpc from 1.41.0 to 1.42.0 [#3101](https://github.com//kumahq/kuma/pull/3101)
* chore: bump github.com/prometheus/common from 0.31.1 to 0.32.1 [#3006](https://github.com//kumahq/kuma/pull/3006)
* chore: bump github.com/envoyproxy/protoc-gen-validate [#3007](https://github.com//kumahq/kuma/pull/3007)
* chore: bump github.com/google/uuid from 1.2.0 to 1.3.0 [#2839](https://github.com//kumahq/kuma/pull/2839)
* chore: bump sigs.k8s.io/controller-runtime from 0.10.2 to 0.10.3 [#3132](https://github.com//kumahq/kuma/pull/3132)
* chore: bump k8s.io/client-go from 0.22.2 to 0.22.3 [#3061](https://github.com//kumahq/kuma/pull/3061)
* chore: bump k8s.io/apiextensions-apiserver from 0.22.2 to 0.22.3 [#3059](https://github.com//kumahq/kuma/pull/3059)
* chore: bump k8s.io/api from 0.22.2 to 0.22.3 [#3058](https://github.com//kumahq/kuma/pull/3058)
* chore: bump github.com/golang-migrate/migrate/v4 [#2970](https://github.com//kumahq/kuma/pull/2970)
* chore: bump helm.sh/helm/v3 from 3.6.1 to 3.7.1 [#2968](https://github.com//kumahq/kuma/pull/2968)
* chore: bump github.com/miekg/dns from 1.0.14 to 1.1.43 in /pkg/transparentproxy/istio [#2752](https://github.com//kumahq/kuma/pull/2752)

## [1.3.1]
> Released on 2021/10/06

Changes:

* fix: disable zone [#2884](https://github.com//kumahq/kuma/pull/2884)
* fix: limit number of postgres connection by default [#2866](https://github.com//kumahq/kuma/pull/2866)
* feat: add zone selector to Kuma Service to Service dashboard [#2876](https://github.com//kumahq/kuma/pull/2876)
* feat: add zone selector to Kuma Service dashboard [#2865](https://github.com//kumahq/kuma/pull/2865)
* feat: add zone selector to Kuma Dataplane dashboard [#2864](https://github.com//kumahq/kuma/pull/2864)
* fix: fix duplicates in dataplane list in Kuma Services dashboard [#2845](https://github.com//kumahq/kuma/pull/2845)
* chore: migrate install resources from rbac API v1beta1 to v1 [#2875](https://github.com//kumahq/kuma/pull/2875)
* fix: fault injection matching [#2757](https://github.com//kumahq/kuma/pull/2757)
* fix: delete kuma.io/region and kuma.io/sub-zone [#2824](https://github.com//kumahq/kuma/pull/2824)
* feat: print control plane version with version cmd [#2834](https://github.com//kumahq/kuma/pull/2834)
* fix: Only warn about version compatibility where it makes sense [#2828](https://github.com//kumahq/kuma/pull/2828)
* perf: remove insight update rate limit burst [#2825](https://github.com//kumahq/kuma/pull/2825)
* perf: apply ratelimit to service insights [#2815](https://github.com//kumahq/kuma/pull/2815)
* feat: adds support for specifying specific IP for cloud provider load balancers for ingress service [#2779](https://github.com//kumahq/kuma/pull/2779)
  👍contributed by @jamesdbloom
* fix: send tool output to stdout [#2787](https://github.com//kumahq/kuma/pull/2787)
* fix: switch to a Kuma fork of go-control-plane [#2771](https://github.com//kumahq/kuma/pull/2771)
* chore: parametrize label on the deployment [#2765](https://github.com//kumahq/kuma/pull/2765)
* perf: set Node only on first DiscoveryRequest [#2741](https://github.com//kumahq/kuma/pull/2741)
* feat: verify ServiceAccountToken bound to a Pod [#2745](https://github.com//kumahq/kuma/pull/2745)
* feat: internal dns should resolve AAAA records [#2760](https://github.com//kumahq/kuma/pull/2760)
* fix: Add FORMERR and NOTIMP in alternate default coredns conf [#2756](https://github.com//kumahq/kuma/pull/2756)
* fix: virtual probes with query [#2706](https://github.com//kumahq/kuma/pull/2706)
* fix: Avoid calling `Send()` from different goroutines [#2573](https://github.com//kumahq/kuma/pull/2573)
* feat: automatically set proxy concurrency [#2691](https://github.com//kumahq/kuma/pull/2691)
* feat: Improve builtin grafana setup to have traces and logs linked [#2716](https://github.com//kumahq/kuma/pull/2716)
* fix: Show gateway services in service-insights [#2711](https://github.com//kumahq/kuma/pull/2711)
* fix: Correct bad merging of duration [#2700](https://github.com//kumahq/kuma/pull/2700)
* fix: Ensure outbounds are set when migrating from old to new [#2698](https://github.com//kumahq/kuma/pull/2698)
* fix: get rid of regex for parsing IPs [#2681](https://github.com//kumahq/kuma/pull/2681)
* feat: add CP config to ZoneInsights [#2661](https://github.com//kumahq/kuma/pull/2661)
* feat: generate GatewayRoute clusters [#2819](https://github.com//kumahq/kuma/pull/2819)
* feat: add GatewayRoute route generation [#2782](https://github.com//kumahq/kuma/pull/2782)
* feat: match gateway routes [#2758](https://github.com//kumahq/kuma/pull/2758)
* feat: initial gateway TrafficRoute support [#2547](https://github.com//kumahq/kuma/pull/2547)
* feat: add a GatewayRoute resource [#2591](https://github.com//kumahq/kuma/pull/2591)
* chore: update base image for kuma-dp [#2881](https://github.com//kumahq/kuma/pull/2881)
* chore: change Go JWT version to fix  security vunerability [#2844](https://github.com//kumahq/kuma/pull/2844)
* chore: bump go.uber.org/zap from 1.17.0 to 1.19.1 [#2768](https://github.com//kumahq/kuma/pull/2768)
* chore: bump google.golang.org/grpc from 1.38.0 to 1.40.0 [#2737](https://github.com//kumahq/kuma/pull/2737)
* chore: bump github.com/miekg/dns from 1.1.42 to 1.1.43 [#2769](https://github.com//kumahq/kuma/pull/2769)
* chore: upgrade github.com/spf13/cobra [#2732](https://github.com//kumahq/kuma/pull/2732)
* chore: bump alpine in /tools/releases/dockerfiles [#2705](https://github.com//kumahq/kuma/pull/2705)
* chore: bump github.com/onsi/gomega from 1.13.0 to 1.16.0 [#2657](https://github.com//kumahq/kuma/pull/2657)
* chore: update envoy to 1.18.4 [#2667](https://github.com//kumahq/kuma/pull/2667)

## [1.3.0]
> Released on 2021/08/25

Changes:
* feat: remove provided ca cert validation [#2663](https://github.com/kumahq/kuma/pull/2663)
 👍contributed by Nikita Pande (@nikita15p)
* feat: Use kuma-sd in kumactl install metrics [#2654](https://github.com/kumahq/kuma/pull/2654)
* feat: Add new datasource to kumactl install metrics [#2640](https://github.com/kumahq/kuma/pull/2640)
* fix: remove extra endline in traffic log default template [#2514](https://github.com//kumahq/kuma/pull/2514)
* fix: TLSInspector is causing tcp healthcheck failures [#2639](https://github.com//kumahq/kuma/pull/2639)
* feat: Add rate-limit to outbound interfaces [#2435](https://github.com//kumahq/kuma/pull/2435)
* fix: print a newline with transparent proxy setup message [#2634](https://github.com//kumahq/kuma/pull/2634)
* chore: bump alpine in /tools/releases/dockerfiles [#2531](https://github.com//kumahq/kuma/pull/2531)
* chore: annotate required fields in proto files [#2556](https://github.com//kumahq/kuma/pull/2556)
* chore: remove MADS v1alpha1 [#2632](https://github.com//kumahq/kuma/pull/2632)
* chore: parametrize kuma tracing in ZipkinCollectorURL [#2635](https://github.com//kumahq/kuma/pull/2635)
* chore: Add the number of services to usage stats [#2628](https://github.com//kumahq/kuma/pull/2628)
* feat: Add the permissive mTLS mode [#2579](https://github.com//kumahq/kuma/pull/2579)
* chore: open CAProvider and MeshValidator for extensions [#2618](https://github.com//kumahq/kuma/pull/2618)
* feat: Add entity for virtual-outbound [#2576](https://github.com//kumahq/kuma/pull/2576)
* fix: Don't set zap.Development() in debug log [#2608](https://github.com//kumahq/kuma/pull/2608)
* chore(kuma-cp) upgrade gui to new version [#2611](https://github.com//kumahq/kuma/pull/2611), [#2452](https://github.com//kumahq/kuma/pull/2452), [#2554](https://github.com//kumahq/kuma/pull/2554), [#2528](https://github.com//kumahq/kuma/pull/2528), [#2497](https://github.com//kumahq/kuma/pull/2497), [#2490](https://github.com//kumahq/kuma/pull/2490), [#2481](https://github.com//kumahq/kuma/pull/2481)
* feat: Build kuma on Windows [#2597](https://github.com//kumahq/kuma/pull/2597), [#2606](https://github.com//kumahq/kuma/pull/2606), [#2559](https://github.com//kumahq/kuma/pull/2559)
* feat: Add CA backend stats in Dataplane and Mesh Insights [#2562](https://github.com//kumahq/kuma/pull/2562)
* fix: missing key for kv in reports logging [#2598](https://github.com//kumahq/kuma/pull/2598)
* chore: split listener configurers across source files [#2592](https://github.com//kumahq/kuma/pull/2592)
* feat: add simple HTTP connection configurers [#2593](https://github.com//kumahq/kuma/pull/2593)
* feat: add virtual host domain name configurer [#2590](https://github.com//kumahq/kuma/pull/2590)
* feat: return instance and cluster IDs in kuma-cp API statuses [#2589](https://github.com//kumahq/kuma/pull/2589)
* tests: allow kuma-specific const to be overridden [#2582](https://github.com//kumahq/kuma/pull/2582)
* feat: Intermediate CA support [#2575](https://github.com//kumahq/kuma/pull/2575)
* fix: Avoid nil dereferencing in dp validator [#2578](https://github.com//kumahq/kuma/pull/2578)
* chore: consistently use utils package for protobuf wrappers [#2570](https://github.com//kumahq/kuma/pull/2570)
* fix: subscription finalizer, rev 2 [#2526](https://github.com//kumahq/kuma/pull/2526)
* tests: fix flaky test for locality aware loadbalancing [#2564](https://github.com//kumahq/kuma/pull/2564)
* fix: DP tracking lock consistency fix [#2567](https://github.com//kumahq/kuma/pull/2567)
* chore: Certificates over ADS [#2558](https://github.com//kumahq/kuma/pull/2558)
* chore: migrate DiscoveryRequest/Response in KDS to V3 [#2541](https://github.com//kumahq/kuma/pull/2541)
* feat: Rewrite dns persistence to allow virtual-outbound to be added [#2484](https://github.com//kumahq/kuma/pull/2484)
* fix: deleted default policy is created on Kuma CP restart [#2507](https://github.com//kumahq/kuma/pull/2507)
* chore: Move kumactl logging arguments to where they can be parameterized [#2544](https://github.com//kumahq/kuma/pull/2544)
* chore: add route and virtual host configuration helpers [#2517](https://github.com//kumahq/kuma/pull/2517)
* chore: fix kumactl generate dataplane proxy-type flag deprecation message [#2522](https://github.com//kumahq/kuma/pull/2522)
 👍contributed by Tharun Rajendran
* chore: Simplify resource-gen.go by generating `ResourceDescriptor` [#2511](https://github.com//kumahq/kuma/pull/2511)
* chore: Replace netcat with test server [#2510](https://github.com//kumahq/kuma/pull/2510)
* feat: configure SNI on ExternalService [#2467](https://github.com//kumahq/kuma/pull/2467)
* chore: add importas to golangci-lint [#2516](https://github.com//kumahq/kuma/pull/2516)
 👍contributed by Tharun Rajendran
* chore: add to resource-gen.go generation of kds options [#2487](https://github.com//kumahq/kuma/pull/2487)
* chore: add to resource-gen.go generation of kumactl options  [#2469](https://github.com//kumahq/kuma/pull/2469)
* fix: add owner when create ZoneIngressInsight [#2456](https://github.com//kumahq/kuma/pull/2456)
* fix: hijacker merge labels [#2476](https://github.com//kumahq/kuma/pull/2476)
* chore: improve resource-gen by auto generating ws code [#2466](https://github.com//kumahq/kuma/pull/2466)
* fix: clarify invalid resource type message [#2473](https://github.com//kumahq/kuma/pull/2473)
* fix: implement TextMarshaler for JSON keys [#2475](https://github.com//kumahq/kuma/pull/2475)
* chore: simplify resourceWsDefinition and server init [#2477](https://github.com//kumahq/kuma/pull/2477)
* fix: Stop adding outbounds to dp for vips [#2421](https://github.com//kumahq/kuma/pull/2421)
* chore(*) make port validation consistent [#2448](https://github.com//kumahq/kuma/pull/2448)

## [1.2.3]
> Released on 2021/07/29
 
Changes:
* fix(kumactl) warn about fail to check the CP version [#2438](https://github.com//kumahq/kuma/pull/2438)
* fix(kuma-cp) handle missing connection info [#2439](https://github.com//kumahq/kuma/pull/2439)
* chore(xds) rename logger to have consistent naming style [#2375](https://github.com//kumahq/kuma/pull/2375)
  👍contributed by burntcarrot
* fix(kuma-cp) set better keep-alive for bootstrap [#2432](https://github.com//kumahq/kuma/pull/2432)
* fix(kuma-dp) validate the DP proxy type [#2186](https://github.com//kumahq/kuma/pull/2186)
* fix(kuma-cp) use the typed config for TLS Inspector [#2373](https://github.com//kumahq/kuma/pull/2373)

## [1.2.2]
> Released on 2021/07/16

Changes:
* feat: add datadog traffic tracing [#2269](https://github.com//kumahq/kuma/pull/2247)
* refactor: add kumactl install tracing context [#2343](https://github.com//kumahq/kuma/pull/2343)
* chore: improve kumactl install transparent-proxy flags description, add extra validation [#2352](https://github.com//kumahq/kuma/pull/2352)
* fix: broken SDS auth and XDS generation on rapid DP restarts [#2342](https://github.com//kumahq/kuma/pull/2342)
* fix: allow verbose log levels [#2351](https://github.com//kumahq/kuma/pull/2351)
* chore: use resource types for DataplaneInsight tracking [#2324](https://github.com//kumahq/kuma/pull/2324)
* chore: improve resource manager initialization readability [#2316](https://github.com//kumahq/kuma/pull/2316)
* chore: upgrade gui to new version [#2340](https://github.com//kumahq/kuma/pull/2340), [#2325](https://github.com//kumahq/kuma/pull/2325), [#2315](https://github.com//kumahq/kuma/pull/2315)
* fix: allocate a new VIP for ExternalService host [#2302](https://github.com//kumahq/kuma/pull/2302)
* fix: stop components on leader election lost [#2318](https://github.com//kumahq/kuma/pull/2318)
* chore: generate system resource wrappers [#2282](https://github.com//kumahq/kuma/pull/2282), [#2311](https://github.com//kumahq/kuma/pull/2311)
* chore: remove access log V2 [#2301](https://github.com//kumahq/kuma/pull/2301)
* chore: generate DeepCopy interfaces [#2222](https://github.com//kumahq/kuma/pull/2222)
* chore: disable log sampling [#2273](https://github.com//kumahq/kuma/pull/2273)
* chore: upgrade Protocol Buffers [#2244](https://github.com//kumahq/kuma/pull/2244)
* chore: change default number of insights subscriptions [#2266](https://github.com//kumahq/kuma/pull/2266)
* chore: make the authentication interface type oblivious [#2271](https://github.com//kumahq/kuma/pull/2271)
* fix: fix hds disabled on dpserver [#2268](https://github.com//kumahq/kuma/pull/2268)
 👍contributed by Bastien Chatelard
* chore: refactor xDS metadata to store a generic resource [#2264](https://github.com//kumahq/kuma/pull/2264)
* feat: change KDS max message limit [#2265](https://github.com//kumahq/kuma/pull/2265)

## [1.2.1]
> Released on  2021/06/30

Changes:
* fix: Dataplane/ZoneIngress/Zone status problem when control plane forcefully exits [#2246](https://github.com//kumahq/kuma/pull/2246)
* chore: reduce memory usage by reducing cache key size [#2214](https://github.com//kumahq/kuma/pull/2214) [#2230](https://github.com//kumahq/kuma/pull/2230) 
  👍contributed by nhamlh
* fix: ZoneIngress always shows up as 'offline' [#2209](https://github.com//kumahq/kuma/pull/2209)
* feat: dataplane use advertise address to add a routable ip if address is not public ip [#2116](https://github.com//kumahq/kuma/pull/2116)
  👍contributed by sudeeptoroy
* fix: builtin DNS resolve alias with dots [#2208](https://github.com//kumahq/kuma/pull/2208)
* feat: add SNI to TLSed ExternalServices [#2211](https://github.com//kumahq/kuma/pull/2211)
* fix: fix race condition in cache [#2202](https://github.com//kumahq/kuma/pull/2202)
  👍contributed by nhamlh
* fix: supported versions of Kuma DP in the GUI [#2193](https://github.com//kumahq/kuma/pull/2193)

## [1.2.0]
> Released on  2021/06/17

Changes:

* feat: Introduce ZoneIngress [#2147](https://github.com//kumahq/kuma/pull/2147) [#2169](https://github.com//kumahq/kuma/pull/2169)
* feat: enable dataplane dns by default [#2152](https://github.com//kumahq/kuma/pull/2152)
* feat: add --verbose flag to kuma-init [#2156](https://github.com//kumahq/kuma/pull/2156)
* feat: log rotation [#2100](https://github.com//kumahq/kuma/pull/2100)
  👍contributed by @nikita15p
* feat: mads, allow specifying fetch-timeout via query param [#2148](https://github.com//kumahq/kuma/pull/2148)
  👍contributed by @austince
* feat: mads, add support for HTTP long polling [#2121](https://github.com//kumahq/kuma/pull/2121)
  👍contributed by @austince
* feat(mads) implement v1 API [#1753](https://github.com//kumahq/kuma/pull/1753)
  👍contributed by @austince
* feat: add RateLimit policy [#2083](https://github.com//kumahq/kuma/pull/2083)
* feat: TrafficRoute L7  [#2013](https://github.com//kumahq/kuma/pull/2013)
  [#2042](https://github.com//kumahq/kuma/pull/2042) [#2062](https://github.com//kumahq/kuma/pull/2062)
  [#2072](https://github.com//kumahq/kuma/pull/2072) [#2168](https://github.com//kumahq/kuma/pull/2168)

* feat: allow renegotiation for TLS in ExternalServices [#2135](https://github.com//kumahq/kuma/pull/2135)
* feat: pass header when communicating with CP [#2049](https://github.com//kumahq/kuma/pull/2049)
  👍contributed by sudeeptoroy
* feat: change default traffic route policy [#2075](https://github.com//kumahq/kuma/pull/2075)
* feat: command to install kong enterprise ingress [#1999](https://github.com//kumahq/kuma/pull/1999)
* feat: add postgres max idle connections configuration [#2020](https://github.com//kumahq/kuma/pull/2020)
  👍contributed by @nikita15p
* feat: add kumactl --no-config flag [#2048](https://github.com//kumahq/kuma/pull/2048)
* feat: nodeselector across all pods with HELM [#2012](https://github.com//kumahq/kuma/pull/2012)
* feat: enable forwarding XFCC header [#1941](https://github.com//kumahq/kuma/pull/1941)
  👍contributed by @jewertow
* feat: TrafficPermission for ExternalServices [#1957](https://github.com//kumahq/kuma/pull/1957)
* feat: metrics hijacker [#1899](https://github.com//kumahq/kuma/pull/1899)
* feat: extend CircuitBreaker [#1655](https://github.com//kumahq/kuma/pull/1655)
* chore: remove API V2 [#2119](https://github.com//kumahq/kuma/pull/2119)
* chore: bump webhooks version [#2126](https://github.com//kumahq/kuma/pull/2126)
* chore: drop deprecated Envoy options [#2143](https://github.com//kumahq/kuma/pull/2143)
* chore: dockerfiles, add a user for kuma-cp [#2129](https://github.com//kumahq/kuma/pull/2129)
* chore: bump cni version to 0.0.9 [#2137](https://github.com//kumahq/kuma/pull/2137)
* chore: rename remote cp to zone cp [#2125](https://github.com//kumahq/kuma/pull/2125)
* chore: bump versions of logging, metrics, tracing [#2178](https://github.com//kumahq/kuma/pull/2178)
* chore: parametrize bitnami/kubectl [#2151](https://github.com//kumahq/kuma/pull/2151)
* chore: backwards compatible metrics [#2173](https://github.com//kumahq/kuma/pull/2173)
* chore: upgrade Envoy version to 1.18.3 [#2145](https://github.com//kumahq/kuma/pull/2145)
* chore updated go-control-plane [#2082](https://github.com//kumahq/kuma/pull/2082)
  👍contributed by @sudeeptoroy
* chore: fix misspelled words [#1984](https://github.com//kumahq/kuma/pull/1984)
  👍contributed by @tharun208
* chore: upgrade GUI [#2157](https://github.com//kumahq/kuma/pull/2157)
* chore namespace source names for v1 API [#1896](https://github.com//kumahq/kuma/pull/1896)
  👍contributed by @austince
* chore: use cmux for MADS server [#1887](https://github.com//kumahq/kuma/pull/1887)
* chore: Add internal support for outbound UDP listeners [#1618](https://github.com//kumahq/kuma/pull/1618)
  👍contributed by @lahabana
* chore: Avoid generating duplicate subsets in ingress
  👍contributed by @lahabana
* chore: upgrade to apiextensions.k8s.io/v1 [#1108](https://github.com//kumahq/kuma/pull/1108)
  👍contributed by @austince
* fix: Clear snapshots from cache on disconnect [#2172](https://github.com//kumahq/kuma/pull/2172)
  👍contributed by @lahabana
* fix: use service account name to identify sync [#2127](https://github.com//kumahq/kuma/pull/2127)
* fix: raise the regex program size limit [#2139](https://github.com//kumahq/kuma/pull/2139)
* fix: pass query parameters through the metrics hijacker [#2124](https://github.com//kumahq/kuma/pull/2124)
* fix: matching endpoints by tags [#2096](https://github.com//kumahq/kuma/pull/2096)
* fix: manage and warn on control plane file limits [#2057](https://github.com//kumahq/kuma/pull/2057) [#2106](https://github.com//kumahq/kuma/pull/2106)
* fix: fix transparent-proxy for GCP/GKE [#2051](https://github.com//kumahq/kuma/pull/2051)
* fix: set death signal on child processes [#2045](https://github.com//kumahq/kuma/pull/2045)
* fix: TrafficRoute in multizone issue [#1979](https://github.com//kumahq/kuma/pull/1979)

## [1.1.6]
> Released on  2021/05/13

Changes:
* feat: expose reuse_connection in healthchecks [#1952](https://github.com//kumahq/kuma/pull/1952)
* feat: allow tcp/http healthchecks together [#1951](https://github.com//kumahq/kuma/pull/1951)
* feat: kumactl option to install gateway types [#1950](https://github.com//kumahq/kuma/pull/1950)
* feat: kumactl option to install kuma demo app [#1932](https://github.com//kumahq/kuma/pull/1932)
* feat: kumactl option to install Kong ingress [#1929](https://github.com//kumahq/kuma/pull/1929)
* feat: support all tags in traffic permission [#1902](https://github.com//kumahq/kuma/pull/1902)
* fix: gateway status was always reporting offline [#1946](https://github.com//kumahq/kuma/pull/1946)
* fix: don't cache failed calls [#1894](https://github.com//kumahq/kuma/pull/1894)
  👍contributed by @lahabana
* chore: add hostname when sending traces to the collector [#1962](https://github.com//kumahq/kuma/pull/1962)
* docs: prepare api docs generation [#1741](https://github.com//kumahq/kuma/pull/1741)
* test: azure aks and e2e improvements for the CI [#1880](https://github.com//kumahq/kuma/pull/1880)
  [#1871](https://github.com//kumahq/kuma/pull/1871)
  [#1933](https://github.com//kumahq/kuma/pull/1933)
  [#1953](https://github.com//kumahq/kuma/pull/1953)
  [#1972](https://github.com//kumahq/kuma/pull/1972)

## [1.1.5]
> Released on  2021/04/28

Changes:
* feat: generate outbounds for itself [#1900](https://github.com//kumahq/kuma/pull/1900)
* chore: migrate from bintray [#1901](https://github.com//kumahq/kuma/pull/1901)
* chore: GUI updates and fixes [#1897](https://github.com//kumahq/kuma/pull/1897)
* chore: kumactl check version after loading config [#1879](https://github.com/kumahq/kuma/pull/1879)
* chore: transparent proxy improvements [#1852](https://github.com//kumahq/kuma/pull/1852)
* chore upgrade Go to 16.3 and use go embed [#1864](https://github.com//kumahq/kuma/pull/1864) [#1865](https://github.com//kumahq/kuma/pull/1865)
* fix: always set locality in multizone [#1863](https://github.com//kumahq/kuma/pull/1863)
* fix: Envoy config is created based on old Dataplane [#1848](https://github.com//kumahq/kuma/pull/1848)

## [1.1.4]
> Released on  2021/04/19

Changes:
* chore: force all DNS traffic capture [#1842](https://github.com//kumahq/kuma/pull/1842)

## [1.1.3]
> Released on  2021/04/16

Changes:
* feat: support External Services with original hostname and port (built-in DNS) 
  [#1807](https://github.com//kumahq/kuma/pull/1807) [#1811](https://github.com//kumahq/kuma/pull/1811) [#1817](https://github.com//kumahq/kuma/pull/1817) [#1812](https://github.com//kumahq/kuma/pull/1812) [#1821](https://github.com//kumahq/kuma/pull/1821) [#1824](https://github.com//kumahq/kuma/pull/1824) [#1828](https://github.com//kumahq/kuma/pull/1828) [#1822](https://github.com//kumahq/kuma/pull/1822)
* fix: pass validation of V3 specific configs in ProxyTemplate [#1819](https://github.com//kumahq/kuma/pull/1819)
* chore: support ingress annotations (kuma.io/ingress-public-address and kuma.io/ingress-public-port) in HELM [#1796](https://github.com//kumahq/kuma/pull/1796)

## [1.1.2]
> Released on  2021/04/09

Changes:
* feat: extend CircuitBreaker policy with Thresholds [#1688](https://github.com//kumahq/kuma/pull/1688)
* feat: enable IPv6 support and tests [#1726](https://github.com//kumahq/kuma/pull/1726) [#1734](https://github.com//kumahq/kuma/pull/1734)
* feat: unuversal mode transparent-proxy firewalld support [#1702](https://github.com//kumahq/kuma/pull/1702)
* feat: new Grafana charts for golden signals and L7 metrics [#1739](https://github.com//kumahq/kuma/pull/1739) [#1786](https://github.com//kumahq/kuma/pull/1786)
* chore: verify e2e tests run in EKS [#1684](https://github.com//kumahq/kuma/pull/1684)  [#1685](https://github.com//kumahq/kuma/pull/1685) [#1744](https://github.com//kumahq/kuma/pull/1744)
* chore: upgrade CRDS to apiextensions.k8s.io/v1 [#1108](https://github.com//kumahq/kuma/pull/1108)
* fix: helm cp service annotations [#1767](https://github.com//kumahq/kuma/pull/1767)
  👍contributed by nbrink91
* fix: gui fixes [#1773](https://github.com//kumahq/kuma/pull/1773)
* fix: KDS may delete ConfigMaps on Control Plane restarts [#1769](https://github.com//kumahq/kuma/pull/1769)
* fix: Kuma CP restart may cause stale Envoy configs on Universal [#1749](https://github.com//kumahq/kuma/pull/1749)
* fix: use EnvoyGRPC to fix DNS resolving [#1740](https://github.com//kumahq/kuma/pull/1740)
* fix: fix ingress-enabled [#1725](https://github.com//kumahq/kuma/pull/1725)
* fix: pick HTTP health checker version depending on outbound's protocol [#1714](https://github.com//kumahq/kuma/pull/1714)
* fix: improve the DNS server bind message [#1701](https://github.com//kumahq/kuma/pull/1701)
* fix: validate --name and --mesh when dataplane is provided [#1771](https://github.com//kumahq/kuma/pull/1771)
* fix: better error messages when there is problem with pod dataplane convertion [#1743](https://github.com//kumahq/kuma/pull/1743)
* fix: crashes under load [#1694](https://github.com//kumahq/kuma/pull/1694) [#1695](https://github.com//kumahq/kuma/pull/1695)

## [1.1.1]
> Released on  2021/03/11

* fix: make sure we enumerate all types in kumactl [#1673](https://github.com//kumahq/kuma/pull/1673)
* fix: annnotate service with ingress that has no annotations [#1671](https://github.com//kumahq/kuma/pull/1671)
* fix: improve err message if $HOME is not defined [#1664](https://github.com//kumahq/kuma/pull/1664)
* feat: zipkin config add shared span context option [#1660](https://github.com//kumahq/kuma/pull/1660)
  👍contributed by @ericmustin
* feat: get rid of 'changed' check [#1663](https://github.com//kumahq/kuma/pull/1663)

## [1.1.0]
> Released on  2021/03/08

* feat: default to xDS v3 [#1642](https://github.com//kumahq/kuma/pull/1642)
  [#1589](https://github.com//kumahq/kuma/pull/1589)
  [#1442](https://github.com//kumahq/kuma/pull/1442)
  [#1399](https://github.com//kumahq/kuma/pull/1399)
  [#1408](https://github.com//kumahq/kuma/pull/1408)
  [#1383](https://github.com//kumahq/kuma/pull/1383)
* feat: set timeouts [#1554](https://github.com//kumahq/kuma/pull/1554)
* feat: add default retry policy [#1606](https://github.com//kumahq/kuma/pull/1606)
* feat: appProtocol for application protocol [#1413](https://github.com//kumahq/kuma/pull/1413)
  👍contributed by @tharun208
* feat: Resource counter based on Mesh insights [#1423](https://github.com//kumahq/kuma/pull/1423)
  👍contributed by @jewertow
* feat: set auto_host_rewrite to true for ExternalServices [#1635](https://github.com//kumahq/kuma/pull/1635)
* feat: health check add event log support [#1631](https://github.com//kumahq/kuma/pull/1631)
* feat: introduce 'healthy_panic_theshold' for HealthCheck policy [#1625](https://github.com//kumahq/kuma/pull/1625)
* feat: working directory for kuma-cp [#1573](https://github.com//kumahq/kuma/pull/1573)
* feat: inject ingress.kubernetes.io/service-upstream [#1608](https://github.com//kumahq/kuma/pull/1608)
* feat: global secrets [#1603](https://github.com//kumahq/kuma/pull/1603)
* feat: NACK backoff [#1591](https://github.com//kumahq/kuma/pull/1591)
* feat: degraded status in insights [#1563](https://github.com//kumahq/kuma/pull/1563)
* chore: remove deprecated options [#1652](https://github.com//kumahq/kuma/pull/1652)
* chore: inlineString in DataSource [#1514](https://github.com//kumahq/kuma/pull/1514)
* chore: use `kumactl install transparent-proxy` in `kuma-init` [#1599](https://github.com//kumahq/kuma/pull/1599)
* chore: expose versions endpoint to extension
  [#1638](https://github.com//kumahq/kuma/pull/1638)
  [#1639](https://github.com//kumahq/kuma/pull/1639)
  [#1602](https://github.com//kumahq/kuma/pull/1602)
* chore: sidecar env vars override [#1562](https://github.com//kumahq/kuma/pull/1562)
* chore: Add support for UDP Listeners [#1568](https://github.com//kumahq/kuma/pull/1568)
* chore: add msg on transparent proxy that ssh connection may drop [#1630](https://github.com//kumahq/kuma/pull/1630)
* chore: Version bumps
  Envoy 1.16.2 -> 1.17.1
  Grafana 7.1.4 -> 7.4.3
  Prometheus 2.18.2 -> 2.25.0
  Jaeger 1.18 -> 1.22  
  [#1637](https://github.com//kumahq/kuma/pull/1637) [#1619](https://github.com//kumahq/kuma/pull/1619)
* chore: improve Kuma extensibility
  [#1572](https://github.com//kumahq/kuma/pull/1572)
  [#1550](https://github.com//kumahq/kuma/pull/1550)
  [#1516](https://github.com//kumahq/kuma/pull/1516)
  [#1513](https://github.com//kumahq/kuma/pull/1513)
  [#1493](https://github.com//kumahq/kuma/pull/1493)
* fix: Postgres TLS modes in kuma-cp.defaults.yaml [#1611](https://github.com//kumahq/kuma/pull/1611)
  👍contributed by Andy Bailey
* fix: typo [#1628](https://github.com//kumahq/kuma/pull/1628)
  👍contributed by @pgold30
* fix: require -f parameter to kumactl apply [#1590](https://github.com//kumahq/kuma/pull/1590)

## [1.0.8]
> Released on  2021/02/19

Changes:
* feat: health checks: http, custom strings and jitter with tcp [#1261](https://github.com//kumahq/kuma/pull/1261) [#1570](https://github.com//kumahq/kuma/pull/1570)
* chore: set GRPC keepalives [#1580](https://github.com//kumahq/kuma/pull/1580)
* chore: gui update [#1564](https://github.com//kumahq/kuma/pull/1564)
* fix: network attachment definitions for CNI [#1569](https://github.com//kumahq/kuma/pull/1569)
* fix: remove the datplane of any completed Pod [#1576](https://github.com//kumahq/kuma/pull/1576)
* fix: check health while configure the ingress [#1508](https://github.com//kumahq/kuma/pull/1508)
* fix: close/unlink the dp access log unix socket [#1574](https://github.com//kumahq/kuma/pull/1574)

## [1.0.7]
> Released on  2021/02/10

Changes:
* feat: upgraded GUI to include charts [#1532](https://github.com//kumahq/kuma/pull/1532) [#1545](https://github.com//kumahq/kuma/pull/1545)
* feat: support Kubernetes Jobs [#1497](https://github.com//kumahq/kuma/pull/1497) [#1480](https://github.com//kumahq/kuma/pull/1480) [#1481](https://github.com//kumahq/kuma/pull/1481)
* feat: support Service-less Pods [#1460](https://github.com//kumahq/kuma/pull/1460)
* chore: resolve both DNS compliant and non-compliant names [#1485](https://github.com//kumahq/kuma/pull/1485) [#1533](https://github.com//kumahq/kuma/pull/1533)
* chore: support Helm chart upgrades through `kumactl install crds` [#1419](https://github.com//kumahq/kuma/pull/1419)
* fix: Avoid computing vips for ingress [#1490](https://github.com//kumahq/kuma/pull/1490)
  👍contributed by @lahabana
* fix: Fix mesh outbound leak in VIPs [#1489](https://github.com//kumahq/kuma/pull/1489)
  👍contributed by @lahabana

## [1.0.6]
> Released on  2021/01/22

Changes:

* feat: dataplane and service status improvements. Introduce Health Discovery Service [#1437](https://github.com//kumahq/kuma/pull/1437) [#1404](https://github.com//kumahq/kuma/pull/1404) [#1378](https://github.com//kumahq/kuma/pull/1378) [#1418](https://github.com//kumahq/kuma/pull/1418)
* feat: user specified load balancer type per route [#1402](https://github.com//kumahq/kuma/pull/1402)
* feat: remote-cp version in ZoneInsight [#1380](https://github.com//kumahq/kuma/pull/1380)
  👍contributed by @tharun208
* feat: update mesh insights with DP versions [#1372](https://github.com//kumahq/kuma/pull/1372)
  👍contributed by @jewertow
* chore: allow resolution of service names with "." [#1448](https://github.com//kumahq/kuma/pull/1448)
  👍contributed by @lahabana
* chore: Expose the possibility to add ProxyTemplates publicly [#1452](https://github.com//kumahq/kuma/pull/1452)
  👍contributed by @lahabana
* chore: Envoy XDS v2/v3 support [#1412](https://github.com//kumahq/kuma/pull/1412) [#1398](https://github.com//kumahq/kuma/pull/1398) [#1379](https://github.com//kumahq/kuma/pull/1379)
* chore: disable prometheus/grafana in install metrics [#1447](https://github.com//kumahq/kuma/pull/1447)
* chore: change exec probes to http [#1407](https://github.com//kumahq/kuma/pull/1407)
* chore: update ecs examples [#1446](https://github.com//kumahq/kuma/pull/1446)
* fix: nil check in IsIngress() to avoid panic when validating [#1424](https://github.com//kumahq/kuma/pull/1424)
  👍contributed by @nikita15p
* fix: outbound reconciler for Universal [#1422](https://github.com//kumahq/kuma/pull/1422)
* fix: get rid of sending 'spec' in event from postgres [#1406](https://github.com//kumahq/kuma/pull/1406)
* fix: de-duplicate passed env-vars [#1367](https://github.com//kumahq/kuma/pull/1367)
  👍contributed by @hvydya


## [1.0.5]
> Released on  2021/01/07

Changes:

* perf: cached client for fetching secrets on k8s [#1393](https://github.com//kumahq/kuma/pull/1393)
* feat: added envoy gRPC status command to accesslog [#1223](https://github.com//kumahq/kuma/pull/1223)
  👍contributed by @tharun208
* feat: add control plane identifier in DiscoveryResponse [#1319](https://github.com//kumahq/kuma/pull/1319)
  👍contributed by @jewertow
* fix: allow kuma-cp config to grpcs scheme [#1390](https://github.com//kumahq/kuma/pull/1390)
  👍contributed by @lahabana
* fix: traffic logging to tcp backends [#1389](https://github.com//kumahq/kuma/pull/1389) [#1394](https://github.com//kumahq/kuma/pull/1394)
* fix: validate mesh deletion on Universal [#1285](https://github.com//kumahq/kuma/pull/1285)
* fix: logging of resource conflicts [#1254](https://github.com//kumahq/kuma/pull/1254)
* fix: allow fault injection for http2 and grpc [#1350](https://github.com//kumahq/kuma/pull/1350)


## [1.0.4]
> Released on  2020/12/22

Changes:

* feat: retry policy [#1325](https://github.com//kumahq/kuma/pull/1325) [#1352](https://github.com//kumahq/kuma/pull/1352)
* feat: add `kumactl install transparent-proxy` [#1321](https://github.com//kumahq/kuma/pull/1321)
* feat: Kuma DP + Envoy version in Dataplane Insights [#1192](https://github.com//kumahq/kuma/pull/1192)
  👍contributed by @jewertow
* feat: return Kuma DP and Envoy version in the output of `inspect dataplanes` [#1298](https://github.com//kumahq/kuma/pull/1298)
  👍contributed by @jewertow
* feat: add horizontal pod autoscaler [#1271](https://github.com//kumahq/kuma/pull/1271)
  👍contributed by @austince
* fix: bug with lost update of Dataplane [#1313](https://github.com//kumahq/kuma/pull/1313)
* fix: probe path gets a / prepended if not supplied [#1326](https://github.com//kumahq/kuma/pull/1326)
  👍contributed by @lennartquerter
* fix: rename field dpVersion to kumaDp in version schema [#1287](https://github.com//kumahq/kuma/pull/1287)
  👍contributed by @jewertow
* fix: FaultInjection will not validate source protocol [#1315](https://github.com//kumahq/kuma/pull/1315)
* chore: Automatic readme generation for chart [#1209](https://github.com//kumahq/kuma/pull/1209)
  👍contributed by @tharun208
* chore: generate filterChainMatchers based on TrafficRoutes [#1294](https://github.com//kumahq/kuma/pull/1294)
* feat: handle zone deletion [#1348](https://github.com//kumahq/kuma/pull/1348)
* fix: bug with Ingress not belonging to the mesh [#1344](https://github.com//kumahq/kuma/pull/1344)
* fix: validate empty mesh on k8s [#1340](https://github.com//kumahq/kuma/pull/1340)
* fix: scaling-up issue  [#1282](https://github.com//kumahq/kuma/pull/1282)
* chore: prepare for XDS v3 migration [#1334](https://github.com//kumahq/kuma/pull/1334) [#1323](https://github.com//kumahq/kuma/pull/1323) [#1312](https://github.com//kumahq/kuma/pull/1312) [#1290](https://github.com//kumahq/kuma/pull/1290) [#1284](https://github.com//kumahq/kuma/pull/1284)


## [1.0.3]
> Released on  2020/12/03

Changes:
* chore: disable timeout on route level [#1275](https://github.com//kumahq/kuma/pull/1275)

## [1.0.2]
> Released on  2020/12/02

Changes:

* feat: expand service insight endpoints [#1259](https://github.com//kumahq/kuma/pull/1259)
* feat: GUI improvements [#1267](https://github.com//kumahq/kuma/pull/1267)
* feat: service insights [#1163](https://github.com//kumahq/kuma/pull/1163) 
* chore: better handling of SAN mismatch when DP connects to the CP [#1205](https://github.com//kumahq/kuma/pull/1205)
* chore: increase default mesh retry timeout [#1269](https://github.com//kumahq/kuma/pull/1269)
* chore: change default flush intervals [#1237](https://github.com//kumahq/kuma/pull/1237)
* chore: deploy many instances using HELM [#1208](https://github.com//kumahq/kuma/pull/1208)
* chore: rename 'kumactl generate dataplane-token' CLI arg [#1206](https://github.com//kumahq/kuma/pull/1206) 
* chore: update Envoy to 1.16.1 [#1214](https://github.com//kumahq/kuma/pull/1214)
* fix: handle resource conflicts more gracefully [#1236](https://github.com//kumahq/kuma/pull/1236)
* fix: insight resyncer fixes [#1246](https://github.com//kumahq/kuma/pull/1246)
* fix: resolver set global VIPs [#1243](https://github.com//kumahq/kuma/pull/1243)
* fix: handle concurrent ensure default resource invocations [#1248](https://github.com//kumahq/kuma/pull/1248)
* fix: resolve probe's named port before converting to virtual [#1232](https://github.com//kumahq/kuma/pull/1232)
* fix: errors on delete secrets [#1222](https://github.com//kumahq/kuma/pull/1222)
* fix: resync IPAM in case of Kuma CP restart (bp #1213) [#1227](https://github.com//kumahq/kuma/pull/1227)
* fix: do not replace autogenerated certs [#1215](https://github.com//kumahq/kuma/pull/1215)

## [1.0.1]
> Released on  2020/11/23

* chore: update GUI to the newest version [#1202](https://github.com//kumahq/kuma/pull/1202)
* fix: probes without inbound [#1199](https://github.com//kumahq/kuma/pull/1199)
* fix: create default mesh resources when default mesh is skipped [#1178](https://github.com//kumahq/kuma/pull/1178)
* chore: handle mesh delete more gracefully [#1185](https://github.com//kumahq/kuma/pull/1185)
* fix: fix virtual probes disabling and envs [#1171](https://github.com//kumahq/kuma/pull/1171)
* feat: parametrize Kuma CP config via HELM [#1175](https://github.com//kumahq/kuma/pull/1175)
* fix: handle missing TrafficRoute [#1188](https://github.com//kumahq/kuma/pull/1188)

## [1.0.0]
> Released on  2020/11/17

* feat: new multizone deployment flow [#1122](https://github.com//kumahq/kuma/pull/1122) [#1125](https://github.com//kumahq/kuma/pull/1125) [#1133](https://github.com//kumahq/kuma/pull/1133)
⚠️ warning: breaking change
* feat: performance optimisation [#1045](https://github.com//kumahq/kuma/pull/1045) [#1113](https://github.com//kumahq/kuma/pull/1113)
* feat: improved control plane communication security [#1065](https://github.com//kumahq/kuma/pull/1065) [#1069](https://github.com//kumahq/kuma/pull/1069) [#1083](https://github.com//kumahq/kuma/pull/1083) [#1084](https://github.com//kumahq/kuma/pull/1084) [#1092](https://github.com//kumahq/kuma/pull/1092) [#1115](https://github.com//kumahq/kuma/pull/1115) [#1118](https://github.com//kumahq/kuma/pull/1118)
⚠️ warning: breaking change
* feat: locality aware load balancing [#1111](https://github.com//kumahq/kuma/pull/1111) 
* feat: add ExternalService  [#1025](https://github.com//kumahq/kuma/pull/1025) [#1058](https://github.com//kumahq/kuma/pull/1058) [#1062](https://github.com//kumahq/kuma/pull/1062) [#1080](https://github.com//kumahq/kuma/pull/1080) [#1094](https://github.com//kumahq/kuma/pull/1094) 
* feat: add kafka protocol support [#1121](https://github.com//kumahq/kuma/pull/1121)
* feat: exclude injection from pods that match labels [#1072](https://github.com//kumahq/kuma/pull/1072)
* feat: create default resources for Mesh [#1141](https://github.com//kumahq/kuma/pull/1141) [#1149](https://github.com/kumahq/kuma/pull/1149) [#1154](https://github.com//kumahq/kuma/pull/1154) [#1155](https://github.com//kumahq/kuma/pull/1155) 
⚠️ warning: breaking change
* chore: GUI updates [#1061](https://github.com/kumahq/kuma/pull/1061) [#1123](https://github.com//kumahq/kuma/pull/1123) [#1156](https://github.com/kumahq/kuma/pull/1156)
* feat: auth on XDS [#1040](https://github.com//kumahq/kuma/pull/1040)
* feat: merge install ingress into install control-plane [#1038](https://github.com//kumahq/kuma/pull/1038) 
 👍contributed by @austince
⚠️ warning: breaking change
* feat: support HTTP probes with mTLS enabled [#1036](https://github.com//kumahq/kuma/pull/1036)
* feat: mesh insights [#1143](https://github.com/kumahq/kuma/pull/1143)
* feat: autoconfigure single cert for all services [#1032](https://github.com//kumahq/kuma/pull/1032)
⚠️ warning: breaking change
* feat: cache with better performance and debug endpoints [#1018](https://github.com//kumahq/kuma/pull/1018) 
* feat: Kuma CP metrics [#993](https://github.com//kumahq/kuma/pull/993) [#1014](https://github.com//kumahq/kuma/pull/1014)
* fix: signing token in multizone [#1007](https://github.com//kumahq/kuma/pull/1007) 
* feat: dataplane token bound to a service [#1004](https://github.com//kumahq/kuma/pull/1004) [#1136](https://github.com//kumahq/kuma/pull/1136)
* feat: new dataplane lifecycle [#999](https://github.com//kumahq/kuma/pull/999) 
* feat: apply multiple resources with kumactl [#1057](https://github.com//kumahq/kuma/pull/1057) 
 👍contributed by @tharun208
 * chore: Helm improvements [#990](https://github.com//kumahq/kuma/pull/990) [#1053](https://github.com//kumahq/kuma/pull/1053) [#1066](https://github.com//kumahq/kuma/pull/1066)  [#1120](https://github.com//kumahq/kuma/pull/1120) [#1147](https://github.com//kumahq/kuma/pull/1147)
  👍contributed by @austince
* chore: change policies on K8S to scope global [#1148](https://github.com//kumahq/kuma/pull/1148) [#1127](https://github.com//kumahq/kuma/pull/1127)
⚠️ warning: breaking change
* feat: protocol tag for gateway & ingress [#984](https://github.com//kumahq/kuma/pull/984) 
* feat: domain name support in dataplane.networking.address [#965](https://github.com//kumahq/kuma/pull/965) 
* feat: examples for ECS Universal deployments [#1003](https://github.com//kumahq/kuma/pull/1003) 
* chore: get rid of advertised hostname [#1159](https://github.com//kumahq/kuma/pull/1159)
⚠️ warning: breaking change 
* chore: improve DP insights API filtering [#1104](https://github.com//kumahq/kuma/pull/1104)
* chore: Ingress Dataplane on K8S can only be deployed in system namespace [#1070](https://github.com//kumahq/kuma/pull/1070)
* chore: use /ready endpoint for sidecar health-check [#1055](https://github.com//kumahq/kuma/pull/1055) 
 👍contributed by @tharun208
* fix: missing kuma.io prefix in example dataplane [#1054](https://github.com//kumahq/kuma/pull/1054) 
 👍contributed by @nikita15p
* fix: CNI relies on annotations [#1043](https://github.com//kumahq/kuma/pull/1043) 
* Fixed Developer.md for make build/kumactl [#1027](https://github.com//kumahq/kuma/pull/1027) 
 👍contributed by @nikita15p
* chore: added missing variables to default config file [#1073](https://github.com//kumahq/kuma/pull/1073) 
 👍contributed by @sudeeptoroy
* chore: add dependabot config [#1067](https://github.com//kumahq/kuma/pull/1067) 
 👍contributed by @austince
* chore: added update homebrew formula github workflow [#1150](https://github.com//kumahq/kuma/pull/1150) 
 👍contributed by @tharun208
* fix: disable virtual probes for pods with gateway annotation [#1157](https://github.com//kumahq/kuma/pull/1157)  
* fix apply command to throw error when no resources are passed [#1103](https://github.com//kumahq/kuma/pull/1103) 
* fix: move diagnostics port configuration [#1140](https://github.com//kumahq/kuma/pull/1140) 
* chore: drop k8s 1.13 support [#1026](https://github.com//kumahq/kuma/pull/1026)
⚠️ warning: breaking change
* chore: migrate to golang 1.15.5 [#981](https://github.com//kumahq/kuma/pull/981) [#1153](https://github.com//kumahq/kuma/pull/1153)
* chore: envoy 1.16.0 [#1130](https://github.com//kumahq/kuma/pull/1130) [#1139](https://github.com//kumahq/kuma/pull/1139) 
* chore: choose the namespace when install dns [#1128](https://github.com//kumahq/kuma/pull/1128) 
* fix: make install metrics use --mesh flag [#1129](https://github.com//kumahq/kuma/pull/1129) 

## [0.7.3]
> Released on  2020/10/22
* chore: generate static outbound routes [#1098](https://github.com/kumahq/kuma/pull/1098/)
* feat: apply multiple resources [#1057](https://github.com/kumahq/kuma/pull/1057/)
 👍contributed by @tharun208
* chore: generate cert with SAN for the newest K8S [#1078](https://github.com/kumahq/kuma/pull/1078)
* feat: specify nodeSelectors for CP and CNI pods [#990](https://github.com/kumahq/kuma/pull/990/)
 👍contributed by @austince
* feat: exclude injection from pods that match labels [#1072](https://github.com/kumahq/kuma/pull/1072/)
* chore: use /ready endpoint for sidecar health-check [#1055](https://github.com/kumahq/kuma/pull/1055/)
 👍contributed by @tharun208

## [0.7.2]

* feat: fix CNI with the latest changes and bump the CNI image to 0.0.2 [#1049](https://github.com//kumahq/kuma/pull/1049) [#1043](https://github.com//kumahq/kuma/pull/1043) 
* feat: exclude traffic interceptions on port using annotations [#1046](https://github.com//kumahq/kuma/pull/1046) 
* feat: central place for creating defaults [#1017](https://github.com//kumahq/kuma/pull/1017) 
* fix: metric to DP-CP connection should rely on control_plane.connected_state [#1009](https://github.com//kumahq/kuma/pull/1009) 
* fix: use not deprecated value to disable auth on universal [#1008](https://github.com//kumahq/kuma/pull/1008) 
* fix: signing token in multizone [#1007](https://github.com//kumahq/kuma/pull/1007) 
* Generate inbound/outbound for HTTP/2 [#998](https://github.com//kumahq/kuma/pull/998) 
* feat: cleanup dataplanes after 3d of the offline state [#987](https://github.com//kumahq/kuma/pull/987) 
* feat: validate zone location apply [#986](https://github.com//kumahq/kuma/pull/986) 
* feat: change failpolicy of service hook to ignore [#983](https://github.com//kumahq/kuma/pull/983) 
* fix: direct access for ingress [#985](https://github.com//kumahq/kuma/pull/985) 
* feat: retry connection to the CP and for fetching bootstrap [#982](https://github.com//kumahq/kuma/pull/982) 
* fix: ignore services without selectors [#978](https://github.com//kumahq/kuma/pull/978) 
* feat: parametrize kuma deploy [#973](https://github.com//kumahq/kuma/pull/973) 
* fix: zone insights manager and limits [#976](https://github.com//kumahq/kuma/pull/976) 
* feat: validate zone and global addresses [#967](https://github.com//kumahq/kuma/pull/967) 

## [0.7.1]
> Released on  2020/08/12

Changes:
* feat: add Helm chart for kuma [#916](https://github.com//kumahq/kuma/pull/916)
 👍contributed by @austince
 [#945](https://github.com//kumahq/kuma/pull/945) [#956](https://github.com//kumahq/kuma/pull/956) [#957](https://github.com//kumahq/kuma/pull/957) [#962](https://github.com//kumahq/kuma/pull/962) [#966](https://github.com//kumahq/kuma/pull/966)

* feat: gRPC support [#924](https://github.com//kumahq/kuma/pull/924)
 👍contributed by @tharun208

* fix: support http2 and grpc on outbound [#958](https://github.com//kumahq/kuma/pull/958)

* feat: compile Kuma with custom Runtime and Bootstrap plugins [#947](https://github.com//kumahq/kuma/pull/947)

* fix: GUI access from remote hosts [#963](https://github.com//kumahq/kuma/pull/963)

* fix: dry-run after Kuma installed [#944](https://github.com//kumahq/kuma/pull/944)


## [0.7.0]
> Released on  2020/07/29

Changes:
 
* feat: Updated Proxy Template [#883](https://github.com//kumahq/kuma/pull/883)
[#877](https://github.com//kumahq/kuma/pull/877)
[#909](https://github.com//kumahq/kuma/pull/909) 

* chore: CNCF donation [#896](https://github.com//kumahq/kuma/pull/896)
[#897](https://github.com//kumahq/kuma/pull/897)
[#899](https://github.com//kumahq/kuma/pull/899)
[#931](https://github.com//kumahq/kuma/pull/931)
 
* docs: update contributing readme [#918](https://github.com//kumahq/kuma/pull/918) 
 👍contributed by @tharun208
 
* feat: add Zone resource to register Remotes to Global [#895](https://github.com//kumahq/kuma/pull/895) 
[#917](https://github.com//kumahq/kuma/pull/917)
[#919](https://github.com//kumahq/kuma/pull/919)
[#921](https://github.com//kumahq/kuma/pull/921)
[#932](https://github.com//kumahq/kuma/pull/932)
⚠️ warning: breaking change of Distributed Kuma

*  feat: support selectively enabling Pods [#748](https://github.com//kumahq/kuma/pull/748) 
 ⚠️ warning: breaking change of K8s

*  feat: move the GUI from :5683 to :5681/gui [#915](https://github.com//kumahq/kuma/pull/915) 
 ⚠️ warning: breaking change of GUI

*  chore: prefix Kuma native tags with `kuma.io` [#910](https://github.com//kumahq/kuma/pull/910) 
 ⚠️ warning: breaking change of Dataplanes on Universal and Policies on both Kubernetes and Universal

* chore: updated versions 
[#855](https://github.com//kumahq/kuma/pull/855)
[#927](https://github.com//kumahq/kuma/pull/927)
[#933](https://github.com//kumahq/kuma/pull/933)

    - jaegertracing/all-in-one:1.17.1 -> 1.18
    - envoy 1.14.2 -> 1.15.0
    - jimmidyson/configmap-reload 0.3.0 -> 0.4.0
    - grafana/grafana 7.0.5 -> 7.1.1
    - prom/alertmanager 0.20.0 -> 0.21.0
    - quay.io/coreos/kube-state-metrics 1.9.1 -> 1.9.7
    - prom/node-exporter 0.18.1 -> 1.0.1
    - prom/pushgateway 1.0.1 -> 1.2.0
    - prom/prometheus 2.15.2 -> 2.18.2

* feat: dynamic tracing [#930](https://github.com//kumahq/kuma/pull/930) 
 
* fix: support empty labels on Pod [#922](https://github.com//kumahq/kuma/pull/922) 
 👍contributed by @tharun208
 
* feat: statefulset support [#901](https://github.com//kumahq/kuma/pull/901) 

* fix: add creation time on synced resources [#903](https://github.com//kumahq/kuma/pull/903) 

* feat: support for http2 [#911](https://github.com//kumahq/kuma/pull/911) 
 
* feat: add flag to skip default mesh creation, remove config option [#904](https://github.com//kumahq/kuma/pull/904) 
 👍contributed by @austince

* feat: add ServiceAddress to dataplane Inbound [#892](https://github.com//kumahq/kuma/pull/892) 

* fix: safely delete the kuma-system namespace [#908](https://github.com//kumahq/kuma/pull/908) 

* feat: added total weight for route configurer [#905](https://github.com//kumahq/kuma/pull/905) 
 👍contributed by @tharun208

* fix: support dry run [#906](https://github.com//kumahq/kuma/pull/906)

* fix: reduce size of access log address [#894](https://github.com//kumahq/kuma/pull/894) 
 👍contributed by @xbauquet 

* feat: check for incompatible versions on kumactl init [#736](https://github.com//kumahq/kuma/pull/736) 
 👍contributed by @tharun208

* fix: ingress per cluster (not per mesh) [#881](https://github.com//kumahq/kuma/pull/881)
 
* chore: skip Ingress endpoint if mTLS is off [#925](https://github.com//kumahq/kuma/pull/925) 

*  fix: Add the permissions to create and patch events [#884](https://github.com//kumahq/kuma/pull/884) 
 👍contributed by @andrew-teirney

*  feat: add install loki for log aggregation [#820](https://github.com//kumahq/kuma/pull/820) 
 👍contributed by @xbauquet

Breaking changes:
* ⚠️ This release changes the namespace label `kuma.io/sidecar-injection` to an annotation
* ⚠️ This release moves the GUI from a dedicated port, which defaults to `:5683` to a `/gui` path on the API server (`:5681`)
* ⚠️ This release prefixes the Kuma built-in tags with `kuma.io` as follows: `kuma.io/service`, `kuma.io/protocol`, `kuma.io/zone`
* ⚠️ This release changes the way that Distributed and Hybrid Kuma Control planes are deployed. Please refer to the [documentation](https://kuma.io/docs/0.7.0/documentation/deployments/#usage) for more details.


## [0.6.0]
> Released on  2020/06/30

Changes:
*  feat(gui) new GUI build files and binaries generated. [#873](https://github.com/kumahq/kuma/pull/873) 
*  feat: Kuma Discovery Service (KDS) [#870](https://github.com/kumahq/kuma/pull/870) [#871](https://github.com/kumahq/kuma/pull/871) [#864](https://github.com/kumahq/kuma/pull/864) [#866](https://github.com/kumahq/kuma/pull/866) [#865](https://github.com/kumahq/kuma/pull/865) [#861](https://github.com/kumahq/kuma/pull/861) [#860](https://github.com/kumahq/kuma/pull/860) [#857](https://github.com/kumahq/kuma/pull/857) [#839](https://github.com/kumahq/kuma/pull/839) [#833](https://github.com/kumahq/kuma/pull/833) [#847](https://github.com/kumahq/kuma/pull/847) [#843](https://github.com/kumahq/kuma/pull/843) [#834](https://github.com/kumahq/kuma/pull/834) [#830](https://github.com/kumahq/kuma/pull/830) 
*  feat: ingress for cross-cluster communication [#818](https://github.com/kumahq/kuma/pull/818) [#825](https://github.com/kumahq/kuma/pull/825) [#840](https://github.com/kumahq/kuma/pull/840) [#842](https://github.com/kumahq/kuma/pull/842) [#856](https://github.com/kumahq/kuma/pull/856) [#851](https://github.com/kumahq/kuma/pull/851)   
*  feat: kuma-cp DNS service [#821](https://github.com/kumahq/kuma/pull/821) [#798](https://github.com/kumahq/kuma/pull/798) [#850](https://github.com/kumahq/kuma/pull/850) [#862](https://github.com/kumahq/kuma/pull/862)
*  feat: flatten svc k8s tag [#848](https://github.com/kumahq/kuma/pull/848)
⚠️ warning: breaking change for service tag format 
*  feat: multiple outbound tags [#831](https://github.com/kumahq/kuma/pull/831)
*  chore: remove interface from dataplane model [#832](https://github.com/kumahq/kuma/pull/832)
⚠️ warning: breaking change for dataplane model
*  feat: block resources based on kuma-cp mode [#812](https://github.com/kumahq/kuma/pull/812) 
 👍contributed by @tharun208
*  feat: Multicluster config infrastructure [#788](https://github.com/kumahq/kuma/pull/788) 
 👍contributed by @tharun208
*  fix: expose Jaeger only inside of K8S cluster [#824](https://github.com/kumahq/kuma/pull/824) 
 👍contributed by @xbauquet
*  chore: update envoy 1.14.2 and alpine 3.12 [#829](https://github.com/kumahq/kuma/pull/829)
*  chore: remove passive healthchecks [#869](https://github.com/kumahq/kuma/pull/869) 
⚠️ warning: breaking change of healthchecks
*  chore: change default skipMTLS flag [#849](https://github.com/kumahq/kuma/pull/849)
⚠️ warning: breaking change of metrics

Breaking changes:
* ⚠️ This release removes [Passive Health Check](https://kuma.io/docs/0.5.1/policies/health-check/) in favor of [Circuit Breaking](https://kuma.io/docs/0.6.0/policies/circuit-breaker/). Please refer to [UPGRADE.md](UPGRADE.md).
* ⚠️ This release requires Prometheus to be a part of the mesh by default, if MTLs is enabled
* ⚠️ The previously deprecated Interface field is now removed. 

## [0.5.1]

> Released on  2020/06/03

Changes:

*  chore: Prometheus overrides on Kubernetes [#808](https://github.com/kumahq/kuma/pull/808) 
*  feat: Prometheus metrics over mTLS [#793](https://github.com/kumahq/kuma/pull/793) 
*  feat: GUI build for 0.5.1 [#785](https://github.com/kumahq/kuma/pull/785)
*  feat: circuit breaker [#751](https://github.com/kumahq/kuma/pull/751)[#781](https://github.com/kumahq/kuma/pull/781)
*  feat: CA rotation time supports months and year [#750](https://github.com/kumahq/kuma/pull/750)
[#794](https://github.com/kumahq/kuma/pull/794) 
 👍contributed by @tharun208
*  feat: send start signal [#783](https://github.com/kumahq/kuma/pull/783) 
*  fix: mesh delete validation [#770](https://github.com/kumahq/kuma/pull/770) 
*  feat: Improve certificate verification [#779](https://github.com/kumahq/kuma/pull/779) 
*  feat: generate cert with multiple SAN URIs [#774](https://github.com/kumahq/kuma/pull/774) 
*  fix: reject conflicting bootstrap when AdminPort is set [#758](https://github.com/kumahq/kuma/pull/758) 
*  feat: limit number subscription [#747](https://github.com/kumahq/kuma/pull/747) 
*  fix: OpenShift owner role [#780](https://github.com/kumahq/kuma/pull/780) 
*  chore: refactor cluster generation [#752](https://github.com/kumahq/kuma/pull/752)
*  feat: secrets delete validation [#746](https://github.com/kumahq/kuma/pull/746)
*  fix: allow slash validation so standard K8S tags are supported [#762](https://github.com/kumahq/kuma/pull/762)
*  feat: direct access to services and support for Headless Service [#749](https://github.com/kumahq/kuma/pull/749) [#790](https://github.com/kumahq/kuma/pull/790) 
*  feat: owners for Dataplane on k8s [#742](https://github.com/kumahq/kuma/pull/742) 
*  chore: updating Alpine to 3.11 [#672](https://github.com/kumahq/kuma/pull/672)

NOTE:

⚠️ This release introduces [Circuit Breaking](https://kuma.io/docs/0.5.1/policies/circuit-breaker/) as a superior alternative to [Passive Health Check](https://kuma.io/docs/0.5.1/policies/health-check/). The latter will be deprecated in 0.6.0. Please consider migrating your deployments.

## [0.5.0]

> Released on 2020/05/12

Changes:

* feat: configure expiration and rsa bits of the CA
  [#730](https://github.com/kumahq/kuma/pull/730)
* feat: provide `total` field when listing resources in the HTTP API
  [#723](https://github.com/kumahq/kuma/pull/723)
* fix: turn off transparent proxy for prometheus scraping
  [#733](https://github.com/kumahq/kuma/pull/733)  
* feat: dataplane certificate rotation
  [#721](https://github.com/kumahq/kuma/pull/721)
  [#722](https://github.com/kumahq/kuma/pull/722)
  [#739](https://github.com/kumahq/kuma/pull/739)
* сhore: update k8s to 1.18
  [#720](https://github.com/kumahq/kuma/pull/720)
* chore: update go up to 1.14.2
  [#718](https://github.com/kumahq/kuma/pull/718)
* feat: added age column for get commands and updated `inspect dataplanes` lastConnected and lastUpdated to the new format. 
  [#702](https://github.com/kumahq/kuma/pull/702)
  👍contributed by @tharun208
* chore: upgrade Envoy up to v1.14.1
  [#705](https://github.com/kumahq/kuma/pull/705)
* feat: friendly response in K8s mode
  [#712](https://github.com/kumahq/kuma/pull/712)  
* chore: upgrade go-control-plane up to v0.9.5
  [#707](https://github.com/kumahq/kuma/pull/707)
* fix: change the config to kuma-cp.conf.yml
  [#716](https://github.com/kumahq/kuma/pull/716)
* fix: kuma-cp migrate help text
  [#713](https://github.com/kumahq/kuma/pull/713)
  👍contributed by @tharun208
* fix: envoy binary not found
  [#695](https://github.com/kumahq/kuma/pull/695)
  👍contributed by @tharun208
* feat: merge injector into kuma-cp
  [#701](https://github.com/kumahq/kuma/pull/701)
* feat: refactor other pars of the Mesh to be consistent with CA
  [#704](https://github.com/kumahq/kuma/pull/704)
  ⚠️ warning: breaking change of Mesh model
* feat: secret validation on K8S
  [#696](https://github.com/kumahq/kuma/pull/696)
* feat: include traffic direction in access log
  [#682](https://github.com/kumahq/kuma/pull/682)
  👍contributed by @tharun208
* feat: validate tags and selectors
  [#691](https://github.com/kumahq/kuma/pull/691) 
* feat: refactor CA to plugins
  [#694](https://github.com/kumahq/kuma/pull/694)
* feat: expose CreationTime and modificationTime
  [#677](https://github.com/kumahq/kuma/pull/677)
  👍contributed by @tharun208
* feat: secret management API
  [#684](https://github.com/kumahq/kuma/pull/684)
  [#735](https://github.com/kumahq/kuma/pull/735)
* docs: adopting CNCF code of conduct
  [#692](https://github.com/kumahq/kuma/pull/692)
* chore: updating to version 1.1.17
  [#688](https://github.com/kumahq/kuma/pull/688)
* feat: CNI plugin for openshift support
  [#681](https://github.com/kumahq/kuma/pull/681)
  [#689](https://github.com/kumahq/kuma/pull/689)
* chore: removing tcp-echo
  [#671](https://github.com/kumahq/kuma/pull/671)
* feat: pagination in the API and kumactl
  [#673](https://github.com/kumahq/kuma/pull/673)
  [#690](https://github.com/kumahq/kuma/pull/690)
* chore: unify matching for TrafficPermission
  [#668](https://github.com/kumahq/kuma/pull/668)
  ⚠️ warning: breaking change of matching mechanism
* fix: reduce Prometheus scrape_interval to work with Kong Prometheus plugin 
  [#674](https://github.com/kumahq/kuma/pull/674)
* feat: added `kumactl get` command for individual resources
  [#667](https://github.com/kumahq/kuma/pull/667)
  👍contributed by @tharun208
* feat: kuma-dp and kumactl can communiate with kuma-cp over https
  [#633](https://github.com/kumahq/kuma/pull/633)
  👍contributed by @sudeeptoroy
* docs: introducing open-governance to the project
  [#659](https://github.com/kumahq/kuma/pull/659)
* feat: added logging and tracing information for meshes
  [#665](https://github.com/kumahq/kuma/pull/665)
  👍contributed by @tharun208
* feat: endpoints for fetching resources from all meshes 
  [#657](https://github.com/kumahq/kuma/pull/657)
* feature: validate `<port>.service.kuma.io/protocol` annotations on K8S Service objects
  [#611](https://github.com/kumahq/kuma/pull/611)
* feature: filter gateway dataplanes through api and through `kumactl inspect dataplanes --gateway`
  [#654](https://github.com/kumahq/kuma/pull/654)
  👍contributed by @tharun208
* fix: added shorthand command name for mesh in kumactl
  [#664](https://github.com/kumahq/kuma/pull/664)
  👍contributed by @tharun208
* feat: added a new `kumactl install tracing` CLI command
  [#655](https://github.com/kumahq/kuma/pull/655)
* chore: prevent dataplane creation with a headless services and provide more descriptive error message on pod converter error
  [#651](https://github.com/kumahq/kuma/pull/651)
* chore: migrate deprecated Envoy config to support newest version of Envoy 
  [#652](https://github.com/kumahq/kuma/pull/652)
* chore: replace deprected field ORIGINAL_DST_LB to CLUSTER_PROVIDED 
  [#656](https://github.com/kumahq/kuma/pull/656)
  👍contributed by @Lynskylate
* feat: save service's tags to header for L7-traffic
  [#647](https://github.com/kumahq/kuma/pull/647/files)
* chore: the API root `/` now returns the hostname
  [#645](https://github.com/kumahq/kuma/pull/645) 
* feat: FaultInjection policy
  [#643](https://github.com/kumahq/kuma/pull/643)
  [#649](https://github.com/kumahq/kuma/pull/649)
  [#734](https://github.com/kumahq/kuma/pull/734)
* feat: add response flag to default format
  [#635](https://github.com/kumahq/kuma/pull/635)
* chore: merge mTLS and CA status into one column
  [#637](https://github.com/kumahq/kuma/pull/637)
* fix: `kumactl apply -v ...` support dots in variables name
  [#636](https://github.com/kumahq/kuma/pull/636)
* feat: read only cached manager
  [#634](https://github.com/kumahq/kuma/pull/634)
* fix: explicitly set parameters in securityContext of kuma-init
  [#631](https://github.com/kumahq/kuma/pull/631)
* feature: log requests to external services
  [#630](https://github.com/kumahq/kuma/pull/630)
* feature: added flag `--dry-run` for `kumactl apply`
  [#622](https://github.com/kumahq/kuma/pull/622)
* feat: add the mesh to the access logs - http and network 
  [#620](https://github.com/kumahq/kuma/pull/620)
  👍contributed by @pradeepmurugesan

Breaking changes:
* ⚠️ Mesh can now have multiple CAs of the same type. Also it can use CA loaded as a plugins. For migration details, please refer to [UPGRADE.md](UPGRADE.md)

* ⚠️ before the change TrafficPermission worked in cumulative way, which means that all policies that matched a connection were applied.
  We changed TrafficPermission to work like every other policy so only "the most specific" matching policy is chosen.
  Consult [docs](https://kuma.io/docs/0.4.0/policies/how-kuma-chooses-the-right-policy-to-apply/) to learn more how Kuma picks the right policy.
  [668](https://github.com/kumahq/kuma/pull/668)

## [0.4.0]

> Released on 2020/02/28

Changes:

* feature: added a `Traffic Traces` page to `Kuma GUI`
  [#610](https://github.com/kumahq/kuma/pull/610)
* feature: added styling for `Tags` column on the `Dataplanes` page in `Kuma GUI`
  [#610](https://github.com/kumahq/kuma/pull/610)
* feature: improved data loading experience in `Kuma GUI`
  [#610](https://github.com/kumahq/kuma/pull/610)
* feature: on `k8s`, when a Dataplane cannot be generated automatically for a particular `Pod`, emit `k8s` `Events` to make the error state apparent to a user
  [#609](https://github.com/kumahq/kuma/pull/609)
* feature: include `k8s` namespace into a set of labels that describe a `Dataplane` to `Prometheus`
  [#601](https://github.com/kumahq/kuma/pull/601)
* feature: provision Grafana with Kuma Dashboards
  [#608](https://github.com/kumahq/kuma/pull/608)
* feature: add support for `kuma.io/sidecar-injection: disabled` annotation on `Pods` to let users selectively opt out of side-car injection on `k8s`
  [#607](https://github.com/kumahq/kuma/pull/607)
* fix: remove the requirement to a `Pod` to explicitly list container ports in a case where a `Service` defines target port by number
  [#605](https://github.com/kumahq/kuma/pull/605)
* feature: kumactl install metrics for one line Prometheus and Grafana install on K8S
  [#604](https://github.com/kumahq/kuma/pull/604)
* feature: order of meta in REST Resource JSON 
  [#600](https://github.com/kumahq/kuma/pull/600)
* feature: extend embedded gRPC Access Log Server to support the entire Envoy access log format
  [#595](https://github.com/kumahq/kuma/pull/595)
* feature: generate HTTP-specific configuration of access log
  [#590](https://github.com/kumahq/kuma/pull/590)
* feature: add support for Kuma-specific placeholders, such as `%KUMA_SOURCE_SERVICE%`, inside Envoy access log format
  [#594](https://github.com/kumahq/kuma/pull/594)
* feature: add support for the entire Envoy access log command operator syntax
  [#589](https://github.com/kumahq/kuma/pull/589)
* feature: generate tracing configuration in boostrap configuration
  [#592](https://github.com/kumahq/kuma/pull/592)
* feature: generate tracing configuration on listeners
  [#591](https://github.com/kumahq/kuma/pull/591)
* chore: generify proxy template matching (it now supports Gateway dataplane and '*' selector)
  [#588](https://github.com/kumahq/kuma/pull/588)
* feature: generate HTTP-specific outbound listeners for services tagged with `protocol: http`
  [#585](https://github.com/kumahq/kuma/pull/585)
* feature: TracingTrace in kumactl
  [#584](https://github.com/kumahq/kuma/pull/584)
* feature: TracingTrace in Kuma REST API
  [#583](https://github.com/kumahq/kuma/pull/583)
* feature: TracingTrace entity
  [#582](https://github.com/kumahq/kuma/pull/582)
* feature: Tracing section in Mesh entity
  [#581](https://github.com/kumahq/kuma/pull/581)
* chore: use new Dataplane format across the project
  [#580](https://github.com/kumahq/kuma/pull/580)
* feature: support new format of the Dataplane including scraping metrics from Gateway Dataplane
  [#579](https://github.com/kumahq/kuma/pull/579)
* feature: new Dataplane format
  [#578](https://github.com/kumahq/kuma/pull/578)
* feature: validate value of `protocol` tag on a Dataplane resource
  [#576](https://github.com/kumahq/kuma/pull/576)
* feature: support `<port>.service.kuma.io/protocol` annotation on k8s as a way for users to indicate protocol of a service
  [#575](https://github.com/kumahq/kuma/pull/575)
* feature: generate HTTP-specific inbound listeners for services tagged with `protocol: http`
  [#574](https://github.com/kumahq/kuma/pull/574)
* feature: support IPv6 in Dataplane resource
  [#567](https://github.com/kumahq/kuma/pull/567)
* fix: separate tcp access logs with a new line
  [#566](https://github.com/kumahq/kuma/pull/566)
* feature: validate certificates that users want to use as a `provided` CA
  [#565](https://github.com/kumahq/kuma/pull/565)
* fix: add MADS port to K8S install script
  [#564](https://github.com/kumahq/kuma/pull/564)
* feature: sanitize metrics for StatsD and Prometheus
  [#562](https://github.com/kumahq/kuma/pull/562)
* feature: reformat some Envoy metrics available in Prometheus
  [#558](https://github.com/kumahq/kuma/pull/558)
* feature: make maximum number of open connections to Postgres configurable
  [#557](https://github.com/kumahq/kuma/pull/557)
* feature: DB migrations for Postgres
  [#552](https://github.com/kumahq/kuma/pull/552)
* feature: order matching policies by creation time
  [#522](https://github.com/kumahq/kuma/pull/522)
* feature: add creation and modification time to core entities
  [#521](https://github.com/kumahq/kuma/pull/521)

## [0.3.2]

> Released on 2020/01/10

A new `Kuma` release that brings in many highly-requested features:

* **support for ingress traffic into the service mesh** - it is now possible to re-use
  existing, feature-rich `API Gateway` solutions at the front doors of
  your service mesh.
  E.g., check out our [instructions](https://kuma.io/docs/0.3.2/documentation/#gateway) how to leverage `Kuma` and [Kong](https://github.com/Kong/kong) together. Or, if you're a hands-on kind of person, play with our demos for [kubernetes](https://github.com/kumahq/kuma-demo/tree/master/kubernetes) and [universal](https://github.com/kumahq/kuma-demo/tree/master/vagrant).
* **access to Prometheus metrics collected by individual dataplanes** (Envoys) -
  as a user, you only need to enable `Prometheus` metrics as part of your `Mesh` policy,
  and that's it - every dataplane (Envoy) will automatically make its metrics available for scraping. Read more about it in the [docs](https://kuma.io/docs/0.3.2/policies/#traffic-metrics).
* **native integration with Prometheus auto-discovery** - be it `kubernetes` or `universal` (😮), `Prometheus` will automatically find all dataplanes in your mesh and scrape metrics out of them. Sounds interesting? See our [docs](https://kuma.io/docs/0.3.2/policies/#traffic-metrics) and play with our demos for [kubernetes](https://github.com/kumahq/kuma-demo/tree/master/kubernetes) and [universal](https://github.com/kumahq/kuma-demo/tree/master/vagrant).
* **brand new Kuma GUI** - following the very first preview release, `Kuma GUI` have been significantly overhauled to include more features, like support for every Kuma policy. Read more about it in the [docs](https://kuma.io/docs/0.3.2/documentation/#gui), see it live as part of our demos for [kubernetes](https://github.com/kumahq/kuma-demo/tree/master/kubernetes) and [universal](https://github.com/kumahq/kuma-demo/tree/master/vagrant).

Changes:

* feature: enable proxying of Kuma REST API via Kuma GUI
  [#542](https://github.com/kumahq/kuma/pull/542)
* feature: add a brand new version of Kuma GUI
  [#538](https://github.com/kumahq/kuma/pull/538)
* feature: add support for `MonitoringAssignment`s with arbitrary `Target` labels (rather than only `__address__`) to `kuma-prometheus-sd`
  [#540](https://github.com/kumahq/kuma/pull/540)
* feature: on `kuma-prometheus-sd` start-up, check write permissions on the output dir
  [#539](https://github.com/kumahq/kuma/pull/539)
* feature: implement MADS xDS client and integrate `kuma-prometheus-sd` with `Prometheus` via `file_sd` discovery
  [#537](https://github.com/kumahq/kuma/pull/537)
* feature: add configuration options to `kuma-prometheus-sd run`
  [#536](https://github.com/kumahq/kuma/pull/536)
* feature: add `kuma-prometheus-sd` binary
  [#535](https://github.com/kumahq/kuma/pull/535)
* feature: advertise MonitoringAssignment server via API Catalog
  [#534](https://github.com/kumahq/kuma/pull/534)
* feature: generate MonitoringAssignment for each Dataplane in a Mesh
  [#532](https://github.com/kumahq/kuma/pull/532)
* feature: add a Monitoring Assignment Discovery Service (MADS) server
  [#531](https://github.com/kumahq/kuma/pull/531)
* feature: add a generic watchdog for xDS streams
  [#530](https://github.com/kumahq/kuma/pull/530)
* feature: add a generic versioner for xDS Snapshots
  [#529](https://github.com/kumahq/kuma/pull/529)
* feature: add a custom version of SnapshotCache that supports arbitrary xDS resources
  [#528](https://github.com/kumahq/kuma/pull/528)
* feature: add proto definition for Monitoring Assignment Discovery Service (MADS)
  [#525](https://github.com/kumahq/kuma/pull/525)
* feature: enable Envoy Admin API by default with an option to opt out
  [#523](https://github.com/kumahq/kuma/pull/523)
* feature: add integration with Prometheus on K8S
  [#524](https://github.com/kumahq/kuma/pull/524)
* feature: redirect requests to /api path on GUI server to API Server
  [#520](https://github.com/kumahq/kuma/pull/520)
* feature: generate Envoy configuration that exposes Prometheus metrics
  [#510](https://github.com/kumahq/kuma/pull/510)
* feature: make port of Envoy Admin API available to Envoy config generators
  [#508](https://github.com/kumahq/kuma/pull/508)
* feature: add option to run dataplane as a gateway without inbounds
  [#503](https://github.com/kumahq/kuma/pull/503)
* feature: add `METRICS` column to the table output of `kumactl get meshes` to make it visible whether Prometheus settings have been configured
  [#502](https://github.com/kumahq/kuma/pull/502)
* feature: automatically set default values for Prometheus settings in the Mesh resource
  [#501](https://github.com/kumahq/kuma/pull/501)
* feature: add proto definitions for metrics that should be collected and exposed by dataplanes
  [#500](https://github.com/kumahq/kuma/pull/500)
* chore: encapsulate proxy init into kuma-init container
  [#495](https://github.com/kumahq/kuma/pull/495)
* feature: display CA type in kumactl get meshes
  [#494](https://github.com/kumahq/kuma/pull/494)
* chore: update Envoy to v1.12.2
  [#493](https://github.com/kumahq/kuma/pull/493)

Breaking changes:

* ⚠️ An `--dataplane-init-version` argument was removed. Init container was changed to `kuma-init` which version is in sync with the rest of the Kuma containers.

## [0.3.1]

> Released on 2019/12/13

Changes:

* feature: added Kuma UI
  [#461](https://github.com/kumahq/kuma/pull/461)
* feature: support TLS in Postgres-based storage backend
  [#472](https://github.com/kumahq/kuma/pull/472)
* feature: prevent removal of a signing certificate from a "provided" CA in use
  [#490](https://github.com/kumahq/kuma/pull/490)
* feature: validate consistency of changes to "provided" CA on `k8s`
  [#485](https://github.com/kumahq/kuma/pull/485)
* feature: validate consistency of changes to "provided" CA on `universal`
  [#475](https://github.com/kumahq/kuma/pull/475)
* feature: add `kumactl manage ca` commands to support "provided" CA
  [#474](https://github.com/kumahq/kuma/pull/474)
  ⚠️ warning: api breaking change
* feature: include health checks into generated Envoy configuration (#483)
  [#483](https://github.com/kumahq/kuma/pull/483)
* feature: pick a single the most specific `HealthCheck` for every service reachable from a given `Dataplane`
  [#481](https://github.com/kumahq/kuma/pull/481)
* feature: add REST API for managing "provided" CA
  [#473](https://github.com/kumahq/kuma/pull/473)
* feature: reuse policy matching logic for `TrafficLog` resource
  [#482](https://github.com/kumahq/kuma/pull/482)
  ⚠️ warning: backwards-incompatible change of behaviour
* feature: refactor policy matching logic into reusable function
  [#479](https://github.com/kumahq/kuma/pull/479)
* feature: add `kumactl get healthchecks` command
  [#477](https://github.com/kumahq/kuma/pull/477)
* feature: validate `HealthCheck` resource
  [#476](https://github.com/kumahq/kuma/pull/476)
* feature: add `HealthCheck` CRD on kubernetes
  [#471](https://github.com/kumahq/kuma/pull/471)
* feature: add `HealthCheck` to core model
  [#470](https://github.com/kumahq/kuma/pull/470)
* feature: add proto definition for `HealthCheck` resource
  [#446](https://github.com/kumahq/kuma/pull/446)
* feature: ground work for "provided" CA support
  [#467](https://github.com/kumahq/kuma/pull/467)
* feature: remove "namespace" from core model
  [#458](https://github.com/kumahq/kuma/pull/458)
  ⚠️ warning: api breaking change
* feature: expose effective configuration of `kuma-cp` as part of REST API
  [#454](https://github.com/kumahq/kuma/pull/454)
* feature: improve error messages in `kumactl config control-planes add`
  [#455](https://github.com/kumahq/kuma/pull/455)
* feature: delete resource operation should return 404 if resource is not found
  [#450](https://github.com/kumahq/kuma/pull/450)
* feature: autoconfigure bootstrap server on `kuma-cp` startup
  [#449](https://github.com/kumahq/kuma/pull/449)
* feature: update envoy to v1.12.1
  [#448](https://github.com/kumahq/kuma/pull/448)

Breaking changes:
* ⚠️ a few arguments of `kumactl config control-planes add` have been renamed: `--dataplane-token-client-cert => --admin-client-cert` and `--dataplane-token-client-key => --admin-client-key`
  [474](https://github.com/kumahq/kuma/pull/474)
* ⚠️ instead of applying all matching `TrafficLog` policies to a given `outbound` interface of a `Dataplane`, only a single the most specific `TrafficLog` policy is now applied
  [#482](https://github.com/kumahq/kuma/pull/482)
* ⚠️ `Mesh` CRD on Kubernetes is now Cluster-scoped
  [#458](https://github.com/kumahq/kuma/pull/458)

## [0.3.0]

> Released on 2019/11/18

Changes:

* fix: fixed discrepancy between `ProxyTemplate` documentation and actual implementation
  [#422](https://github.com/kumahq/kuma/pull/422)
* chore: dropped support for `Mesh`-wide logging settings
  [#438](https://github.com/kumahq/kuma/pull/438)
  ⚠️ warning: api breaking change
* feature: validate `ProxyTemplate` resource on CREATE/UPDATE in universal mode
  [#431](https://github.com/kumahq/kuma/pull/431)
  ⚠️ warning: api breaking change
* feature: add `kumactl generate tls-certificate` command
  [#437](https://github.com/kumahq/kuma/pull/437)
* feature: validate `TrafficLog` resource on CREATE/UPDATE in universal mode
  [#435](https://github.com/kumahq/kuma/pull/435)
* feature: validate `TrafficPermission` resource on CREATE/UPDATE in universal mode
  [#436](https://github.com/kumahq/kuma/pull/436)
* feature: dropped support for multiple rules per single `TrafficPermission` resource
  [#434](https://github.com/kumahq/kuma/pull/434)
  ⚠️ warning: api breaking change
* feature: added configuration for Kuma UI
  [#428](https://github.com/kumahq/kuma/pull/428)
* feature: included Kuma UI into `kuma-cp`
  [#410](https://github.com/kumahq/kuma/pull/410)
* feature: dropped support for multiple rules per single `TrafficLog` resource
  [#433](https://github.com/kumahq/kuma/pull/433)
  ⚠️ warning: api breaking change
* feature: validate `Mesh` resource on CREATE/UPDATE in universal mode
  [#430](https://github.com/kumahq/kuma/pull/430)
* feature: `kumactl` commands now do custom formating of errors returned by the Kuma REST API
  [#411](https://github.com/kumahq/kuma/pull/411)
* feature: `tcp_proxy` configuration now routes to a list of weighted clusters according to `TrafficRoute`
  [#423](https://github.com/kumahq/kuma/pull/423)
* feature: included tags of a dataplane into `ClusterLoadAssignment`
  [#422](https://github.com/kumahq/kuma/pull/422)
* feature: validate Kuma CRDs on Kubernetes
  [#401](https://github.com/kumahq/kuma/pull/401)
* feature: improved feedback given to a user when `kuma-dp run` is configured with an invalid dataplane token
  [#418](https://github.com/kumahq/kuma/pull/418)
* release: included Docker image with `kumactl` into release build
  [#425](https://github.com/kumahq/kuma/pull/425)
* feature: support enabling/disabling DataplaneToken server via a configuration flag
  [#415](https://github.com/kumahq/kuma/pull/415)
* feature: pick a single the most specific `TrafficRoute` for every outbound interface of a `Dataplane`
  [#421](https://github.com/kumahq/kuma/pull/421)
* feature: validate `TrafficRoute` resource on CREATE/UPDATE in universal mode
  [#424](https://github.com/kumahq/kuma/pull/424)
* feature: `kumactl apply` can now download a resource from URL
  [#402](https://github.com/kumahq/kuma/pull/402)
* chore: migrated to the latest version of `go-control-plane`
  [#419](https://github.com/kumahq/kuma/pull/419)
* feature: added `kumactl get traffic-routes` command
  [#400](https://github.com/kumahq/kuma/pull/400)
* feature: added `TrafficRoute` CRD on Kubernetes
  [#398](https://github.com/kumahq/kuma/pull/398)
* feature: added `TrafficRoute` resource to core model
  [#397](https://github.com/kumahq/kuma/pull/397)
* feature: added support for CORS to Kuma REST API
  [#412](https://github.com/kumahq/kuma/pull/412)
* feature: validate `Dataplane` resource on CREATE/UPDATE in universal mode
  [#388](https://github.com/kumahq/kuma/pull/388)
* feature: added support for client certificate-based authentication to `kumactl generate dataplane-token` command
  [#372](https://github.com/kumahq/kuma/pull/372)
* feature: added `--overwrite` flag to the `kumactl config control-planes add` command
  [#381](https://github.com/kumahq/kuma/pull/381)
  👍contributed by @Gabitchov
* feature: added `MESH` column into the output of `kumactl get proxytemplates`
  [#399](https://github.com/kumahq/kuma/pull/399)
  👍contributed by @programmer04
* feature: `kuma-dp run` is now configured with a URL of the API server instead of a former URL of the boostrap config server
  [#417](https://github.com/kumahq/kuma/pull/417)
  ⚠️ warning: interface breaking change
* feature: added a REST endpoint to advertize location of various sub-components of the control plane
  [#369](https://github.com/kumahq/kuma/pull/369)
* feature: added protobuf descriptor for `TrafficRoute` resource
  [#396](https://github.com/kumahq/kuma/pull/396)
* fix: added reconciliation on Dataplane delete to handle a case where a user manually deletes Dataplane on Kubernetes
  [#392](https://github.com/kumahq/kuma/pull/392)
* feature: Kuma REST API on Kubernetes is now restricted to READ operations only
  [#377](https://github.com/kumahq/kuma/pull/377)
  👍contributed by @sterchelen
* fix: ignored errors in unit tests
  [#376](https://github.com/kumahq/kuma/pull/376)
  👍contributed by @alrs
* feature: JSON output of `kumactl` is now pretty-printed
  [#360](https://github.com/kumahq/kuma/pull/360)
  👍contributed by @sterchelen
* feature: DataplaneToken server is now exposed for remote access over HTTPS with mandatory client certificate-based authentication
  [#349](https://github.com/kumahq/kuma/pull/349)
* feature: `kuma-dp` now passes a path to a file with a dataplane token as an argumenent for bootstrap config API
  [#348](https://github.com/kumahq/kuma/pull/348)
* feature: added support for mTLS on Kubernetes v1.13+
  [#356](https://github.com/kumahq/kuma/pull/356)
* feature: added `kumactl delete` command
  [#343](https://github.com/kumahq/kuma/pull/343)
  👍contributed by @pradeepmurugesan
* feature: added `kumactl gerenerate dataplane-token` command
  [#342](https://github.com/kumahq/kuma/pull/342)
* feature: added a DataplaneToken server to support dataplane authentication in universal mode
  [#342](https://github.com/kumahq/kuma/pull/342)
* feature: on removal of a Mesh remove all policies defined in it
  [#332](https://github.com/kumahq/kuma/pull/332)
* docs: documented release process
  [#341](https://github.com/kumahq/kuma/pull/341)
* docs: DEVELOPER.md was brought up to date
  [#346](https://github.com/kumahq/kuma/pull/346)
* docs: added instructions how to deploy `kuma-demo` on Kubernetes
  [#347](https://github.com/kumahq/kuma/pull/347)

Community contributions from:

* 👍@pradeepmurugesan
* 👍@alrs
* 👍@sterchelen
* 👍@programmer04
* 👍@Gabitchov

Breaking changes:

* ⚠️ fixed discrepancy between `ProxyTemplate` documentation and actual implementation
  [#422](https://github.com/kumahq/kuma/pull/422)
* ⚠️ `selectors` in `ProxyTemplate` now always require `service` tag
  [#431](https://github.com/kumahq/kuma/pull/431)
* ⚠️ dropped support for `Mesh`-wide logging settings
  [#438](https://github.com/kumahq/kuma/pull/438)
* ⚠️ dropped support for multiple rules per single `TrafficPermission` resource
  [#434](https://github.com/kumahq/kuma/pull/434)
* ⚠️ dropped support for multiple rules per single `TrafficLog` resource
  [#433](https://github.com/kumahq/kuma/pull/433)
* ⚠️ value of `--cp-address` parameter in `kuma-dp run` is now a URL of the API server instead of a former URL of the boostrap config server
  [#417](https://github.com/kumahq/kuma/pull/417)

## [0.2.2]

> Released on 2019/10/11

Changes:

* Draining time is now configurable
  [#310](https://github.com/kumahq/kuma/pull/310)
* Validation that Control Plane is running when adding it with `kumactl`
  [#181](https://github.com/kumahq/kuma/issues/181)
* Upgraded version of go-control-plane
* Upgraded version of Envoy to 1.11.2
* Connection timeout to ADS server is now configurable (part of `envoy` bootstrap config)
  [#340](https://github.com/kumahq/kuma/pull/340)

Fixed issues:
* Cluster never went out warming state
  [#331](https://github.com/kumahq/kuma/pull/331)
* SDS server didn't handle requests with empty resources list
  [#337](https://github.com/kumahq/kuma/pull/337) 

## [0.2.1]

> Released on 2019/10/03

Fixed issues:

* Issue with `Access Log Server` (integrated into `kuma-dp`) on k8s:
 `kuma-cp` was configuring Envoy to use a Unix socket other than
 `kuma-dp` was actually listening on
  [#307](https://github.com/kumahq/kuma/pull/307)

## [0.2.0]

> Released on 2019/10/02

Changes:

* Fix an issue with `Access Log Server` (integrated into `kuma-dp`) on Kubernetes
  by replacing `Google gRPC client` with `Envoy gRPC client`
  [#306](https://github.com/kumahq/kuma/pull/306)
* Settings of a `kuma-sidecar` container, such as `ReadinessProbe`, `LivenessProbe` and `Resources`,
  are now configurable
  [#304](https://github.com/kumahq/kuma/pull/304)
* Added support for `TCP` logging backends, such as `ELK` and `Splunk`
  [#300](https://github.com/kumahq/kuma/pull/300)
* `Builtin CA` on `Kubernetes` is now (re-)generated by a `Controller`
  [#299](https://github.com/kumahq/kuma/pull/299)
* Default `Mesh` on `Kubernetes` is now (re-)generated by a `Controller`
  [#298](https://github.com/kumahq/kuma/pull/298)
* Added `Kubernetes Admission WebHook` to apply defaults to `Mesh` resources
  [#297](https://github.com/kumahq/kuma/pull/297)
* Upgraded version of `kubernetes-sigs/controller-runtime` dependency
  [#293](https://github.com/kumahq/kuma/pull/293)
* Added a concept of `RuntimePlugin` to `kuma-cp`
  [#296](https://github.com/kumahq/kuma/pull/296)
* Updated `LDS` to configure `access_loggers` on `outbound` listeners
  according to `TrafficLog` resources
  [#276](https://github.com/kumahq/kuma/pull/276)
* Changed default locations where `kuma-dp` is looking for `envoy` binary
  [#268](https://github.com/kumahq/kuma/pull/268)
* Added model for `TrafficLog` resource with `File` as a logging backend
  [#266](https://github.com/kumahq/kuma/pull/266)
* Added `kumactl install database-schema` command to generate DB schema
  used by `kuma-cp` on `universal` environment
  [#236](https://github.com/kumahq/kuma/pull/236)
* Automated release of `Docker` images
  [#265](https://github.com/kumahq/kuma/pull/265)
* Changed default location where auto-generated Envoy bootstrap configuration is saved to
  [#261](https://github.com/kumahq/kuma/pull/261)
* Added support for multiple `kuma-dp` instances on a single Linux machine
  [#260](https://github.com/kumahq/kuma/pull/260)
* Automated release of `*.tar` artifacts
  [#250](https://github.com/kumahq/kuma/pull/250)

Fixed issues (user feedback):

* Dataplanes cannot connect to a non-default Mesh with mTLS enabled on k8s
  [262](https://github.com/kumahq/kuma/issues/262)
* Starting multiple services on the same Linux machine
  [254](https://github.com/kumahq/kuma/issues/254)
* Fallback when invoking `envoy` from `kuma-dp`
  [249](https://github.com/kumahq/kuma/issues/249)

## [0.1.2]

> Released on 2019/09/11

* Upgraded version of Go to address CVE-2019-14809.
  [#248](https://github.com/kumahq/kuma/pull/248)
* Improved support for mTLS on `kubernetes`.
  [#238](https://github.com/kumahq/kuma/pull/238)

## [0.1.1]

> Released on 2019/09/10

* Bugfix in the distribution process that caused `kumactl install control-plane` to not work properly.

## [0.1.0]

> Released on 2019/09/10

The main features of this release are:

* Multi-Tenancy: With the `Mesh` entity.
* Platform-Agnosticity: With `universal` and `kubernetes` modes.
* Mutual TLS: By setting mtls property in Mesh.
* Logging: By setting the logging property in Mesh.
* Traffic Permissions: With the `TrafficPermission` entity.
* Proxy Templating: For low-level Envoy configuration via the `ProxyTemplate` entity.
