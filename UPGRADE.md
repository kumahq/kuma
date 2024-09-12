This document guides you through the process of upgrading `Kuma`.

First, check if a section named `Upgrade to x.y.z` exists,
with `x.y.z` being the version you are planning to upgrade to.

If such a section does not exist, the upgrade you want to perform
does not have any particular instructions.

## Upgrade to `2.9.x`

### Upgrading Transparent Proxy Configuration

#### Removal of Deprecated IPv6 Redirection Flag and Annotation

In this release, the following deprecated options for configuring IPv6 transparent proxy redirection have been removed:

- The `--redirect-inbound-port-ipv6` flag in `kumactl install transparent-proxy`.
- The `kuma.io/transparent-proxying-inbound-v6-port` annotation.

Previously, disabling IPv6 transparent proxy redirection could be achieved by setting these options to `0`. This method is no longer supported.

To disable IPv6 transparent proxy redirection, you should now use the `--ip-family-mode` flag or the `kuma.io/transparent-proxying-ip-family-mode` annotation and set their value to `ipv4`. The default value for these options is `dualstack`.

**Example:**

In Universal mode, to install a transparent proxy:

```sh
kumactl install transparent-proxy --ip-family-mode ipv4 ...
```

In the definition of the `Dataplane` resource:

```yaml
type: Dataplane
mesh: default
name: dp-1
networking:
  # ...
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortOutbound: 15001
    ipFamilyMode: ipv4
```

To set the configuration for Kubernetes workloads:

```sh
kumactl install control-plane --set controlPlane.envVars.KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_IP_FAMILY_MODE=ipv4 ...
```

or

```sh
helm install --set controlPlane.envVars.KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_IP_FAMILY_MODE=ipv4 ... kuma kuma/kuma
```

