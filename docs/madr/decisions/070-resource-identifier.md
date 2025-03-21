# Resource Identifier

* Status: accepted

## Out of Scope

This document defines the format of the resource identifier only. 
It does not cover the rollout strategy for adopting the identifier across Envoy configurations, 
nor does it define the shape or behavior of HTTP APIs for retrieving resources by resource identifier. 
Although we reference potential usages such as Envoy naming or future Inspect APIs, 
these are mentioned solely to illustrate the motivation. 
Renaming secrets in Envoy is also out of scope.

## Context and Problem Statement

Kuma multizone mode allows synchronizing state between multiple clusters. Zone Control Planes connect to the Global Control Plane, forming a tree with the Global CP as the root.

There are several resource synchronization patterns:
- A resource is created on the Global CP and synced down to all Zone CPs.
- A resource is created on a Zone CP and synced to the Global CP.
- A resource is created on a Zone CP and synced to other Zone CPs.

For example, the Zone CP in `zone-1` can possess the following MeshTimeouts:
- MeshTimeouts created in `zone-1`
- MeshTimeouts created on the Global CP and synced to `zone-1`
- MeshTimeouts created in `zone-2` and synced to `zone-1`

To uniquely identify a resource regardless of its place of origin, Kuma uses the following Go structure:

```go
type ResourceIdentifier struct {
    Name      string
    Mesh      string
    Namespace string
    Zone      string // when resource is created on Global zone is empty
}
```

When mixing identifiers of different resource types, an extended version is used:

```go
type TypedResourceIdentifier struct {
    ResourceIdentifier

    ResourceType ResourceType
    SectionName  string
}
```

Currently, the resource identifier lacks a string representation and is not exposed to the Kuma public API.
Without a standard way to uniquely identify resources, the same problem gets solved repeatedly in different places in different ways. 

### Problems

#### Kuma API endpoint require core name

If you want to fetch a resource you have to create a URL path `:5681/meshes/<mesh>/<resource-type>/<core-name>`.
Core name is very complex and can take different forms depending on the resource origin and on the environment:
* on Universal, it's `<name>`, it is a column in PostgreSQL `resources` table
* on Kubernetes, core name is `<name>.<namespace>`, name and namespace from metadata
* when resource is synced from another cluster core name contains hash-suffix

For example, when running Global on Kubernetes the core name for synced DPP might look like `my-dpp-0-x4x9bbd4vw89f4fx.kuma-system`.

#### [Issue #2363](https://github.com/kumahq/kuma/issues/2363): Stop using ':' in envoy resource names

Envoy admin pages use `:`
```bash
$ curl http://localhost:9901/listeners
inbound:127.0.0.1:8000::127.0.0.1:8000
```

There is a workaround to use `format=json`
```bash
$ curl http://localhost:9901/listeners\?format\=json
{
 "listener_statuses": [
  {
   "name": "inbound:127.0.0.1:8000",
   "local_address": {
    "socket_address": {
     "address": "127.0.0.1",
     "port_value": 8000
    }
   }
  }
 ]
}
```

#### [Issue #3249](https://github.com/kumahq/kuma/issues/3249): local envoy clusters should have a better name than localhost_<port> 

Today inbound clusters have name `localhost:<port>` and in stats they're visible as `localhost_<port>`.
That's not very informative.

#### [Issue #12733](https://github.com/kumahq/kuma/issues/12733): Envoy statPrefix should not use ip addresses

Using IP address in envoy config statPrefix field can cause issues or generate unnecessary changes in metrics.
This is also problematic for our e2e tests that test envoy config using golden files.

#### [Issue #12093](https://github.com/kumahq/kuma/issues/12093): xds configs, outbound listeners should use the clustername instead of an IP/port combo

We name outbounds like `outbound:10.43.205.116:6379` where IP address doesn't give any useful information.

### Places to use resource identifier

#### URL path

