# CHANGELOG

## master

## [0.5.0]

> Released on 2020/05/12

Changes:

* feat: configure expiration and rsa bits of the CA
  [#730](https://github.com/Kong/kuma/pull/730)
* feat: provide `total` field when listing resources in the HTTP API
  [#723](https://github.com/Kong/kuma/pull/723)
* fix: turn off transparent proxy for prometheus scraping
  [#733](https://github.com/Kong/kuma/pull/733)  
* feat: dataplane certificate rotation
  [#721](https://github.com/Kong/kuma/pull/721)
  [#722](https://github.com/Kong/kuma/pull/722)
  [#739](https://github.com/Kong/kuma/pull/739)
* сhore: update k8s to 1.18
  [#720](https://github.com/Kong/kuma/pull/720)
* chore: update go up to 1.14.2
  [#718](https://github.com/Kong/kuma/pull/718)
* feat: added age column for get commands and updated `inspect dataplanes` lastConnected and lastUpdated to the new format. 
  [#702](https://github.com/Kong/kuma/pull/702)
  👍contributed by @tharun208
* chore: upgrade Envoy up to v1.14.1
  [#705](https://github.com/Kong/kuma/pull/705)
* feat: friendly response in K8s mode
  [#712](https://github.com/Kong/kuma/pull/712)  
* chore: upgrade go-control-plane up to v0.9.5
  [#707](https://github.com/Kong/kuma/pull/707)
* fix: change the config to kuma-cp.conf.yml
  [#716](https://github.com/Kong/kuma/pull/716)
* fix: kuma-cp migrate help text
  [#713](https://github.com/Kong/kuma/pull/713)
  👍contributed by @tharun208
* fix: envoy binary not found
  [#695](https://github.com/Kong/kuma/pull/695)
  👍contributed by @tharun208
* feat: merge injector into kuma-cp
  [#701](https://github.com/Kong/kuma/pull/701)
* feat: refactor other pars of the Mesh to be consistent with CA
  [#704](https://github.com/Kong/kuma/pull/704)
  ⚠️ warning: breaking change of Mesh model
* feat: secret validation on K8S
  [#696](https://github.com/Kong/kuma/pull/696)
* feat: include traffic direction in access log
  [#682](https://github.com/Kong/kuma/pull/682)
  👍contributed by @tharun208
* feat: validate tags and selectors
  [#691](https://github.com/Kong/kuma/pull/691) 
* feat: refactor CA to plugins
  [#694](https://github.com/Kong/kuma/pull/694)
* feat: expose CreationTime and modificationTime
  [#677](https://github.com/Kong/kuma/pull/677)
  👍contributed by @tharun208
* feat: secret management API
  [#684](https://github.com/Kong/kuma/pull/684)
  [#735](https://github.com/Kong/kuma/pull/735)
* docs: adopting CNCF code of conduct
  [#692](https://github.com/Kong/kuma/pull/692)
* chore: updating to version 1.1.17
  [#688](https://github.com/Kong/kuma/pull/688)
* feat: CNI plugin for openshift support
  [#681](https://github.com/Kong/kuma/pull/681)
  [#689](https://github.com/Kong/kuma/pull/689)
* chore: removing tcp-echo
  [#671](https://github.com/Kong/kuma/pull/671)
* feat: pagination in the API and kumactl
  [#673](https://github.com/Kong/kuma/pull/673)
  [#690](https://github.com/Kong/kuma/pull/690)
* chore: unify matching for TrafficPermission
  [#668](https://github.com/Kong/kuma/pull/668)
  ⚠️ warning: breaking change of matching mechanism
* fix: reduce Prometheus scrape_interval to work with Kong Prometheus plugin 
  [#674](https://github.com/Kong/kuma/pull/674)
* feat: added `kumactl get` command for individual resources
  [#667](https://github.com/Kong/kuma/pull/667)
  👍contributed by @tharun208
* feat: kuma-dp and kumactl can communiate with kuma-cp over https
  [#633](https://github.com/Kong/kuma/pull/633)
  👍contributed by @sudeeptoroy
* docs: introducing open-governance to the project
  [#659](https://github.com/Kong/kuma/pull/659)
* feat: added logging and tracing information for meshes
  [#665](https://github.com/Kong/kuma/pull/665)
  👍contributed by @tharun208
* feat: endpoints for fetching resources from all meshes 
  [#657](https://github.com/Kong/kuma/pull/657)
* feature: validate `<port>.service.kuma.io/protocol` annotations on K8S Service objects
  [#611](https://github.com/Kong/kuma/pull/611)
* feature: filter gateway dataplanes through api and through `kumactl inspect dataplanes --gateway`
  [#654](https://github.com/Kong/kuma/pull/654)
  👍contributed by @tharun208
* fix: added shorthand command name for mesh in kumactl
  [#664](https://github.com/Kong/kuma/pull/664)
  👍contributed by @tharun208
* feat: added a new `kumactl install tracing` CLI command
  [#655](https://github.com/Kong/kuma/pull/655)
* chore: prevent dataplane creation with a headless services and provide more descriptive error message on pod converter error
  [#651](https://github.com/Kong/kuma/pull/651)
* chore: migrate deprecated Envoy config to support newest version of Envoy 
  [#652](https://github.com/Kong/kuma/pull/652)
* chore: replace deprected field ORIGINAL_DST_LB to CLUSTER_PROVIDED 
  [#656](https://github.com/Kong/kuma/pull/656)
  👍contributed by @Lynskylate
* feat: save service's tags to header for L7-traffic
  [#647](https://github.com/Kong/kuma/pull/647/files)
* chore: the API root `/` now returns the hostname
  [#645](https://github.com/Kong/kuma/pull/645) 
* feat: FaultInjection policy
  [#643](https://github.com/Kong/kuma/pull/643)
  [#649](https://github.com/Kong/kuma/pull/649)
  [#734](https://github.com/Kong/kuma/pull/734)
* feat: add response flag to default format
  [#635](https://github.com/Kong/kuma/pull/635)
* chore: merge mTLS and CA status into one column
  [#637](https://github.com/Kong/kuma/pull/637)
* fix: `kumactl apply -v ...` support dots in variables name
  [#636](https://github.com/Kong/kuma/pull/636)
* feat: read only cached manager
  [#634](https://github.com/Kong/kuma/pull/634)
* fix: explicitly set parameters in securityContext of kuma-init
  [#631](https://github.com/Kong/kuma/pull/631)
* feature: log requests to external services
  [#630](https://github.com/Kong/kuma/pull/630)
* feature: added flag `--dry-run` for `kumactl apply`
  [#622](https://github.com/Kong/kuma/pull/622)
* feat: add the mesh to the access logs - http and network 
  [#620](https://github.com/Kong/kuma/pull/620)
  👍contributed by @pradeepmurugesan

Breaking changes:
* ⚠ Mesh can now have multiple CAs of the same type. Also it can use CA loaded as a plugins. For migration details, please refer to [UPGRADE.md](UPGRADE.md)

* ⚠️ before the change TrafficPermission worked in cumulative way, which means that all policies that matched a connection were applied.
  We changed TrafficPermission to work like every other policy so only "the most specific" matching policy is chosen.
  Consult [docs](https://kuma.io/docs/0.4.0/policies/how-kuma-chooses-the-right-policy-to-apply/) to learn more how Kuma picks the right policy.
  [668](https://github.com/Kong/kuma/pull/668)

## [0.4.0]

> Released on 2020/02/28

Changes:

* feature: added a `Traffic Traces` page to `Kuma GUI`
  [#610](https://github.com/Kong/kuma/pull/610)
* feature: added styling for `Tags` column on the `Dataplanes` page in `Kuma GUI`
  [#610](https://github.com/Kong/kuma/pull/610)
* feature: improved data loading experience in `Kuma GUI`
  [#610](https://github.com/Kong/kuma/pull/610)
* feature: on `k8s`, when a Dataplane cannot be generated automatically for a particular `Pod`, emit `k8s` `Events` to make the error state apparent to a user
  [#609](https://github.com/Kong/kuma/pull/609)
* feature: include `k8s` namespace into a set of labels that describe a `Dataplane` to `Prometheus`
  [#601](https://github.com/Kong/kuma/pull/601)
* feature: provision Grafana with Kuma Dashboards
  [#608](https://github.com/Kong/kuma/pull/608)
* feature: add support for `kuma.io/sidecar-injection: disabled` annotation on `Pods` to let users selectively opt out of side-car injection on `k8s`
  [#607](https://github.com/Kong/kuma/pull/607)
* fix: remove the requirement to a `Pod` to explicitly list container ports in a case where a `Service` defines target port by number
  [#605](https://github.com/Kong/kuma/pull/605)
* feature: kumactl install metrics for one line Prometheus and Grafana install on K8S
  [#604](https://github.com/Kong/kuma/pull/604)
* feature: order of meta in REST Resource JSON 
  [#600](https://github.com/Kong/kuma/pull/600)
* feature: extend embedded gRPC Access Log Server to support the entire Envoy access log format
  [#595](https://github.com/Kong/kuma/pull/595)
* feature: generate HTTP-specific configuration of access log
  [#590](https://github.com/Kong/kuma/pull/590)
* feature: add support for Kuma-specific placeholders, such as `%KUMA_SOURCE_SERVICE%`, inside Envoy access log format
  [#594](https://github.com/Kong/kuma/pull/594)
* feature: add support for the entire Envoy access log command operator syntax
  [#589](https://github.com/Kong/kuma/pull/589)
* feature: generate tracing configuration in boostrap configuration
  [#592](https://github.com/Kong/kuma/pull/592)
* feature: generate tracing configuration on listeners
  [#591](https://github.com/Kong/kuma/pull/591)
* chore: generify proxy template matching (it now supports Gateway dataplane and '*' selector)
  [#588](https://github.com/Kong/kuma/pull/588)
* feature: generate HTTP-specific outbound listeners for services tagged with `protocol: http`
  [#585](https://github.com/Kong/kuma/pull/585)
* feature: TracingTrace in kumactl
  [#584](https://github.com/Kong/kuma/pull/584)
* feature: TracingTrace in Kuma REST API
  [#583](https://github.com/Kong/kuma/pull/583)
* feature: TracingTrace entity
  [#582](https://github.com/Kong/kuma/pull/582)
* feature: Tracing section in Mesh entity
  [#581](https://github.com/Kong/kuma/pull/581)
* chore: use new Dataplane format across the project
  [#580](https://github.com/Kong/kuma/pull/580)
* feature: support new format of the Dataplane including scraping metrics from Gateway Dataplane
  [#579](https://github.com/Kong/kuma/pull/579)
* feature: new Dataplane format
  [#578](https://github.com/Kong/kuma/pull/578)
* feature: validate value of `protocol` tag on a Dataplane resource
  [#576](https://github.com/Kong/kuma/pull/576)
* feature: support `<port>.service.kuma.io/protocol` annotation on k8s as a way for users to indicate protocol of a service
  [#575](https://github.com/Kong/kuma/pull/575)
* feature: generate HTTP-specific inbound listeners for services tagged with `protocol: http`
  [#574](https://github.com/Kong/kuma/pull/574)
* feature: support IPv6 in Dataplane resource
  [#567](https://github.com/Kong/kuma/pull/567)
* fix: separate tcp access logs with a new line
  [#566](https://github.com/Kong/kuma/pull/566)
* feature: validate certificates that users want to use as a `provided` CA
  [#565](https://github.com/Kong/kuma/pull/565)
* fix: add MADS port to K8S install script
  [#564](https://github.com/Kong/kuma/pull/564)
* feature: sanitize metrics for StatsD and Prometheus
  [#562](https://github.com/Kong/kuma/pull/562)
* feature: reformat some Envoy metrics available in Prometheus
  [#558](https://github.com/Kong/kuma/pull/558)
* feature: make maximum number of open connections to Postgres configurable
  [#557](https://github.com/Kong/kuma/pull/557)
* feature: DB migrations for Postgres
  [#552](https://github.com/Kong/kuma/pull/552)
* feature: order matching policies by creation time
  [#522](https://github.com/Kong/kuma/pull/522)
* feature: add creation and modification time to core entities
  [#521](https://github.com/Kong/kuma/pull/521)

## [0.3.2]

> Released on 2020/01/10

A new `Kuma` release that brings in many highly-requested features:

* **support for ingress traffic into the service mesh** - it is now possible to re-use
  existing, feature-rich `API Gateway` solutions at the front doors of
  your service mesh.
  E.g., check out our [instructions](https://kuma.io/docs/0.3.2/documentation/#gateway) how to leverage `Kuma` and [Kong](https://github.com/Kong/kong) together. Or, if you're a hands-on kind of person, play with our demos for [kubernetes](https://github.com/Kong/kuma-demo/tree/master/kubernetes) and [universal](https://github.com/Kong/kuma-demo/tree/master/vagrant).
* **access to Prometheus metrics collected by individual dataplanes** (Envoys) -
  as a user, you only need to enable `Prometheus` metrics as part of your `Mesh` policy,
  and that's it - every dataplane (Envoy) will automatically make its metrics available for scraping. Read more about it in the [docs](https://kuma.io/docs/0.3.2/policies/#traffic-metrics).
* **native integration with Prometheus auto-discovery** - be it `kubernetes` or `universal` (😮), `Prometheus` will automatically find all dataplanes in your mesh and scrape metrics out of them. Sounds interesting? See our [docs](https://kuma.io/docs/0.3.2/policies/#traffic-metrics) and play with our demos for [kubernetes](https://github.com/Kong/kuma-demo/tree/master/kubernetes) and [universal](https://github.com/Kong/kuma-demo/tree/master/vagrant).
* **brand new Kuma GUI** - following the very first preview release, `Kuma GUI` have been significantly overhauled to include more features, like support for every Kuma policy. Read more about it in the [docs](https://kuma.io/docs/0.3.2/documentation/#gui), see it live as part of our demos for [kubernetes](https://github.com/Kong/kuma-demo/tree/master/kubernetes) and [universal](https://github.com/Kong/kuma-demo/tree/master/vagrant).

Changes:

* feature: enable proxying of Kuma REST API via Kuma GUI
  [#542](https://github.com/Kong/kuma/pull/542)
* feature: add a brand new version of Kuma GUI
  [#538](https://github.com/Kong/kuma/pull/538)
* feature: add support for `MonitoringAssignment`s with arbitrary `Target` labels (rather than only `__address__`) to `kuma-prometheus-sd`
  [#540](https://github.com/Kong/kuma/pull/540)
* feature: on `kuma-prometheus-sd` start-up, check write permissions on the output dir
  [#539](https://github.com/Kong/kuma/pull/539)
* feature: implement MADS xDS client and integrate `kuma-prometheus-sd` with `Prometheus` via `file_sd` discovery
  [#537](https://github.com/Kong/kuma/pull/537)
* feature: add configuration options to `kuma-prometheus-sd run`
  [#536](https://github.com/Kong/kuma/pull/536)
* feature: add `kuma-prometheus-sd` binary
  [#535](https://github.com/Kong/kuma/pull/535)
* feature: advertise MonitoringAssignment server via API Catalog
  [#534](https://github.com/Kong/kuma/pull/534)
* feature: generate MonitoringAssignment for each Dataplane in a Mesh
  [#532](https://github.com/Kong/kuma/pull/532)
* feature: add a Monitoring Assignment Discovery Service (MADS) server
  [#531](https://github.com/Kong/kuma/pull/531)
* feature: add a generic watchdog for xDS streams
  [#530](https://github.com/Kong/kuma/pull/530)
* feature: add a generic versioner for xDS Snapshots
  [#529](https://github.com/Kong/kuma/pull/529)
* feature: add a custom version of SnapshotCache that supports arbitrary xDS resources
  [#528](https://github.com/Kong/kuma/pull/528)
* feature: add proto definition for Monitoring Assignment Discovery Service (MADS)
  [#525](https://github.com/Kong/kuma/pull/525)
* feature: enable Envoy Admin API by default with an option to opt out
  [#523](https://github.com/Kong/kuma/pull/523)
* feature: add integration with Prometheus on K8S
  [#524](https://github.com/Kong/kuma/pull/524)
* feature: redirect requests to /api path on GUI server to API Server
  [#520](https://github.com/Kong/kuma/pull/520)
* feature: generate Envoy configuration that exposes Prometheus metrics
  [#510](https://github.com/Kong/kuma/pull/510)
* feature: make port of Envoy Admin API available to Envoy config generators
  [#508](https://github.com/Kong/kuma/pull/508)
* feature: add option to run dataplane as a gateway without inbounds
  [#503](https://github.com/Kong/kuma/pull/503)
* feature: add `METRICS` column to the table output of `kumactl get meshes` to make it visible whether Prometheus settings have been configured
  [#502](https://github.com/Kong/kuma/pull/502)
* feature: automatically set default values for Prometheus settings in the Mesh resource
  [#501](https://github.com/Kong/kuma/pull/501)
* feature: add proto definitions for metrics that should be collected and exposed by dataplanes
  [#500](https://github.com/Kong/kuma/pull/500)
* chore: encapsulate proxy init into kuma-init container
  [#495](https://github.com/Kong/kuma/pull/495)
* feature: display CA type in kumactl get meshes
  [#494](https://github.com/Kong/kuma/pull/494)
* chore: update Envoy to v1.12.2
  [#493](https://github.com/Kong/kuma/pull/493)

Breaking changes:

* ⚠️ An `--dataplane-init-version` argument was removed. Init container was changed to `kuma-init` which version is in sync with the rest of the Kuma containers.

## [0.3.1]

> Released on 2019/12/13

Changes:

* feature: added Kuma UI
  [#461](https://github.com/Kong/kuma/pull/461)
* feature: support TLS in Postgres-based storage backend
  [#472](https://github.com/Kong/kuma/pull/472)
* feature: prevent removal of a signing certificate from a "provided" CA in use
  [#490](https://github.com/Kong/kuma/pull/490)
* feature: validate consistency of changes to "provided" CA on `k8s`
  [#485](https://github.com/Kong/kuma/pull/485)
* feature: validate consistency of changes to "provided" CA on `universal`
  [#475](https://github.com/Kong/kuma/pull/475)
* feature: add `kumactl manage ca` commands to support "provided" CA
  [#474](https://github.com/Kong/kuma/pull/474)
  ⚠️ warning: api breaking change
* feature: include health checks into generated Envoy configuration (#483)
  [#483](https://github.com/Kong/kuma/pull/483)
* feature: pick a single the most specific `HealthCheck` for every service reachable from a given `Dataplane`
  [#481](https://github.com/Kong/kuma/pull/481)
* feature: add REST API for managing "provided" CA
  [#473](https://github.com/Kong/kuma/pull/473)
* feature: reuse policy matching logic for `TrafficLog` resource
  [#482](https://github.com/Kong/kuma/pull/482)
  ⚠️ warning: backwards-incompatible change of behaviour
* feature: refactor policy matching logic into reusable function
  [#479](https://github.com/Kong/kuma/pull/479)
* feature: add `kumactl get healthchecks` command
  [#477](https://github.com/Kong/kuma/pull/477)
* feature: validate `HealthCheck` resource
  [#476](https://github.com/Kong/kuma/pull/476)
* feature: add `HealthCheck` CRD on kubernetes
  [#471](https://github.com/Kong/kuma/pull/471)
* feature: add `HealthCheck` to core model
  [#470](https://github.com/Kong/kuma/pull/470)
* feature: add proto definition for `HealthCheck` resource
  [#446](https://github.com/Kong/kuma/pull/446)
* feature: ground work for "provided" CA support
  [#467](https://github.com/Kong/kuma/pull/467)
* feature: remove "namespace" from core model
  [#458](https://github.com/Kong/kuma/pull/458)
  ⚠️ warning: api breaking change
* feature: expose effective configuration of `kuma-cp` as part of REST API
  [#454](https://github.com/Kong/kuma/pull/454)
* feature: improve error messages in `kumactl config control-planes add`
  [#455](https://github.com/Kong/kuma/pull/455)
* feature: delete resource operation should return 404 if resource is not found
  [#450](https://github.com/Kong/kuma/pull/450)
* feature: autoconfigure bootstrap server on `kuma-cp` startup
  [#449](https://github.com/Kong/kuma/pull/449)
* feature: update envoy to v1.12.1
  [#448](https://github.com/Kong/kuma/pull/448)

Breaking changes:
* ⚠️ a few arguments of `kumactl config control-planes add` have been renamed: `--dataplane-token-client-cert => --admin-client-cert` and `--dataplane-token-client-key => --admin-client-key`
  [474](https://github.com/Kong/kuma/pull/474)
* ⚠️ instead of applying all matching `TrafficLog` policies to a given `outbound` interface of a `Dataplane`, only a single the most specific `TrafficLog` policy is now applied
  [#482](https://github.com/Kong/kuma/pull/482)
* ⚠️ `Mesh` CRD on Kubernetes is now Cluster-scoped
  [#458](https://github.com/Kong/kuma/pull/458)

## [0.3.0]

> Released on 2019/11/18

Changes:

* fix: fixed discrepancy between `ProxyTemplate` documentation and actual implementation
  [#422](https://github.com/Kong/kuma/pull/422)
* chore: dropped support for `Mesh`-wide logging settings
  [#438](https://github.com/Kong/kuma/pull/438)
  ⚠️ warning: api breaking change
* feature: validate `ProxyTemplate` resource on CREATE/UPDATE in universal mode
  [#431](https://github.com/Kong/kuma/pull/431)
  ⚠️ warning: api breaking change
* feature: add `kumactl generate tls-certificate` command
  [#437](https://github.com/Kong/kuma/pull/437)
* feature: validate `TrafficLog` resource on CREATE/UPDATE in universal mode
  [#435](https://github.com/Kong/kuma/pull/435)
* feature: validate `TrafficPermission` resource on CREATE/UPDATE in universal mode
  [#436](https://github.com/Kong/kuma/pull/436)
* feature: dropped support for multiple rules per single `TrafficPermission` resource
  [#434](https://github.com/Kong/kuma/pull/434)
  ⚠️ warning: api breaking change
* feature: added configuration for Kuma UI
  [#428](https://github.com/Kong/kuma/pull/428)
* feature: included Kuma UI into `kuma-cp`
  [#410](https://github.com/Kong/kuma/pull/410)
* feature: dropped support for multiple rules per single `TrafficLog` resource
  [#433](https://github.com/Kong/kuma/pull/433)
  ⚠️ warning: api breaking change
* feature: validate `Mesh` resource on CREATE/UPDATE in universal mode
  [#430](https://github.com/Kong/kuma/pull/430)
* feature: `kumactl` commands now do custom formating of errors returned by the Kuma REST API
  [#411](https://github.com/Kong/kuma/pull/411)
* feature: `tcp_proxy` configuration now routes to a list of weighted clusters according to `TrafficRoute`
  [#423](https://github.com/Kong/kuma/pull/423)
* feature: included tags of a dataplane into `ClusterLoadAssignment`
  [#422](https://github.com/Kong/kuma/pull/422)
* feature: validate Kuma CRDs on Kubernetes
  [#401](https://github.com/Kong/kuma/pull/401)
* feature: improved feedback given to a user when `kuma-dp run` is configured with an invalid dataplane token
  [#418](https://github.com/Kong/kuma/pull/418)
* release: included Docker image with `kumactl` into release build
  [#425](https://github.com/Kong/kuma/pull/425)
* feature: support enabling/disabling DataplaneToken server via a configuration flag
  [#415](https://github.com/Kong/kuma/pull/415)
* feature: pick a single the most specific `TrafficRoute` for every outbound interface of a `Dataplane`
  [#421](https://github.com/Kong/kuma/pull/421)
* feature: validate `TrafficRoute` resource on CREATE/UPDATE in universal mode
  [#424](https://github.com/Kong/kuma/pull/424)
* feature: `kumactl apply` can now download a resource from URL
  [#402](https://github.com/Kong/kuma/pull/402)
* chore: migrated to the latest version of `go-control-plane`
  [#419](https://github.com/Kong/kuma/pull/419)
* feature: added `kumactl get traffic-routes` command
  [#400](https://github.com/Kong/kuma/pull/400)
* feature: added `TrafficRoute` CRD on Kubernetes
  [#398](https://github.com/Kong/kuma/pull/398)
* feature: added `TrafficRoute` resource to core model
  [#397](https://github.com/Kong/kuma/pull/397)
* feature: added support for CORS to Kuma REST API
  [#412](https://github.com/Kong/kuma/pull/412)
* feature: validate `Dataplane` resource on CREATE/UPDATE in universal mode
  [#388](https://github.com/Kong/kuma/pull/388)
* feature: added support for client certificate-based authentication to `kumactl generate dataplane-token` command
  [#372](https://github.com/Kong/kuma/pull/372)
* feature: added `--overwrite` flag to the `kumactl config control-planes add` command
  [#381](https://github.com/Kong/kuma/pull/381)
  👍contributed by @Gabitchov
* feature: added `MESH` column into the output of `kumactl get proxytemplates`
  [#399](https://github.com/Kong/kuma/pull/399)
  👍contributed by @programmer04
* feature: `kuma-dp run` is now configured with a URL of the API server instead of a former URL of the boostrap config server
  [#417](https://github.com/Kong/kuma/pull/417)
  ⚠️ warning: interface breaking change
* feature: added a REST endpoint to advertize location of various sub-components of the control plane
  [#369](https://github.com/Kong/kuma/pull/369)
* feature: added protobuf descriptor for `TrafficRoute` resource
  [#396](https://github.com/Kong/kuma/pull/396)
* fix: added reconciliation on Dataplane delete to handle a case where a user manually deletes Dataplane on Kubernetes
  [#392](https://github.com/Kong/kuma/pull/392)
* feature: Kuma REST API on Kubernetes is now restricted to READ operations only
  [#377](https://github.com/Kong/kuma/pull/377)
  👍contributed by @sterchelen
* fix: ignored errors in unit tests
  [#376](https://github.com/Kong/kuma/pull/376)
  👍contributed by @alrs
* feature: JSON output of `kumactl` is now pretty-printed
  [#360](https://github.com/Kong/kuma/pull/360)
  👍contributed by @sterchelen
* feature: DataplaneToken server is now exposed for remote access over HTTPS with mandatory client certificate-based authentication
  [#349](https://github.com/Kong/kuma/pull/349)
* feature: `kuma-dp` now passes a path to a file with a dataplane token as an argumenent for bootstrap config API
  [#348](https://github.com/Kong/kuma/pull/348)
* feature: added support for mTLS on Kubernetes v1.13+
  [#356](https://github.com/Kong/kuma/pull/356)
* feature: added `kumactl delete` command
  [#343](https://github.com/Kong/kuma/pull/343)
  👍contributed by @pradeepmurugesan
* feature: added `kumactl gerenerate dataplane-token` command
  [#342](https://github.com/Kong/kuma/pull/342)
* feature: added a DataplaneToken server to support dataplane authentication in universal mode
  [#342](https://github.com/Kong/kuma/pull/342)
* feature: on removal of a Mesh remove all policies defined in it
  [#332](https://github.com/Kong/kuma/pull/332)
* docs: documented release process
  [#341](https://github.com/Kong/kuma/pull/341)
* docs: DEVELOPER.md was brought up to date
  [#346](https://github.com/Kong/kuma/pull/346)
* docs: added instructions how to deploy `kuma-demo` on Kubernetes
  [#347](https://github.com/Kong/kuma/pull/347)

Community contributions from:

* 👍@pradeepmurugesan
* 👍@alrs
* 👍@sterchelen
* 👍@programmer04
* 👍@Gabitchov

Breaking changes:

* ⚠️ fixed discrepancy between `ProxyTemplate` documentation and actual implementation
  [#422](https://github.com/Kong/kuma/pull/422)
* ⚠️ `selectors` in `ProxyTemplate` now always require `service` tag
  [#431](https://github.com/Kong/kuma/pull/431)
* ⚠️ dropped support for `Mesh`-wide logging settings
  [#438](https://github.com/Kong/kuma/pull/438)
* ⚠️ dropped support for multiple rules per single `TrafficPermission` resource
  [#434](https://github.com/Kong/kuma/pull/434)
* ⚠️ dropped support for multiple rules per single `TrafficLog` resource
  [#433](https://github.com/Kong/kuma/pull/433)
* ⚠️ value of `--cp-address` parameter in `kuma-dp run` is now a URL of the API server instead of a former URL of the boostrap config server
  [#417](https://github.com/Kong/kuma/pull/417)

## [0.2.2]

> Released on 2019/10/11

Changes:

* Draining time is now configurable
  [#310](https://github.com/Kong/kuma/pull/310)
* Validation that Control Plane is running when adding it with `kumactl`
  [#181](https://github.com/Kong/kuma/issues/181)
* Upgraded version of go-control-plane
* Upgraded version of Envoy to 1.11.2
* Connection timeout to ADS server is now configurable (part of `envoy` bootstrap config)
  [#340](https://github.com/Kong/kuma/pull/340)

Fixed issues:
* Cluster never went out warming state
  [#331](https://github.com/Kong/kuma/pull/331)
* SDS server didn't handle requests with empty resources list
  [#337](https://github.com/Kong/kuma/pull/337) 

## [0.2.1]

> Released on 2019/10/03

Fixed issues:

* Issue with `Access Log Server` (integrated into `kuma-dp`) on k8s:
 `kuma-cp` was configuring Envoy to use a Unix socket other than
 `kuma-dp` was actually listening on
  [#307](https://github.com/Kong/kuma/pull/307)

## [0.2.0]

> Released on 2019/10/02

Changes:

* Fix an issue with `Access Log Server` (integrated into `kuma-dp`) on Kubernetes
  by replacing `Google gRPC client` with `Envoy gRPC client`
  [#306](https://github.com/Kong/kuma/pull/306)
* Settings of a `kuma-sidecar` container, such as `ReadinessProbe`, `LivenessProbe` and `Resources`,
  are now configurable
  [#304](https://github.com/Kong/kuma/pull/304)
* Added support for `TCP` logging backends, such as `ELK` and `Splunk`
  [#300](https://github.com/Kong/kuma/pull/300)
* `Builtin CA` on `Kubernetes` is now (re-)generated by a `Controller`
  [#299](https://github.com/Kong/kuma/pull/299)
* Default `Mesh` on `Kubernetes` is now (re-)generated by a `Controller`
  [#298](https://github.com/Kong/kuma/pull/298)
* Added `Kubernetes Admission WebHook` to apply defaults to `Mesh` resources
  [#297](https://github.com/Kong/kuma/pull/297)
* Upgraded version of `kubernetes-sigs/controller-runtime` dependency
  [#293](https://github.com/Kong/kuma/pull/293)
* Added a concept of `RuntimePlugin` to `kuma-cp`
  [#296](https://github.com/Kong/kuma/pull/296)
* Updated `LDS` to configure `access_loggers` on `outbound` listeners
  according to `TrafficLog` resources
  [#276](https://github.com/Kong/kuma/pull/276)
* Changed default locations where `kuma-dp` is looking for `envoy` binary
  [#268](https://github.com/Kong/kuma/pull/268)
* Added model for `TrafficLog` resource with `File` as a logging backend
  [#266](https://github.com/Kong/kuma/pull/266)
* Added `kumactl install database-schema` command to generate DB schema
  used by `kuma-cp` on `universal` environment
  [#236](https://github.com/Kong/kuma/pull/236)
* Automated release of `Docker` images
  [#265](https://github.com/Kong/kuma/pull/265)
* Changed default location where auto-generated Envoy bootstrap configuration is saved to
  [#261](https://github.com/Kong/kuma/pull/261)
* Added support for multiple `kuma-dp` instances on a single Linux machine
  [#260](https://github.com/Kong/kuma/pull/260)
* Automated release of `*.tar` artifacts
  [#250](https://github.com/Kong/kuma/pull/250)

Fixed issues (user feedback):

* Dataplanes cannot connect to a non-default Mesh with mTLS enabled on k8s
  [262](https://github.com/Kong/kuma/issues/262)
* Starting multiple services on the same Linux machine
  [254](https://github.com/Kong/kuma/issues/254)
* Fallback when invoking `envoy` from `kuma-dp`
  [249](https://github.com/Kong/kuma/issues/249)

## [0.1.2]

> Released on 2019/09/11

* Upgraded version of Go to address CVE-2019-14809.
  [#248](https://github.com/Kong/kuma/pull/248)
* Improved support for mTLS on `kubernetes`.
  [#238](https://github.com/Kong/kuma/pull/238)

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