For more information about disabling IPv6 in transparent proxy redirection, visit our documentation: [Disabling IPv6](https://kuma.io/docs/2.8.x/production/dp-config/ipv6/#disabling-ipv6).

Please update your configurations accordingly to ensure a smooth transition and avoid any disruptions in your service.

#### Removal of `redirectPortInboundV6` Field from Dataplane Resource

The `Dataplane` resource no longer includes the `redirectPortInboundV6` field. Any configuration containing this field will fail validation. Update your `Dataplane` resources as shown below:

**Previous configuration:**

```yaml
type: Dataplane
mesh: default
name: dp-1
networking:
  # ...
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortInboundV6: 15006
    redirectPortOutbound: 15001
```

**Updated configuration:**

```yaml
type: Dataplane
mesh: default
name: dp-1
networking:
  # ...
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortOutbound: 15001
```

Ensure to update your Dataplane resources to the new format to avoid any validation errors.

#### Removal of Deprecated Exclude Outbound TCP/UDP Ports for UIDs Flags

The flags `--exclude-outbound-tcp-ports-for-uids` and `--exclude-outbound-udp-ports-for-uids` have been removed from the `kumactl install transparent-proxy` command. Users should now use the consolidated flag `--exclude-outbound-ports-for-uids <protocol:>?<ports:>?<uids>` instead.

##### Examples:

- To disable redirection of outbound TCP traffic on port 22 for users with UID 1000:
  ```sh
  kumactl install transparent-proxy --exclude-outbound-ports-for-uids tcp:22:1000 ...
  ```

- To disable redirection of outbound UDP traffic on port 53 for users with UID 1000:
  ```sh
  kumactl install transparent-proxy --exclude-outbound-ports-for-uids udp:53:1000 ...
  ```

#### Removal of Deprecated Exclude Outbound TCP/UDP Ports for UIDs Annotations

The annotations `traffic.kuma.io/exclude-outbound-tcp-ports-for-uids` and `traffic.kuma.io/exclude-outbound-udp-ports-for-uids` have also been removed. Use the annotation `traffic.kuma.io/exclude-outbound-ports-for-uids` instead.

##### Examples:

- To disable redirection of outbound TCP traffic on port 22 for users with UID 1000:
  ```yaml
  traffic.kuma.io/exclude-outbound-ports-for-uids: tcp:22:1000
  ```

- To disable redirection of outbound UDP traffic on port 53 for users with UID 1000:
  ```yaml
  traffic.kuma.io/exclude-outbound-ports-for-uids: udp:53:1000
  ```

Make sure to update your configuration files and scripts accordingly to accommodate these changes.

#### Deprecation of `--kuma-dp-uid` Flag

In this release, the `--kuma-dp-uid` flag used in the `kumactl install transparent-proxy` command has been deprecated. The functionality of specifying a user by UID is now included in the `--kuma-dp-user` flag, which accepts both usernames and UIDs.

**New Usage Example:**

Instead of using:
```sh
kumactl install transparent-proxy --kuma-dp-uid 1234
```

You should now use:
```sh
kumactl install transparent-proxy --kuma-dp-user 1234
```

If the `--kuma-dp-user` flag is not provided, the system will attempt to use the default UID (`5678`) or the default username (`kuma-dp`).

Please update your scripts and configurations accordingly to accommodate this change.

### Setting `kuma.io/service` in tags of `MeshGatewayInstance` had been forbidden

To increase security, in version 2.7.x, setting a `kuma.io/service` tag for the `MeshGatewayInstance` was deprecated and since 2.9.x is not supported. We generate the `kuma.io/service` tag based on the `MeshGatewayInstance` resource. The service name is constructed as `{MeshGatewayInstance name}_{MeshGatewayInstance namespace}_svc`.

E.g.:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: demo-app
  namespace: kuma-demo
  labels:
    kuma.io/mesh: default
```

The generated `kuma.io/service` value is `demo-app_kuma-demo_svc`.

#### Migration

The migration process requires updating all policies and `MeshGateway` resources using the old `kuma.io/service` value to adopt the new one.

Migration step:
1. Create a copy of policies using the new `kuma.io/service` and the new resource name to avoid overwriting previous policies.
2. Duplicate the `MeshGateway` resource with a selector using the new `kuma.io/service` value.
3. Deploy the gateway and verify if traffic works correctly.
4. Remove the old resources.

### Introduction to Application Probe Proxy and deprecation of Virtual Probes

To support more types of application probes on Kubernetes, in version 2.9, we introduced a new feature named "Application Probe Proxy" which supports HTTP Get, TCP Socket and gRPC application probes. Starting from `2.9.x`, Virtual Probes is deprecated, and Application Probe Proxy is enabled by default.

Application workloads using Virtual Probes will be migrated to Application Probe Proxy automatically on next restart/redeploy on Kubernetes, without other operations. 

Application Probe Proxy will by default listen on port `9001`. If you'd customized the Virtual Probes port, you might also want to customize the port of Application Probe Proxy. You may do so using one of these methods:

1. Configuring on the control plane to apply on all dataplanes: set the port onto configuration key `runtime.kubernetes.injector.sidecarContainer.applicationProbeProxyPort` 
1. Configuring on the control plane to apply on all dataplanes: set the port using environment variable `KUMA_RUNTIME_KUBERNETES_APPLICATION_PROBE_PROXY_PORT` 
1. Configuring for certain dataplanes: set the port using pod annotation `kuma.io/application-probe-proxy-port`

By setting the port to `0`, Application Probe Proxy feature will be disabled.

When the Application Probe Proxy is disabled, Virtual Probes still works as usual before Virtual Probes is removed.

Because of deprecation of Virtual Probes, the following items are considered deprecated:

- Pod annotation `kuma.io/virtual-probes`
- Pod annotation `kuma.io/virtual-probes-port`
- Control plane configuration key `runtime.kubernetes.injector.sidecarContainer.virtualProbesEnabled`
- Control plane configuration key `runtime.kubernetes.injector.sidecarContainer.virtualProbesPort`
- Control plane environment variable `KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_ENABLED`
- Control plane environment variable `KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_PORT`
- Data field `probes` on `Dataplane` objects

### kumactl

#### Default prometheus scrape config removes `service`

If you rely on a scrape config from previous version it's advised to remove the relabel config that was adding `service`.
Indeed `service` is a very common label and metrics were sometimes coliding with Kuma metrics. If you want the label `kuma_io_service` is always the same as `service`.

### Removal of KDS `KUMA_EXPERIMENTAL_KDS_DELTA_ENABLED` configuration option

In this release, KDS Delta is used by default and the CP environment variable `KUMA_EXPERIMENTAL_KDS_DELTA_ENABLED` doesn't exist anymore.

### Deprecation of `yes/no` values for annotation switches

The values `yes` and `no` are deprecated for specifying boolean values in switches based on pod annotations, and support for these values will be removed in a future release. Since these values were undocumented, they are not expected to be widely used.

Please use `true` and `false` as replacements; some boolean switches also support `enabled` and `disabled`. [Check the documentation](https://kuma.io/docs/latest/reference/kubernetes-annotations/) for the specific annotation to confirm the correct replacements.

## Upgrade to `2.8.x`

### MeshFaultInjection responseBandwidth.limit

With [#10371](https://github.com/kumahq/kuma/pull/10371) we have tightened the validation of the `responseBandwidth.limit` field in `MeshFaultInjection` policy. Policies with invalid values, such as `-10kbps`, will be rejected.

### MeshRetry tcp.MaxConnectAttempt

With [#10250](https://github.com/kumahq/kuma/pull/10250) `MeshRetry` policies with `spec.tcp.MaxConnectAttempt=0` will be rejected.
Prior to 2.8.x these were semantically valid but would create invalid Envoy configuration and would cause issues on the dataplane.
Now this is rejected sooner to avoid service disruption.

### Removal of legacy tokens

Tokens issued from versions before 2.1.x needs to renewed before upgrading.

If you observe following log in control-plane logs, please rotate your tokens before upgrade.
```yaml
[WARNING] Using token with KID header, you should rotate this token as it will not be valid in future versions of Kuma
```
* [User token](https://kuma.io/docs/2.7.x/production/secure-deployment/api-server-auth/)
* [Dataplane token](https://kuma.io/docs/2.7.x/production/secure-deployment/dp-auth/)
* [Zone token](https://kuma.io/docs/2.7.x/production/cp-deployment/zoneproxy-auth/#zone-token)

## Upgrade to `2.7.x`

### MeshMetric and cluster stats merging

For MeshMetric we disabled cluster [stats merging](https://github.com/kumahq/kuma/pull/9768) so that metrics are generated per [traffic split](https://kuma.io/docs/2.6.x/policies/meshhttproute/#traffic-split).
This means that in Grafana there will be at least two entries under "Destination service" - one for the service without a hash (e.g. `backend_kuma-demo_svc_3001`) and one per each split ending with a hash (e.g. `backend_kuma-demo_svc_3001-de1397ec09e96dfb`).
If you want to see combined metrics you can run queries with a prefix instead of exact match, e.g.:

```
... envoy_cluster_name=~"$destination_cluster.*" ...
```

instead of

```
... envoy_cluster_name="$destination_cluster" ...
```

To correlate between a hash and a particular pod you have to click on the outbound, and then click on "clusters" and associate pod ip with cluster ip.
This will be improved in the future by having the tags next to the outbound.
[This issue](https://github.com/kumahq/kuma-gui/issues/2412) tracks the progress of that as well as contains screenshots of the steps.

### MeshMetric `sidecar.regex` is replaced by `sidecar.profiles.exclude`

If you're using `sidecar.regex` field it is getting replaced by `sidecar.profiles.exclude`.
Replace usages of:

```yaml
...
  sidecar:
    regex: "my_match.*"
...
```

with:

```yaml
  sidecar:
    profiles:
      exclude:
        - type: Regex
          match: "my_match.*"
```

### Setting `kuma.io/service` in tags of `MeshGatewayInstance` is deprecated

To increase security, since version 2.7.x, setting a `kuma.io/service` tag for the `MeshGatewayInstance` is deprecated. If the tag is not provided, we generate the `kuma.io/service` tag based on the `MeshGatewayInstance` resource. The service name is constructed as `{MeshGatewayInstance name}_{MeshGatewayInstance namespace}_svc`.

E.g.:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayInstance
metadata:
  name: demo-app
  namespace: kuma-demo
  labels:
    kuma.io/mesh: default
```

The generated `kuma.io/service` value is `demo-app_kuma-demo_svc`.

#### Migration

The migration process requires updating all policies and `MeshGateway` resources using the old `kuma.io/service` value to adopt the new one.

Migration step:
1. Create a copy of policies using the new `kuma.io/service` and the new resource name to avoid overwriting previous policies.
2. Duplicate the `MeshGateway` resource with a selector using the new `kuma.io/service` value.
3. Deploy the gateway and verify if traffic works correctly.
4. Remove the old resources.

### ZoneIngress Token support removed

The control-plane does not support tokens generated with `kumactl generate zone-ingress-token`. If you are running Kuma ingress with a zone ingress token generated using the deprecated method, before upgrading, verify if you are still using the old token.

#### How to validate if I am using `zone-ingress-token`?

1. Obtain the ingress token value
2. Run the following command
```bash
jq -R 'split(".") | .[0],.[1] | @base64d | fromjson' <<< $YOUR_TOKEN
```

Example output of a zone token:
```json
{
  "alg": "RS256",
  "kid": "1",
  "typ": "JWT"
}
{
  "Zone": "test",
  "Scope": [
    "ingress",
    "egress",
    "cp",
    "ratelimit"
  ],
  "exp": 1712414035,
  "nbf": 1709821735,
  "iat": 1709822035,
  "jti": "efeb8cca-2341-47a4-b4f2-daf49290e481"
}
```

Example output of a zone ingress token:
```json
{
  "alg": "RS256",
  "kid": "1",
  "typ": "JWT"
}
{
  "Zone": "test",
  "exp": 1709822002,
  "nbf": 1709821702,
  "iat": 1709822002,
  "jti": "c4cf30c5-ca30-42ec-b08d-de56fba75e7b"
}
```
3. If the output does not have the `Scope` field, you need to generate a new zone token using `kumactl generate zone-token` for your ingress before upgrading.
4. Restart the Ingress with the new token.
5. Now, you can safely upgrade the control-plane.

### Configuration option `KUMA_DP_SERVER_AUTH_*`, `dpServer.auth.*` was removed

The option to configure authentication was deprecated and has been removed in release `2.7.x`. If you are still using `KUMA_DP_SERVER_AUTH_*`
environment variables or `dpServer.auth.*` configuration, please migrate your configuration to use `dpServer.authn` before upgrade.

### Deprecation of `--redirect-inbound-port-v6` flag and `runtime.kubernetes.injector.sidecarContainer.redirectPortInboundV6` configuration option.

The `--redirect-inbound-port-v6` flag and the corresponding configuration option `runtime.kubernetes.injector.sidecarContainer.redirectPortInboundV6` are deprecated and will be removed in a future release of Kuma. These flags and configuration options were used to configure the port used for redirecting IPv6 traffic to Kuma.

In the upcoming release, Kuma will redirect IPv6 traffic to the same port as IPv4 traffic (15006). This means that you no longer need to configure a separate port for IPv6 traffic. If you want to disable traffic redirection for IPv6 traffic, you can set `--ip-family-mode ipv4`. We have also added a new configuration option `runtime.kubernetes.injector.sidecarContainer.ipFamilyMode` to switch traffic redirection for IP families.

We recommend that you update your configurations to use the new defaults for IPv6 traffic redirection. If you need to retain separate ports for IPv4 and IPv6 traffic, you can continue to use the deprecated flags and configuration options until they are removed.

### Deprecation of 'from[].targetRef.kind: MeshService'

At this moment only MeshTrafficPermission and MeshFaultInjection allowed `MeshService` in the `from[].targetRef.kind`.
Starting `2.7` this value is deprecated, instead the `MeshSubset` with `kuma.io/service` tag should be used. For example, instead of:

```yaml
type: MeshTrafficPermission
name: allow-orders
mesh: default
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: MeshService
        name: orders
      default:
        action: Allow
```

we should have:

```yaml
type: MeshTrafficPermission
name: allow-orders
mesh: default
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: MeshSubset
        tags:
          kuma.io/service: orders
      default:
        action: Allow
```

### Change in internal resources with Kubernetes Gateway API

This section describes changes to internal resources used by Kuma when configuring the built-in gateway using the Kubernetes Gateway API.

#### Prior Behavior (Before Kuma 2.7.0):

  * Applying a `Gateway` resource resulted in the creation of corresponding `MeshGateway` and `MeshGatewayInstance` resources.
  * An applied `HTTPRoute` resource was converted to a `MeshGatewayRoute` resource.

#### Changes Introduced in Kuma 2.7.0:

  * `HTTPRoute` resources are now converted to `MeshHTTPRoute` resources instead of `MeshGatewayRoute` resources.

#### Upgrade Impact:

  * Existing `MeshGatewayRoute` resources automatically created from `HTTPRoute` definitions will be deleted during the upgrade.
  * New `MeshHTTPRoute` resources will be created to replace the deleted ones.

#### Important Note:

This change is transparent with regard to the generated Envoy configuration. There should be no impact on existing traffic routing.

### Gateway API Promotion to GA

The Gateway API functionality within Kuma is now considered Generally Available (GA). This means the `--experimental-gatewayapi` flag and the `experimental.gatewayAPI` setting are no longer required for installation.

> [!WARNING]
> If you previously used the `--experimental-gatewayapi` flag with `kumactl install control-plane` in your workflows, it's important to note that this flag has been removed and is no longer supported. Using it will now result in an error.

#### Removed Flags and Settings

Previously, these flags were necessary for using the Gateway API feature:

- `--experimental-gatewayapi` flag for `kumactl install control-plane` and `kumactl install crds`
- `experimental.gatewayAPI=true` setting in both `kumactl install control-plane` and Helm charts

### TLS Secrets with Gateway API in namespace other than mesh system namespace

If you use TLS secrets with Gateway API for a builtin gateway deployed in any other namespace than mesh system namespace, set `controlPlane.supportGatewaySecretsInAllNamespaces` HELM value to true.
This change was introduced so that control plane does not have capability to read content of secrets in all namespaces by default.

## Upgrade to `2.6.x`

### Policy

#### Sorting

This change relates only to the new targetRef policies. When 2 policies have a tie on the targetRef kind we compare their names lexicographically.
Policy merging now gives precedence to policies that lexicographically "less" than other policies, i.e. policy "aaa" takes precedence over "bbb" because "aaa" < "bbb".
Previously, before 2.6.0 the order was the opposite.

#### `targetRef.kind: MeshGateway`

Note that when targeting `MeshGateways` you should be using `targetRef.kind:
MeshGateway`. Previously `targetRef.kind: MeshService` was necessary but this
left the control plane unable to fully validate policies for builtin gateway
usage.

##### `to` instead of `from`

With `MeshFaultInjection` and `MeshRateLimit`, `spec.to` with `kind:
MeshGateway` is now required instead of `spec.from` and `kind: MeshService`.

### `MeshGateway`

A new maximum length of 253 characters for listener hostnames has been introduced in order to ensure they are valid DNS names.

### Unifying Default Connection Timeout Values

To simplify configuration and provide a more consistent user experience, we've unified the default connection timeout values. When no `MeshTimeout` or `Timeout` policy is specified, the connection timeout will now be the same as the default `connectTimeout` values for `MeshTimeout` and `Timeout` policies. This value is now `5s`, which is a decrease from the previous default of `10s`.

The connection timeout specifies the amount of time Envoy will wait for an upstream TCP connection to be established.

The only users who need to take action are those who are explicitly relying on the previous default connection timeout value of `10s`. These users will need to create a new `MeshTimeout` policy with the appropriate `connectTimeout` value to maintain their desired behavior.

We encourage all users to review their configuration, but we do not anticipate that this change will require any action for most users.

### Default `TrafficRoute` and `TrafficPermission` resources are not created when creating a new `Mesh`

We decided to remove default `TrafficRoute` and `TrafficPermission` policies that were created during a new mesh creation. Since this release your applications can communicate without need to apply any policy by default.
If you want to keep the previous behaviour set `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES` to `true`.

**The following policies will no longer be created automatically**:
  
  * `CircuitBreaker`
  * `Retry`
  * `Timeout`
  * `TrafficPermission`
  * `TrafficRoute`

**The following policies will be created by default**:

  * `MeshCircuitBreaker`
  * `MeshRetry`
  * `MeshTimeout`

> [!CAUTION]
> Before enabling `mTLS`, remember to add `MeshTrafficPermission.`

Previously, Kuma would automatically create the default `TrafficPermission` policy for traffic routing. However, starting from version `2.6.0`, this is no longer the case.

If you are using `mTLS`, you will need to manually create the `MeshTrafficPermission` policy before enabling `mTLS`.

The `MeshTrafficPermission` policy allows you to specify which services can communicate with each other. This is necessary in a `mTLS` environment because `mTLS` requires that all communication between services be authenticated and authorized.

#### When is it appropriate to set the `KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES` environment variable to `true`?

* When zones connecting to the global control plane may be running an older version than `2.6.0`.
* When recreating an environment using continuous delivery (CD) with legacy policies, missing the `TrafficRoute` policy will prevent legacy policies from being applied.

### Change of underlying envoy RBAC plugin for MeshTrafficPermission policies targeting HTTP services

With the release of Kuma 2.6.0, we've made some changes to the implementation of `MeshTrafficPermission` policies targeting HTTP services. These changes primarily revolve around the use of the `envoy.filters.http.rbac` envoy filter instead of the `envoy.filters.network.rbac` filter. This migration entails the following adjustments:

1. **Denied Request Response**: Rejected requests will now receive a 403 response code with the message `RBAC: access denied` instead of the previous 503 code. This aligns with the typical HTTP response code for authorization failures.

2. **RBAC-Related Envoy Stats**: The prefix for RBAC-related Envoy stats has been updated from `<inbound|outbound>:<stat_prefix>.rbac.` to `http.<stat_prefix>.rbac.`. This reflects the use of the HTTP filter for RBAC enforcement. For instance, the stat `inbound:127.0.0.1:21011.rbac.allowed` will now become `http.127.0.0.1:21011.rbac.allowed.` If you're utilizing these stats in your observability stack, you'll need to update your configuration to reflect the change.

To ensure a smooth transition to Kuma 2.6.0, carefully review your existing configuration files and make necessary adjustments related to denied request responses and RBAC-related Envoy stats.

### Make SI format valid for bandwidth in MeshFaultInjection policy

Prior to this upgrade `mbps` and `gbps` were used for units for parameter `conf.responseBandwidth.percentage`.
These are not valid units according to the [International System of Units](https://en.wikipedia.org/wiki/International_System_of_Units) they are respectively corrected to `Gbps` and `Mbps` if using
these invalid units convert them into `kbps` prior to upgrade to avoid invalid format.

### Deprecation of postgres driverName=postgres (lib/pq)

The postgres driver `postgres` (lib/pq) is deprecated and will be removed in the future.
Please migrate to the new postgres driver `pgx` by setting `DriverName=pgx` configuration option or `KUMA_STORE_POSTGRES_DRIVER_NAME=pgx` env variable.

## Upgrade to `2.5.x`

### Transparent-proxy and CNI v1 removal

v2 has been default since 2.2.x. We are therefore removing v1.

### Deprecated argument to transparent-proxy

Parameters `--exclude-outbound-tcp-ports-for-uids` and `--exclude-outbound-udp-ports-for-uids` are now merged into `--exclude-outbound-ports-for-uids` for `kumactl install transparent-proxy`.
We've also added the matching Kubernetes annotation: `traffic.kuma.io/exclude-outbound-ports-for-uids`.
The previous versions will still work but will be removed in the future.

### More strict validation rules for resource names

In order to be compatible with Kubernetes naming policy we updated the validation rules. Old rule:

> Valid characters are numbers, lowercase latin letters and '-', '_' symbols.

New rule:

> A lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character

New rule is applied for CREATE operations. The old rule is still applied for UPDATE, but this is going to change in Kuma 2.7.x or later.

### API

#### overview API coherency

These endpoints are getting replaced to achieve more coherency on the API:

- `/meshes/{mesh}/zoneegressoverviews` moves to `/meshes/{mesh}/zoneegresses/_overview`
- `/meshes/{mesh}/zoneingresses+insights` moves to `/meshes/{mesh}/zone-ingresses/_overview`
- `/meshes/{mesh}/dataplanes+insights` moves to `/meshes/{mesh}/dataplanes/_overview`
- `/zones+insights` moves to `/zones/_overview`

While you can use the old API they will be removed in a future version

### Prometheus inbound listener is not secured by TrafficPermission anymore

Due to the shadowing [issue](https://github.com/kumahq/kuma/issues/2417) with old TrafficPermission it was quite impossible to protect Prometheus inbound listener as expected.
RBAC rules on the Prometheus inbound listener were blocking users from fully migrate to the new MeshTrafficPermission policy. 
That's why we decided to discontinue TrafficPermission support on the Prometheus inbound listener starting 2.5.x.

### Gateway API

We support `v1` resources and `v1.0.0` of `gateway-api`. `v1beta1` resources are
still supported but support for these WILL be removed in a future release.

### KDS Delta enabled by default

KDS Delta is enabled by default. You can fallback to SOTW KDS by setting `KUMA_EXPERIMENTAL_KDS_DELTA_ENABLED=false`.
As a side effect, on kubernetes policies synced will be persisted in the `kuma-system` namespace instead of `default`.

## Upgrade to `2.4.x`

### Configuration change

The configuration: `Metrics.Mesh.MinResyncTimeout` and `Metrics.Mesh.MaxResyncTimeout` are replaced by `Metrics.Mesh.MinResyncInterval` and `Metrics.Mesh.FullResyncInterval`.
You can still use the current configs but it will be removed in the future.

### **Breaking changes**

#### Removal of service field in Dataplane outbound

After a period of depreciation, the service field in now removed. The service name is only defined by the value of  `kuma.io/service` in the outbound tags field.

## Upgrade to `2.3.x`

### **Breaking changes**

#### `MeshHTTPRoute`

* Changed path match `type` from `Prefix` to `PathPrefix`

#### `MeshAccessLog`

* Added a new field `Type` for `Backend` as a [Discriminator Field](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/1027-api-unions/README.md#discriminator-field)
* Added a new field `Type` for `Format` as a [Discriminator Field](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/1027-api-unions/README.md#discriminator-field)

#### `MeshTrace`

* Added a new field `Type` for `Backend` as a [Discriminator Field](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/1027-api-unions/README.md#discriminator-field)

#### `kumactl` container image

* Changed image's entrypoint to `/usr/bin/kumactl`

This change was introduced to be consistent with `kuma-cp` and `kuma-dp` images,
where names of images refer to binaries set in entrypoint. 

Example valid before:
```sh
docker run kumahq/kumactl:2.2.1 kumactl install transparent-proxy --help
```

Equivalent example valid now:
```sh
docker run kumahq/kumactl:2.3.0 install transparent-proxy --help
```

#### TLS verification between Zone CP and Global CP

If the CA used to sign the Global CP sync server is not provided to a Zone CP (HELM `controlPlane.tls.kdsZoneClient`, ENV: `KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE`), and the certificate is signed by a CA that is not included in the system's CA bundle on the Zone CP machine, you must do one of the following:
* Provide the CA to the Zone CP, see https://kuma.io/docs/2.2.x/production/secure-deployment/certificates/#control-plane-to-control-plane-multizone .
* Configure Zone CP. Set `KUMA_MULTIZONE_ZONE_KDS_TLS_SKIP_VERIFY` or HELM value of `controlPlane.tls.kdsZoneClient.skipVerify` to `true`. 

#### Removal of Common Name from generated certificates

This only affects users who rely on generated certificates having a common name set.

* `kumactl generate tls-certificate` generates certificates without CN
* autogenerated TLS certificate for kuma-cp (when `general.tlsCertFile` is not provided) won't have CN

## Upgrade to `2.2.x`

### Universal

#### CentOS 7

We are dropping support for running Envoy on CentOS 7 with this release and will
not release CentOS 7 compatible Envoy builds.

#### Changed default postgres driver to pgx

- If you encounter any problems with the persistence layer please [submit an issue](https://github.com/kumahq/kuma/issues/new) and temporarily switch to the previous driver (`lib/pq`) by setting
`DriverName=postgres` configuration option or `KUMA_STORE_POSTGRES_DRIVER_NAME='postgres'` env variable.
- Several configuration settings are not supported by the new driver right now, if used to configure them please try running with new defaults or [submit an issue](https://github.com/kumahq/kuma/issues/new).
List of unsupported configuration options:
  - MaxIdleConnections (used in store)
  - MinReconnectInterval (used in events listener)
  - MaxReconnectInterval (used in events listener)

#### Longer name of the resource in postgres

Kuma now permits the creation of a resource with a name of up to 253 characters, which is an increase from the previous limit of 100 characters. This adjustment brings our system in line with the naming convention supported by Kubernetes.
This change requires to run `kuma-cp migrate up` to apply changes to the postgres database.

### K8s

#### Removed deprecated annotations

- `kuma.io/builtindns` and `kuma.io/builtindnsport` are removed in favour of `kuma.io/builtin-dns` and `kuma.io/builtin-dns-port` introduced in 1.8.0. If you are using the legacy CNI you main need to set these old annotations manually in your pod definition.
- `kuma.io/sidecar-injection` is no longer supported as an annotation, you should use it as a label.

#### Helm

All containers now have defaults for `resources.requests.{cpu,memory}` and `resources.limits.{memory}`.
There are new default values for `*.podSecurityContext` and `*.containerSecurityContext`, see `values.yaml`.

#### Gateway API

We now support version `v0.6.0` of the Gateway API. See the [upstream API
changes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.6.0) for
more info.

### Auth configuration of DP server in Kuma CP

`dpServer.auth` configuration of Kuma CP was deprecated. You can still set config in this section, but it will be removed in the future.
It's recommended to migrate to `dpServer.authn` if you explicitly set any of the configuration in this config section.
* `dpServer.auth.type` is now split into two: `dpServer.authn.dpProxy.type` and `dpServer.authn.zoneProxy.type` and is still autoconfigured based on the environment.
* `dpServer.auth.useTokenPath` is now `dpServer.authn.enableReloadableTokens`

### Transparent Proxy Engine v2 and CNI v2 as default

As they matured, in the upcoming release Kuma will by default use transparent
proxy engine v2 and CNI v2.

If you want to still use v1 versions of these components, you will have to install 
Kuma with provided `legacy.transparentProxy=true` or `legacy.cni.enabled=true`
options.

#### Examples

##### CNI

*Helm*

```sh
helm upgrade --install --create-namespace --namespace kuma-system \
  --set "legacy.cni.enabled=true" \
  --set "cni.enabled=true" \
  --set "cni.chained=true" \
  --set "cni.netDir=/etc/cni/net.d" \
  --set "cni.binDir=/opt/cni/bin" \
  --set "cni.confName=10-calico.conflist"
  kuma kuma/kuma
```

*kumactl*

```sh
kumactl install control-plane \
  --set "legacy.cni.enabled=true" \
  --set "cni.enabled=true" \
  --set "cni.chained=true" \
  --set "cni.netDir=/etc/cni/net.d" \
  --set "cni.binDir=/opt/cni/bin" \
  --set "cni.confName=10-calico.conflist" \
  | kubectl apply -f-
```

##### Transparent Proxy Engine

*Helm*

```sh
helm upgrade --install --create-namespace --namespace kuma-system \
  --set "legacy.transparentProxy=true" kuma kuma/kuma
```

*kumactl*

```sh
kumactl install control-plane --set "legacy.transparentProxy=true" | kubectl apply -f-
```

### Removal of deprecated options to reach applications bound to `localhost`

The deprecated options `KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS` and
`defaults.enableLocalhostInboundClusters` were removed.

This change affects only applications using transparent proxy.

Applications that are binding to `localhost` won't be reachable anymore.
This is the default behaviour from Kuma 1.8.0. Until now, it was possible to set
a deprecated kuma-cp configurations `KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS`
or `defaults.enableLocalhostInboundClusters` to `true`, which was allowing to
still reach these applications.

One of the options to upgrade change address which the application is
listening on, to `0.0.0.0`.
Other option is to define `dataplane.networking.inbound[].serviceAddress`
to the address which service is binding to.

## Upgrade to `2.1.x`

### **Breaking changes**

#### **Naming Serviceless dataplanes has changed**

Currently, the `kuma.io/service` value of the inbound of a `Dataplane` generated for a `Pod` without a `Service` is based on the `Pod` name. The Kuma CP takes the pod's name and removes 2 last elements after splitting by `-`. This behavior is correct when the `Pod` is owned by a `Deployment` or `CronJob` but not for other owner kinds. Kuma will now use the name of the owner resource as the `kuma.io/service` value.
Before upgrade:
1. Identify all `Service`less `Pods` that are not managed by a `Deployment` or `CronJob`.
2. Create copies of policies that were created for the services corresponding to these `Pods`. The `kuma.io/service` value is the name of the owner resource. If there is no owner, `Kuma` uses the `Pod`'s name.

This breaking change is required to provide correct naming. The previous behavior could produce the same `kuma.io/service` value of the inbound of a `Dataplane` for many different serviceless Dataplanes.

#### MeshTrafficPermission

Action value have switched to PascalCase. ALLOW is Allow, DENY is Deny and ALLOW_WITH_SHADOW_DENY is AllowWithShadowDeny.

### HTTP api

We've removed the deprecated endpoint `POST /tokens`, use the `POST /tokens/dataplane` endpoint instead (same request and response).
Make sure you are using a recent `kumactl` or that you use the right path if using the API directly to upgrade with no issues.

### Kubernetes

The sidecar container is always injected first (since [#5436](https://github.com/kumahq/kuma/pull/5436)). This should only impact you when modifying the sidecar container with a container-patch. If you do so, upgrade Kuma and then change your container patch to modify the right container.

This version changes the leader election mechanism from leader for life to the more robust leader with lease.
As the result, during the upgrade you may have two leaders in the cluster.
This should not impact the system in any significant way other than logs like `resource was already updated`.

### Kumactl

`--valid-for` must be set for all token types, before it was defaulting to 10 years.

## Upgrade to `2.0.x`

### Built-in gateway

If you're using the `PREFIX` path match for `MeshGatewayRoute`,
note that validation is now stricter.
If you try to update an existing `MeshGatewayRoute` or create a new one,
make sure your `PREFIX` matching `value` does not include a trailing slash.
All prefix matches are checked path-separated,
meaning that `/prefix` only matches
if the request's path is `/prefix` or begins with `/prefix/`.
This has always been the case,
so no behavior has been changed
and existing resources with a trailing slash are not affected.

### Universal

A `lib/pq` change enables SNI by default when connecting to Postgres over TLS.
Either make sure your certificates contain a valid CN or SANs for the hostname
you're using
or update to `2.0.1` and disable `sslsni` by setting the
`KUMA_STORE_POSTGRES_TLS_DISABLE_SSLSNI` environment variable or
`store.postgres.tls.disableSSLSNI` in the config to `true`.

### `kuma-prometheus-sd`

This component has been removed
after [a long period of deprecation](https://github.com/kumahq/kuma/issues/2851).

### Zone Ingress Token migration

This is only relevant to Multizone deployment with Universal zones.
Zone Token that was previously used for authenticating Zone Egress, can now be used to authenticate Zone Ingress.
Please regenerate Zone Ingress token using `kumactl generate zone-token --scope=ingress`.
For the time being you can still use the old Zone Ingress token and Zone Token with scope ingress.
However, Zone Ingress Token is now deprecated and will be removed in the future.

### Helm

`ingress.annotations` and `egress.annotations` are deprecated in favour of `ingress.podAnnotations` and `egress.podAnnotations` which is a better name and aligne with the existing `controlPlane.podAnnoations`.


### Kuma-cp

- By default, the minimum TLS version allowed on servers is TLSv1.2. If you require using TLS < 1.2 you can set `KUMA_GENERAL_TLS_MIN_VERSION`.
- `KUMA_MONITORING_ASSIGNMENT_SERVER_GRPC_PORT` was removed after a long deprecation period use `KUMA_MONITORING_ASSIGNMENT_SERVER_PORT` instead.

### gRPC metrics

With this release, emitting separate statistics for every gRPC method is disabled.
gRPC metrics from different methods are now aggregated under `envoy_cluster_grpc_request_message_count`.
It will be re-enabled again in the future once Envoy with [`replace_dots_in_grpc_service_name`](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/grpc_stats/v3/config.proto#envoy-v3-api-field-extensions-filters-http-grpc-stats-v3-filterconfig-stats-for-all-methods) feature is released.
If you need to enable this setting, you can use ProxyTemplate to patch `envoy.filters.http.grpc_stats` http filter.

## Upgrade to `1.8.x`

### Kumactl

* `kumactl inspect dataplane --config-dump` was deprecated in favour of `kumactl inspect dataplane --type config-dump`. The behaviour of the new flag is unchanged but you should migrate.
* `kumactl install transparent-proxy --skip-resolv-conf` was deprecated as there's no reason for us to update the `/etc/resolv.conf` of the user.
* `kumactl install transparent-proxy --kuma-cp-ip` was removed as it's not possible to run a DNS server on the cp. 

### Helm

* Under `cni.image`, the default values for `repository` and `registry` have been
changed to agree with the other `image` values.

### CP

* The `/versions` endpoint was removed. This is not something that was reliable enough and version compatibility
is checked inside the DP
* We are deprecating `kuma.io/builtindns` and `kuma.io/builtindnsport` annotations in favour of the clearer `kuma.io/builtin-dns` and `kuma.io/builtin-dns-port`. The behavior of the new annotations is unchanged but you should migrate (a warning is present on the log if you are using the deprecated version).
* By default, applications binding to `localhost` are not reachable anymore. A `Dataplane` inbound's default `serviceAddress` is now the inbound's `address`. Before upgrade, if you have applications listening on `localhost` that you want to expose on:
  * Kubernetes: listen on `0.0.0.0` instead
  * Universal: listen on `inbound.address` instead or set `dataplane.networking.inbound[].serviceAddress: "127.0.0.1"`
To make migration easier you can temporarily disable this new behavior by setting `KUMA_DEFAULTS_ENABLE_LOCALHOST_INBOUND_CLUSTERS=true` on `kuma-cp`, this option will be removed in a future version.

## Upgrade to `1.7.x`

### Kumactl

* We're deprecating `kumactl install metrics/tracing/logging`, please use `kumactl install observability` instead

### DNS

The `control-plane` no longer hosts a builtin DNS server. You should always rely on the embedded DNS in the dataplane proxy and VIPs can't be used without transparent proxy.

### Timeout policy

'grpc' section is deprecated.
Timeouts for HTTP, HTTP2 and GRPC should be set in 'http' section:

```yaml
tcp: 
  idleTimeout: 1h 
http: # http, http2, grpc
  requestTimeout: 15s 
  idleTimeout: 1h
  streamIdleTimeout: 30m
  maxStreamDuration: 0s
grpc: # DEPRECATED
  streamIdleTimeout: 30m # DEPRECATED, use 'http.streamIdleTimeout'
  maxStreamDuration: 0s # DEPRECATED, use 'http.maxStreamDuration'
```

## Upgrade to `1.6.x`

### Helm

* the Helm chart for this release requires at least Helm version `3.8.0`.
* `controlPlane.resources` is now on object instead of a string. Any existing value should be adapted accordingly.

### Zone egress and ExternalService

When an `ExternalService` has the tag `kuma.io/zone` and `ZoneEgress` is enabled then the request flow will be different after upgrading Kuma to the newest version.
Previously, the request to the `ExternalService` goes through the `ZoneEgress` in the current zone. The newest version flow is different, and when `ExternalService` is defined in a different zone then the request will go through local `ZoneEgress` to `ZoneIngress` in zone where `ExternalService` is defined and leave the cluster through `ZoneEgress` in this cluster. To keep previous behavior, remove the `kuma.io/zone` tag from the `ExternalService` definition.

### Zone egress

Previously, when mTLS was configured and `ZoneEgress` deployed, requests were routed automatically through `ZoneEgress`. Now it's required to
explicitly set that traffic should be routed through `ZoneEgress` by setting `Mesh` configuration property `routing.zoneEgress: true`. The
default value of the property is set to `false` so in case your network policies don't allow you to reach other external services/zone without
using `ZoneEgress`, set `routing.zoneEgress: true`.

```yaml
type: Mesh
name: default
mtls: # mTLS is required for zoneEgress
 [...]
routing:
 zoneEgress: true
```

The new approach changes the flow of requests to external services. Previously when there was no instance of `ZoneEgress` traffic was routed
directly to the destination, now it won't reach the destination.

### Gateway (experimental)

Previously, a `MeshGatewayInstance` generated a `Deployment` and `Service` whose
names ended with a unique suffix. With this release, those objects will have the
same name as the `MeshGatewayInstance`.

### Inspect API

In connection with the changes around `MeshGateway` and `MeshGatewayRoute`, the output
schema of the `<policy-type>/<policy>/dataplanes` has changed. Every policy can
now affect both normal `Dataplane`s and `Dataplane`s configured as builtin gateways.
The configuration for the latter type is done via `MeshGateway` resources.

Every item in the `items` array now has a `kind` property of either:

* `SidecarDataplane`: a normal `Dataplane` with outbounds, inbounds,
  etc.
* `MeshGatewayDataplane`: a `MeshGateway`-configured `Dataplane` with a new
  structure representing the `MeshGateway` it serves.

Some examples can be found in the [Inspect API
docs](https://kuma.io/docs/1.6.x/reference/http-api/#inspect-api).

## Upgrade to `1.5.x`

### Any type

The `kuma.metrics.dataplane.enabled` and `kuma.metrics.zone.enabled` configurations have been removed.

Kuma always generate the corresponding metrics.

### Kubernetes

- Please migrate your `kuma.io/sidecar-injection` annotations to labels.
  The new version still supports annotation, but to have a guarantee that applications can only start with sidecar, you must use label instead of annotation.
- Configuration parameter `kuma.runtime.kubernetes.injector.sidecarContainer.adminPort` and environment variable `KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT`
  have been deprecated in favor of `kuma.bootstrapServer.params.adminPort` and `KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT`.

### Universal

- We removed support for old Ingress (`Dataplane#networking.ingress`) from pre 1.2 days.
  If you are still using it, please migrate to `ZoneIngress` first (see `Upgrade to 1.2.0` section).
- You can't use 0.0.0.0 or :: in `networking.address` most of the time using loopback is what people intended.
- Kuma DP flag `--admin-port` and environment variable `KUMA_DATAPLANE_ADMIN_PORT` have been deprecated, 
  admin port should be specified in Dataplane or ZoneIngress resources.

## Upgrade to `1.4.0`

Starting with this version, the default API server authentication method is user
tokens. In order to continue using client certificates (the previous default
method), you'll need to explicitly set the authentication method to client
certificates. This can be done by setting the `KUMA_API_SERVER_AUTHN_TYPE` variable to
`"clientCerts"`.

See [Configuration - Control plane](https://kuma.io/docs/1.3.1/documentation/configuration/#control-plane)
for how to set this variable.

## Upgrade to `1.3.0`

Starting with this version `Mesh` resource will limit the maximal number of mtls backends to 1, so please make sure your `Mesh` has correct backend applied before the upgrade.

Outbound generated internally are no longer listed in `dataplane.network.outbound[]`. For Kubernetes, they will automatically disappear. For universal to remove them you should recreate your dataplane resources (either with `kumactl apply` or by restarting your services if the dataplanes lifecycle is managed by Kuma).

Kuma 1.3.0 has additional mechanism for tracking data plane proxies and zone statuses in a more reliable way. This mechanism works as a heartbeat and periodically increments the `generation` counter for the Insights. If the overall time for upgrading all Kuma CP instances is more than 5 minutes, then some data plane proxies or zones may become Offline in the GUI, but this doesn't affect real connectivity, only view. This unwanted effect will disappear as soon as all Kuma CP instances will be upgraded to 1.3.0.

## Upgrade to `1.2.1`

When Global is upgraded to `1.2.1` and Zone CP is still `1.2.0`, ZoneIngresses will always be listed as offline.
After Zone CPs are upgraded to `1.2.1`, the status will work again. ZoneIngress status does not affect cross-zone traffic.

## Upgrade to `1.2.0`

One of the changes introduced by Kuma 1.2.0 is renaming `Remote Control Planes` to `Zone Control Planes` and `Dataplane Ingress` to `Zone Ingress`. 
We think this change makes the naming more consistent with the rest of the application and also removes some of unnecessary confusion.

As a result of this renaming, some values and arguments in multizone/kubernetes environment changed. You can read below more.

### Upgrading with `kumactl` on Kubernetes

1. Changes in arguments/flags for `kumactl install control-plane`

   * `--mode` accepts now values: `standalone`, `zone` and `global` (`remote` changed to `zone`)

   * `--tls-kds-remote-client-secret` flag was renamed to `--tls-kds-zone-client-secret`

2. Service `kuma-global-remote-sync` changed to `kuma-global-zone-sync` so after upgrading `global` control plane you have to manually remote old service. For example:

   ```sh
   kubectl delete -n kuma-system service/kuma-global-remote-sync 
   ```

    Hint: It's worth to remember that often at this point the IP address/hostname which is used as a KDS address when installing Kuma Zone Control Planes will change. Make sure that you update the address when upgrading the Remote CPs to the newest version.

### Upgrading with `helm` on Kubernetes

Changes in values in Kuma's HELM chart

* `controlPlane.mode` accepts now values: `standalone`, `zone` and `global` (`remote` changed to `zone`)

* `controlPlane.globalRemoteSyncService` was renamed to `controlPlane.globalZoneSyncService`

* `controlPlane.tls.kdsRemoteClient` was renamed to `controlPlane.tls.kdsZoneClient`

### Suggested Upgrade Path on Universal

1. Zone Control Planes should be started using new environment variables

   * `KUMA_MODE` accepts now values: `standalone`, `zone` and `global` (`remote` changed to `zone`)

     Old:
     ```sh
     KUMA_MODE="remote" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MODE="zone" [...] kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_ZONE` was renamed to `KUMA_MULTIZONE_ZONE_NAME`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_ZONE="remote-1" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_NAME="remote-1" [...] kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_GLOBAL_ADDRESS` was renamed to `KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_GLOBAL_ADDRESS="grpcs://localhost:5685" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS="grpcs://localhost:5685" [...]  kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE` was renamed to `KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE="/rootCa" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE="/rootCa" [...] kuma-cp run
     ```

   * `KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE` was renamed to `KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE`

     Old:
     ```sh
     KUMA_MULTIZONE_REMOTE_KDS_REFRESH_INTERVAL="9s" [...] kuma-cp run
     ```

     New:
     ```sh
     KUMA_MULTIZONE_ZONE_KDS_REFRESH_INTERVAL="9s" [...] kuma-cp run
     ```

2. Dataplane Ingress resource should be replaced with ZoneIngress resource:

    Old:
    ```yaml
    type: Dataplane
    name: dp-ingress
    mesh: default
    networking:
      address: <ADDRESS>
      ingress:
        publicAddress: <PUBLIC_ADDRESS>
        publicPort: <PUBLIC_PORT>
      inbound:
      - port: <PORT>
        tags:
          kuma.io/service: ingress
    ```

    New:
    ```yaml
    type: ZoneIngress
    name: zone-ingress
    networking:
      address: <ADDRESS>
      port: <PORT>
      advertisedAddress: <PUBLIC_ADDRESS>
      advertisedPort: <PUBLIC_PORT>
    ```

    NOTE: ZoneIngress resource is a global scoped resource, it's not bound to a Mesh
    The old Dataplane resource is still supported but it's considered deprecated and will be removed in the next major version of Kuma


3. Since ZoneIngress resource is not bound to a Mesh, it requires another token type that is bound to a Zone:
   
    ```shell
    kumactl generate zone-ingress-token --zone=zone-1 > /tmp/zone-ingress-token
    ```

4. `kuma-dp run` command should be updated with a new flag `--proxy-type=ingress`:

    ```sh
    kuma-dp run \
      --proxy-type=ingress \
      --dataplane-token-file=/tmp/zone-ingress-token \
      --dataplane-file=zone-ingress.yaml
    ```


## Upgrade to `1.1.0`

The major change in this release is the migration to XDSv3 for the `kuma-cp` to `envoy` data plane proxy communication. The
previous XDSv2 is still available and will continue working. All the existing data plane proxies will still use XDSv2 until
being restarted. The newly deployed `kuma-dp` instances will automatically get bootstrapped to XDSv3. In case that needs to be
changed, `kuma-cp` needs to be started with `KUMA_BOOTSTRAP_SERVER_API_VERSION=v2`.

With Kuma 1.1.0, the `kuma-cp` will installs default retry and timeout policies for each new
created Mesh object. The pre-existing meshes will not automatically get these default policies. If needed, they should be created accordingly.

This version removes the deprecated `--dataplane` flag in `kumactl generate dataplane-token`, please consider migrating to use `--name` instead.

## Upgrade to `1.0.0`

This release introduces a number of breaking changes. If Kuma is being deployed in production we strongly suggest to backup the current configuration, tear down the whole cluster and zones, and install in a clean setup. However, we enumerate the details of these changes below.

### Suggested Upgrade Path on Kubernetes
 * Drop k8s 1.13 support

    Take this into account if you run Kuma on an old Kubernetes version.

 * `kumactl` merged `install ingress` into `install control-plane`

    This change impacts any deployment pipelines that are based on `kumactl` and are used for multi-zone deployments.

 * Change policies on K8S to scope global

    All the CRDs are now in the global scope, therefore all policies need to be backed up. The relevant CRDs need to be deleted, which will clear all the policies. After the upgrade, you can apply the policies again. We do recommend to keep all the Kuma Control Planes down while doing these operations.

 * Autoconfigure single cert for all services

    Deployment flags for providing TLS certificates in Helm and `kumactl` have changed, refer to the relevant [documentation](https://github.com/kumahq/kuma/blob/release-1.0/deployments/charts/kuma/README.md#values) to verify the new naming.

 * Create default resources for Mesh

    The following default resources will be created upon the first start of Kuma Control Plane
        - default signing key
        - default Allow All traffic permission policy `allow-all-<mesh name>`
        - Default Allow All traffic route policy `allow-all-<mesh name>`
    
    Please verify if this conflicts with your deployment and expected policies.

 * New Multizone deployment flow

    Deploying Multizone clusters is now simplified, please refer to the deployment documentation of the updated procedure.
   
 * Improved control plane communication security
   
    Kuma Control Plane exposed ports are reduced, please revise the documentation for detailed list.
    Consider reinstalling the metrics due to the port changes in Kuma Prometheus SD.
 
 * Traffic route format
 
    The format of the TrafficRoute has changed. Please check the documentation and adapt your resources. 

### Suggested Upgrade Path on Universal
 * Get rid of advertised hostname
    `KUMA_GENERAL_ADVERTISED_HOSTNAME` was removed and not needed now.
 
 * Autoconfigure single cert for all services
    Deployment flags for providing TLS certificates in Helm and `kumactl` have changed, refer to the documentation](https://github.com/kumahq/kuma/blob/release-1.0/pkg/config/app/kuma-cp/kuma-cp.defaults.yaml) to verify the new naming.

 * Create default resources for Mesh
    
    The following default resources will be created upon the first start of Kuma Control Plane
        - default signing key
        - default Allow All traffic permission policy `allow-all-<mesh name>`
        - Default Allow All traffic route policy `allow-all-<mesh name>`
    
    Please verify if this conflicts with your deployment and expected policies.

* New Multizone deployment flow

    Deploying Multizone clusters is now simplified, please refer to the deployment documentation of the updated procedure.
   
 * Improved control plane communication security
   
    `kuma-dp` invocation has changed and now allows for a more flexible usage leveraging automated, template based Dataplane resource creation, customizable data-plane token boundaries and additional CA ceritficate validation for the Kuma Control plane boostrap server.
    Kuma Control Plane exposed ports are reduced, please revise the documentation for detailed list.
 
  * Traffic route format
  
     The format of the TrafficRoute has changed. Please check the documentation and adapt your resources. 

 
## Upgrade to `0.7.0`
Support for `kuma.io/sidecar-injection` annotation. On Kubernetes change the namespace resources that host Kuma mesh services with the aforementioned annotation and delete the label. 

Prefix the Kuma built-in tags with `kuma.io/` as follows: `kuma.io/service`, `kuma.io/protocol`, `kuma.io/zone`.

### Suggested Upgrade Path on Kubernetes

Update the applied policy tag selector to include the `kuma.io/` prefix. A sample traffic resource follows:

```yaml
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  namespace: default
  name: allow-all-traffic
spec:
  sources:
    - match:
        kuma.io/service: '*'
  destinations:
    - match:
        kuma.io/service: '*'
```

The Kuma Control Plane will update the relevant Dataplane resources accordingly

### Suggested Upgrade Path on Universal

Update the applied policy tag selector to include the `kuma.io/` prefix. A sample traffic resource follows:

```yaml
type: TrafficPermission
name: allow-all-traffic
mesh: default
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
```

Update the dataplane resources with the new tag format as well. Example:

```bash
echo "type: Dataplane
mesh: default
name: redis-1
networking:
  address: 192.168.0.1
  inbound:
  - port: 9000
    servicePort: 6379
    tags:
      kuma.io/service: redis" | kumactl apply -f -
```

This release changes the way that Distributed and Hybrid Kuma Control planes are deployed. Please refer to the documentation for more details.

## Upgrade to `0.6.0`

Passive Health Check were removed in favor of Circuit Breaking.

Format of Active Health Check changed from :
```yaml
apiVersion: kuma.io/v1alpha1
kind: HealthCheck
mesh: default
metadata:
  namespace: default
  name: web-to-backend-check
mesh: default
spec:
  sources:
  - match:
      service: web
  destinations:
  - match:
      service: backend
  conf:
    activeChecks:
      interval: 10s
      timeout: 2s
      unhealthyThreshold: 3
      healthyThreshold: 1
    passiveChecks:
      unhealthyThreshold: 3
      penaltyInterval: 5s
```
to 
```yaml
apiVersion: kuma.io/v1alpha1
kind: HealthCheck
mesh: default
metadata:
  namespace: default
  name: web-to-backend-check
mesh: default
spec:
  sources:
  - match:
      service: web
  destinations:
  - match:
      service: backend
  conf:
    interval: 10s
    timeout: 2s
    unhealthyThreshold: 3
    healthyThreshold: 1
```

### Suggested Upgrade Path on Kubernetes

In the new Kuma version serivce tag format has been changed. Instead of `backend.kuma-demo.svc:5678` service tag will look like this `backend_kuma-demo_svc_5678`. This is a breaking change and Policies should be updated to be compatible with the new Kuma version.

Please re-install Prometheus via `kubectl install metrics` and make sure that `skipMTLS` is set to `false` or omitted.
```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  metrics:
    enabledBackend: prometheus-1
    backends:
    - name: prometheus-1
      type: prometheus
      conf:
        skipMTLS: false
```

### Suggested Upgrade Path on Universal

Make sure that `skipMTLS` is set to `true`.

```yaml
type: Mesh
name: default
metrics:
  enabledBackend: prometheus-1
  backends:
  - name: prometheus-1
    type: prometheus
    conf:
      skipMTLS: true
```


## Upgrade to `0.5.0`
### Suggested Upgrade Path on Kubernetes

#### Mesh resource format changes

The Mesh resource format in Kubernetes changed from
```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabled: true
    ca:
      builtin: {}
  metrics:
    prometheus: {}
  logging:
    backends:
    - name: file-1
      file:
        path: /var/log/access.log
  tracing:
    backends:
    - name: zipkin-1
      zipkin:
        url: http://zipkin.local:9411/api/v1/spans
```
to
```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabledBackend: ca-1
    backends:
    - name: ca-1
      type: builtin
  metrics:
    enabledBackend: prom-1
    backends:
    - name: prom-1
      type: prometheus
  logging:
    backends:
    - name: file-1
      type: file
      conf:
        path: /var/log/access.log
  tracing:
    backends:
    - name: zipkin-1
      type: zipkin
      conf:
        url: http://zipkin.local:9411/api/v1/spans
```

#### Removing `kuma-injector`

Kuma 0.5.0 ships with `kuma-injector` embedded into the `kuma-cp`, which makes its previously created resources obsolete and potentially
 can cause problems with the deployments. Before deploying the new version, it is strongly advised to run a cleanup script kuma-0.5.0-k8s-remove_injector_resources.sh.
 
 NOTE: if Kuma was deployed in a namespace other than `kuma-system`, please run `export KUMA_SYSTEM=<othernamespace` before running the cleanup script.

#### Kuma resources `ownerReferences` 
Kuma 0.5.0 introduce webhook for setting `ownerReferences` to the Kuma resources. If you have some 
Kuma resources in your k8s cluster, then you can use our script kuma-0.5.0-k8s-set_owner_references.sh 
in order to properly set `ownerReferences` .

### Suggested Upgrade Path on Universal

#### Mesh resource format changes
Mesh format on Universal changed from
```yaml
type: Mesh
name: default
mtls:
  enabled: true
  ca:
    builtin: {}
metrics:
  prometheus: {}
logging:
  backends:
  - name: file-1
    file:
      path: /var/log/access.log
tracing:
  backends:
  - name: zipkin-1
    zipkin:
      url: http://zipkin.local:9411/api/v1/spans
```
to
```yaml
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
metrics:
  enabledBackend: prom-1
  backends:
  - name: prom-1
    type: prometheus
logging:
  backends:
  - name: file-1
    type: file
    conf:
      path: /var/log/access.log
tracing:
  backends:
  - name: zipkin-1
    type: zipkin
    conf:
      url: http://zipkin.local:9411/api/v1/spans
```

## Upgrade to `0.4.0`

### Suggested Upgrade Path on Kubernetes

No additional steps are needed.

### Suggested Upgrade Path on Universal

#### Migrations

Kuma 0.4.0 introduces DB Migrations for Postgres therefore before running the new version of Kuma, run the kuma-cp migration command.
```
kuma-cp migrate up
```
Remember to provide config just like in `kuma-cp run` command.
All existing data will be preserved.

#### New Dataplane Entity format

Kuma 0.4.0 introduces new Dataplane entity format to improve readability as well as add support for scraping metrics of Gateway Dataplanes. 

Here is example of migration to the new format.

**Dataplane**

Old format
```yaml
type: Dataplane
mesh: default
name: web-01
networking:
  inbound:
  - interface: 192.168.0.1:21011:21012
    tags:
      service: web
  outbound:
  - interface: :3000
    service: backend
```

New format
```yaml
type: Dataplane
mesh: default
name: web-01
networking:
  address: 192.168.0.1
  inbound:
  - port: 21011
    servicePort: 21012
    tags:
      service: web
  outbound:
  - port: 3000
    service: backend
```

**Gateway Dataplane**

Old format
```yaml
type: Dataplane
mesh: default
name: kong-01
networking:
  gateway:
    tags:
      service: kong
```

New format
```yaml
type: Dataplane
mesh: default
name: kong-01
networking:
  address: 192.168.0.1
  gateway:
    tags:
      service: kong
```

Although the old format is still supported, it is recommended to migrate since the support for it will be dropped in the next major version of Kuma.

## Upgrade to `0.3.1`

### List of breaking changes

`kuma policies`:
* `Mesh` CRD on Kubernetes is now Cluster-scoped
* `TrafficLog` policy is applied differently now: instead of applying all `TrafficLog` policies that match to a given `outbound` interface of a `Dataplane`, only a single the most specific `TrafficLog` policy is applied

`kumactl`:
* a few options in `kumactl config control-planes add` command have been renamed:
  * `--dataplane-token-client-cert` has been renamed into `--admin-client-cert`
  * `--dataplane-token-client-key` has been renamed into `--admin-client-key`

### Suggested Upgrade Path on Kubernetes

* Users on Kubernetes will have to re-install `Kuma`:

  1. Export all `Kuma` resources
     ```shell
     kubectl get meshes,trafficpermissions,trafficroutes,trafficlogs,proxytemplates --all-namespaces -oyaml > backup.yaml
     ```
  2. Uninstall previous version of `Kuma Control Plane`
     ```shell
     # using previous version of `kumactl`

     kumactl install control-plane | kubectl delete -f -
     ```
  3. Install new version of `Kuma Control Plane`
     ```shell
     # using new version of `kumactl`

     kumactl install control-plane | kubectl apply -f -
     ```
  4. Re-apply `Kuma` resources back again
     ```shell
     kubectl apply -f backup.yaml
     ```

### Suggested Upgrade Path on Universal

* Those users who used `--dataplane-token-client-cert` and `--dataplane-token-client-key` command line options in the past will have to re-run

   ```
   kumactl config control-planes add
   ```

   this time with

    ```shell
    --admin-client-cert <CERT> --admin-client-cert <KEY> --overwrite
    ```
* all components of `Kuma Control Plane` - `kuma-cp`, `kuma-dp`, `envoy` - have to be re-deployed
