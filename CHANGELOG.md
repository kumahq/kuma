# CHANGELOG

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