For example `:5681/path-prefix/<identifier>/path-suffix`. 
Charset is defined by [RFC 3986](https://datatracker.ietf.org/doc/html/rfc3986#section-3.3):
```
path          = path-abempty    ; begins with "/" or is empty
              / path-absolute   ; begins with "/" but not "//"
              / path-noscheme   ; begins with a non-colon segment
              / path-rootless   ; begins with a segment
              / path-empty      ; zero characters

path-abempty  = *( "/" segment )
path-absolute = "/" [ segment-nz *( "/" segment ) ]
path-noscheme = segment-nz-nc *( "/" segment )
path-rootless = segment-nz *( "/" segment )
path-empty    = 0<pchar>

segment       = *pchar
segment-nz    = 1*pchar
segment-nz-nc = 1*( unreserved / pct-encoded / sub-delims / "@" )
              ; non-zero-length segment without any colon ":"

pchar         = unreserved / pct-encoded / sub-delims / ":" / "@"

unreserved    = ALPHA / DIGIT / "-" / "." / "_" / "~"
pct-encoded   = "%" HEXDIG HEXDIG
sub-delims    = "!" / "$" / "&" / "'" / "(" / ")"
                 / "*" / "+" / "," / ";" / "="
```

#### URL query

There are no immediate plans to put resource identifier in URL query, but if we want in the future then we have to limit the charset to:
```
query       = *( pchar / "/" / "?" )
```

#### Envoy resource names

For the context, there is already [MADR](036-internal-listeners.md) that regulated the name of internal listeners.
Also, there was [work](https://docs.google.com/document/d/1OIZK82Tr-4El2FfdlBn7WNRZ7FatkTuEcZKH0FlSTMA/edit?tab=t.0#heading=h.n6cmlf1eel2z) related to Envoy cluster name unification, but it's not finished.
Discoveries in this work helped me to fill the tables.

There are no restriction on the name format from the Envoy's side.
Some Envoy resources could be directly correlated with Kuma resources and that's why we should consider renaming them using the resource identifier.
The tables below are an inventory of resource usage for all Kuma proxies. Resources in `Internal` table exist in all proxies. 
Column "Correlated Resources" provides the Kuma resources that could be used for naming. 

##### Sidecar

|                   | Name                                                                                                                                                                                                    | Correlated Resources                                                                                                                     | ResourceIdentifer                                    |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------|
| Inbound Listener  | `inbound:10.43.205.116:8080`<br>`inbound:[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080`                                                                                                                | Dataplane (with sectionName to select port)                                                                                              | kri_dp_mesh-1_us-east-2_kuma-demo_backend-app_8080   |
| Outbound Listener | `outbound:10.43.205.116:8080`<br>`outbound:[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080`                                                                                                              | Mesh*Service (with sectionName to select port)                                                                                           | kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport |
| VirtualHost       | legacy listeners - `<kuma.io/service>`<br>new outbounds - `<mesh>_<name>_<namespace>_<zone>_<short-name>_<port>`                                                                                        | Mesh*Service (with sectionName to select port)                                                                                           | kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport |
| Inbound Cluster   | `localhost:<port>`                                                                                                                                                                                      | Dataplane (with sectionName to select port)                                                                                              | kri_dp_mesh-1_us-east-2_kuma-demo_backend-app_8080   |
| Outbound Cluster  | legacy clusters - `<kuma.io/service>-hash(dst.tags)`<br>legacy clusters cross-mesh - `<kuma.io/service>-hash(dst.tags)_<mesh>`<br>new clusters - `<mesh>_<name>_<namespace>_<zone>_<short-name>_<port>` | Mesh*Service (with sectionName to select port)                                                                                           | kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport |
| Route             | Routes are set on Listener on VirtualHost.<br>On inbound - `inbound:<kuma.io/service>`<br>On outbound - `<hash_sha256([]Match{...})>`                                                                   | Envoy Route is a merge product of multiple MeshHTTPRoutes.<br/> We can use only the most specific MeshHTTPRoute policy to name the route | kri_mhttpr_mesh-1_us-east-2_kuma-demo_route-1_       |
| Secret            | `name       = <category>:<scope>:<identifier>`<br>`category   = "mesh_ca" \| "identity_cert"`<br>`scope      = "secret"`<br>`identifier = "all" \| <mesh_name>`                                         | See [Secrets naming in Envoy](#secrets-naming-in-envoy)                                                                                  | –                                                    |

##### Builtin Gateway

|             | Name                                                                                                                                                                                                                           | Correlated Resources                                                                                                                           | ResourceIdentifier                             |
|-------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------|
| Listener    | `<gateway-name>:<protocol>:<port>` where `gateway-name` is `MeshGatewayResource.Meta.Name`                                                                                                                                     | MeshGateway (with sectionName to select the listener)                                                                                          | kri_mgw_mesh-1_us-east-2__gw-1_                |
| VirtualHost | `<hostname>`                                                                                                                                                                                                                   |                                                                                                                                                |                                                |
| Cluster     | `<kuma.io/service>-hash(merge(dpp.tags, gateway.tags, listener.tags))`                                                                                                                                                         | Pair of MeshGateway and Mesh*Service, can be only Mesh*Service once we resolve the [Issue #13129](https://github.com/kumahq/kuma/issues/13129) |                                                |
| Route       | Dynamic routes using RDS. <br/>Listener name + `:*`, i.e. `gateway-proxy:HTTP:8080:*` or<br>listener name + `:<hostname>`                                                                                                      | Envoy Route is a merge product of multiple MeshHTTPRoutes.<br/> We can use only the most specific MeshHTTPRoute policy to name the route       | kri_mhttpr_mesh-1_us-east-2_kuma-demo_route-1_ |
| Secret      | `name     = <category>:<scope>:<identifier>`<br>`category = "cert." \| "cert.ecdsa" \| "cert.rsa"`<br>`scope    = "file" \| "inline" \| "inlineString"`<br>`identifier = <file-name> \| <secret-name> \| join(hostnames, ":")` | See [Secrets naming in Envoy](#secrets-naming-in-envoy)                                                                                        | –                                              |


##### Zone Ingress

|          | Name                                                                                                                        | Correlated Resources                           | ResourceIdentifier                                   |
|----------|-----------------------------------------------------------------------------------------------------------------------------|------------------------------------------------|------------------------------------------------------|
| Listener | `inbound:10.43.205.116:10001`<br>`inbound:[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:10001`                                  | ZoneIngress                                    | kri_zi__us-east-2_kuma-system_zi1_                   |
| Cluster  | legacy services - `<mesh>:<kuma.io/service>`<br>new Mesh*Services - `<mesh>_<name>_<namespace>_<zone>_<short-name>_<port>`  | Mesh*Service (with sectionName to select port) | kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport |

##### Zone Egress

|             | Name                                                                                                                                                            | Correlated Resources                                     | ResourceIdentifier                                   |
|-------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------|------------------------------------------------------|
| Listener    | `inbound:10.43.205.116:10001`<br>`inbound:[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:10001`                                                                      | ZoneEgress                                               | kri_ze__us-east-2_kuma-system_ze1_                   |
| Cluster     | legacy services - `<mesh>:<kuma.io/service>`<br>new Mesh*Services - `<mesh>_<name>_<namespace>_<zone>_<short-name>_<port>`                                      | Mesh*Service (with sectionName to select port)           | kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport |
| FilterChain | for external services - `<kuma.io/service>_<mesh>`                                                                                                              | MeshExternalService                                      | kri_extsvc_mesh-1___es1_                             |
| Route       | for external services - `outbound:<kuma.io/service>`                                                                                                            | MeshExternalService                                      | kri_extsvc_mesh-1___es1_                             |
| Secret      | `name       = <category>:<scope>:<identifier>`<br>`category   = "mesh_ca" \| "identity_cert"`<br>`scope      = "secret"`<br>`identifier = "all" \| <mesh_name>` | See [Secrets naming in Envoy](#secrets-naming-in-envoy)  | –                                                    |

##### MeshPassthrough

|             | Name                                                                                                                                                                                                                                                                                                                                                                                                                               | Correlated Resources |
|-------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------|
| Cluster     | `meshpassthrough_<protocol>_<match-value>_<port>`<br>when `<port> == 0` Kuma sets port equal to `*`<br>`match-value = CIDR \| IP \| Domain`<br>`CIDR     = i.e. "192.0.2.0/24" or "2001:db8::/32"`<br>`IP       = i.e. "192.0.2.1", or 2001:db8::68", or ::ffff:192.0.2.1"`<br>`Domain   = <dns-name> \| *.<dns-name>`<br>`dns-name = ^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`<br>  | –                    |
| FilterChain | when `<match-value> == <protocol>` -> `meshpassthrough_<protocol>_<port>`<br>when `<match-value> != <protocol>` -> `meshpassthrough_<protocol>_<match-value>_<port>`<br>when port is 0 we put `*`, i.e. `meshpassthrough_http_*`<br>`<match-value>` is defined above                                                                                                                                                               | –                    |

##### Internal

|                  | Name                                                                                                                                                                 | Correlated Resources |
|------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------|
| Listener         | `kuma:dns`<br>`kuma:envoy:admin`<br>`kuma:metrics:prometheus`<br>`_kuma:metrics:opentelemetry:<backendName>`                                                         | –                    |
| VirtualHost      | listener's name                                                                                                                                                      | –                    |
| Internal Cluster | `kuma:readiness`<br>`kuma:envoy:admin`<br>`kuma:metrics:hijacker`<br>`_kuma:metrics:opentelemetry:<backendName>`<br>`tracing:`<br>`access_log_sink`<br>`ads_cluster` | –                    |

#### Envoy stats

There are 2 fields to specify alternative name for emitting stats – `stat_prefix`, `alt_stat_name`.
Field `alt_stat_name` is used only for Cluster, and it's automatically converts all `:` to `_` by Envoy when emitting statistics.
Field `stat_prefix` is used for HTTPConnectionManager, TCPProxy, KafkaBroker (network filter) and RBAC (network filter).
Unlike `alt_stat_name` the field `stat_prefix` doesn't have any restrictions from Envoy.
Kuma sanitizes both fields using `SanitizeMetric` function before setting them on Envoy's resource:

```go
var illegalChars = regexp.MustCompile(`[^a-zA-Z_\-0-9]`)

// We need to sanitize metrics in order to  not break statsd and prometheus format.
// StatsD only allow [a-zA-Z_\-0-9.] characters, everything else is removed
// Extra dots breaks many regexes that converts statsd metric to prometheus one with tags
func SanitizeMetric(metric string) string {
	return illegalChars.ReplaceAllString(metric, "_")
}
```

As the goal of this MADR is to have unified names in Envoy, Metrics, and API,
we want `stat_prefix` and `alt_stat_name` to always be the same as `name`.
This makes the use of `stat_prefix` and `alt_stat_name` unnecessary.
Once the migration to the new naming is complete, Kuma won't set these fields anymore.

### Constraints

#### What characters we can use in the identifier?
 
In URL path segment we can use:

```
ALPHA / DIGIT / "-" / "." / "_" / "~" / "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "," / ";" / "=" / ":" / "@"
```

In Envoy resource names we can use any character except `:` if we want to solve [Issue #2363](https://github.com/kumahq/kuma/issues/2363) without workaround.

In Envoy stats fields we can use any character except `:` to avoid on the fly conversion of `:` to `_` for Cluster names in Prometheus label values.

OpenTelemetry defines "Attribute", it's a key-value pair similar to Prometheus labels.
There are [no charset limitation on attribute's key or value](https://opentelemetry.io/docs/specs/otel/common).
SDK [provides a way](https://opentelemetry.io/docs/specs/otel/common/#attribute-limits) to configure attribute length limit,
but it's set to `Infinity` by default.

This leaves us with the following charset:

```
resource-identifier = *(ALPHA / DIGIT / "-" / "." / "_" / "~" / "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "," / ";" / "=" / "@")
```

_Note: if we'd like to use resource identifier in URL query the charset is significantly smaller:_

```
resource-identifier = *(ALPHA / DIGIT / "-" / "." / "_" / "~" )
```

#### What characters we can use as a delimiter?

As a delimiter we can use only characters that can't be present in resource identifier components – `name`, `mesh`, `namespace`, `zone`, `resourceType` and `sectionName`:

* `name` and `zone`
  * contain no more than 253 characters
  * contain only lowercase alphanumeric characters, '-' or '.'
  * start with an alphanumeric character
  * end with an alphanumeric character
* `mesh` and `namespace`
  * contain at most 63 characters
  * contain only lowercase alphanumeric characters or '-'
  * start with an alphanumeric character
  * end with an alphanumeric character
* `resourceType`
  * only alphanumeric
* `sectionName` (even though [ContainerPort's name](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#containerport-v1-core) is limited by 15 characters, [ServicePort's name](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#serviceport-v1-core) can contain upto 63 character)
  * contain at most 63 characters
  * contain only lowercase alphanumeric characters or '-'
  * start with an alphanumeric character
  * end with an alphanumeric character

Since we want to use resource identifier in URL query as well, this leaves us with the following charset:

```
delimiter = "_" / "~" 
```

## Decision Drivers

- New Inspect API endpoints that'd allow getting resolved configuration by the resource identifier
- Kuma API GET endpoints shouldn't require core name to read the resource 
- Human-readable: users should be able to type the identifier manually if needed, so it's not hashed
- Envoy resources that have direct correlation with Kuma resources should be named by using resource identifier, i.e. outbound cluster can be named after MeshService
- Ability to use resource identifier in Prometheus queries, i.e
  - `my_metric{envoy_cluster_name~="kri.*_meshservice_.*"} // Get only actual outbounds from meshservice`
  - `my_metric{envoy_cluster_name~="kri_default_.*.*"} // Get only stuff on the default mesh`

## Considered Options

* Option 1 - Order-based, no field names in the identifier
* Option 2 - Field names in the identifier

## Decision Outcome

* Option 1 - Order-based, no field names in the identifier

## Pros and Cons of the Options

### Option 1 - Order-based, no field names in the identifier

We need to pick a delimiter that's present in the `delimiter` charset we've specified above.

There is an identifier format from Amazon called [ARN](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference-arns.html). We can adopt a similar approach, but using `_`:

```
kri_<resource-type>_<mesh>_<zone>_<namespace>_<resource-name>_<section-name>
```

`resource-type` is a short name of the resource type, it's already part of the [ResourceTypeDescriptor](https://github.com/kumahq/kuma/blob/d7ec0a2b1ac19208fb7dd9726309e3cf8cdc5848/pkg/core/resources/apis/meshservice/api/v1alpha1/zz_generated.resource.go#L168)

When a field is not present, it must be represented as an empty string between underscores, even if this results in consecutive underscores or trailing underscores. 
This ensures the format remains consistent and field positions are preserved.

For example:
```
kri_msvc_mesh-1_us-east-2_kuma-demo_backend_
kri_msvc_mesh-1_us-east-2_kuma-demo_backend_http-port
kri_msvc_mesh-1___global-timeouts_
```

Having a prefix like `kri` (Kuma Resource Identifier) is useful for two reasons:
* It visually clarifies the format for users, who can then search for the format description in the documentation.
* It acts as an implicit version. If we need to update the format, we can use a different prefix (e.g., `uri` or `ri`).
* It's also ensure no overlap between internal entities (prefixed by `_kuma`) and the rest

#### Secrets naming in Envoy

Renaming secrets in Envoy is out of the scope of this MADR. 
Only `mesh_ca` secret in `builtin` mode has a corresponding Secret resource in the store.
In all other cases, secrets are either stored in memory, or obtained using DataSource or fetched from external systems,
so it's not straightforward how to apply `kri` concept.  
Also, it's not clear what problems and limitations do we have when it comes to naming secrets in Envoy.

**Pros:**
- Shorter
- Resembles existing formats from Amazon and Konnect KRN 
- Still possible to use in URL query

**Cons:**
- Hard to read when names are poorly chosen, e.g., `kri_msvc_default_default_default_backend_`

### Option 2 - Field names in the identifier

We need to pick two delimiters: one to separate keys from values and another to separate key-value pairs.

```yaml
kri;meshservice;mesh=mesh-1;zone=zone-1;namespace=kuma-demo;name=backend;section=http-port
```

**Pros:**
- Better handling of gaps; no need for `;;` when a value is not defined

**Cons:**
- Longer
- The order of fields still matters if we want to compare identifiers with `==`
- Very hard to write good prom regexps to select all matching labels.
