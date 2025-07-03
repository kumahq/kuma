# Defining and migrating to consistent naming for non-system Envoy resources and stats

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13264

## Context and problem statement

Right now, Envoy resources and stats in Kuma use different naming styles. Some follow legacy or default Envoy formats, which makes them harder to understand and trace. Resource names and their related stats often don’t match, even when they come from the same Kuma resource. This makes it difficult to work with observability tools and troubleshoot issues.

## Affected systems and users

Two main groups are impacted:

* Internal systems: Kuma GUI, Grafana dashboards shipped with `kumactl install observability`
* External users: Dashboards, Prometheus

### Kuma GUI

The `Dataplane` view in the GUI shows inbound and outbound endpoints by parsing Envoy stat names such as clusters and listeners, and matching them to the config. Changes to how names are built will affect this matching. The GUI must be updated to support both old and new formats during the migration. This includes handling KRI-based names for outbound and routing resources, and `self_` names for inbounds. Unlike before, stat names and resource names will now match, so GUI logic must reflect that.

## Scope

The original goal of this MADR was to describe how to switch all non-system Envoy resources to use the KRI naming format. But while working on it, it became clear that using full KRI names for some resources, like inbounds and passthrough outbounds, would lead to a sharp increase in metrics cardinality without providing real value. As a result, the scope changed. This document now:

* Clearly defines how non-system Envoy resources and stats should be named
* Specifies which resources should use KRI names and which should use simpler formats
* Describes a migration path for switching to the new names safely and gradually

The goal is to make naming consistent, easy to understand, and better connected to Kuma resources, without causing problems like high metrics cardinality or churn. The new approach builds on the ideas from [MADR-070 Resource Identifier](070-resource-identifier.md) but avoids unnecessary complexity.

### Service discovery model

This decision applies only to environments using the new service discovery model, which is based on the following resources:

* `MeshService`
* `MeshExternalService`
* `MeshMultiZoneService`

These resources replace the legacy `kuma.io/service` tag previously used to define services in the mesh. Since the legacy tag is being phased out and will eventually be removed, this decision only affects the naming of resources generated from the new model.

### Observability tooling

This decision includes updates to Grafana dashboards installed via `kumactl install observability`.

While there is an open discussion and no final decision yet about the future of these observability components (see [kumahq/kuma#11693](https://github.com/kumahq/kuma/issues/11693)), they have not been officially removed or deprecated. Although current direction leans toward deprecation and shifting responsibility to users to follow setup instructions for each tool individually, this has not been finalized.

Given that:

* the changes needed to align existing dashboards with the new naming scheme are minimal
* the time effort to make those changes is low

We will update the included Grafana dashboards to use the new resource and stat names introduced in this decision document.

## Out of scope

Renaming Envoy resources and stat names generated using the legacy service discovery model based on the `kuma.io/service` tag is out of scope. This means the new, consistent and well-defined naming, described later in this document, will not apply to those resources even if the related data plane proxy feature flag is enabled.

### Resource exclusions

This document does not cover renaming of Envoy resources that:

1. Do not directly map to Kuma resources. This includes system-generated resources, some of which are already covered by [MADR-076 Standardized Naming for internal xDS Resources](076-naming-internal-envoy-resources.md), such as:

   * Secrets
   * System listeners and clusters

   Others will be covered in a separate MADR, such as:

   * Default routes (when no `MeshHTTPRoute` or `MeshTCPRoute` is defined)

2. Are related to the `MeshPassthrough` resource, which will be addressed separately.

3. Are part of the built-in gateway, which is excluded from this effort and may be handled independently.

## Design areas requiring decisions

### Naming for inbound-related resources

The current KRI naming format, as defined in the [MADR-070 Resource Identifier](070-resource-identifier.md), proposes that non-system inbound resources (such as listeners and clusters) use the full KRI name of the originating `Dataplane` plus the inbound port or name as the `sectionName`. This would result in names like:

```
kri_dp_default_kuma-2_kuma-demo_demo-app-ddd8546d5-vg5ql_5050
```

While technically correct and traceable, this significantly increases metrics cardinality compared to current formats, such as `localhost_{port}` for clusters and `{address}_{port}` for listeners, which use Envoy’s default format. This would lead to regressions in performance, cost, and compatibility with observability tooling. The following alternatives are considered:

#### Option 1: Introduce a dedicated contextual naming scheme for inbound resources

This option proposes a new, distinct naming format specifically for non-system inbound resources. Instead of reusing the full KRI format (which includes the `Dataplane` name) or the system naming (which uses the `system_` prefix), inbound resources would follow a third format with a contextual keyword prefix and either a port number or named port:

```
<keyword>_5050
<keyword>_httpport
```

This approach introduces a clear and purpose-specific naming scheme for inbound resources that are local to the `Dataplane` defining them. The justification is that inbound resources exist only within the context of the owning proxy and are not independently addressable Kuma resources. Using the KRI format without a resource reference would undermine its semantic purpose. A separate naming pattern avoids that problem.

**Benefits:**

* Avoids increase in metric cardinality
* Keeps resource and stat names short and easy to read
* Minimizes tooling impact by keeping formats simple and stable
* Makes it clear that the resource is local to the proxy, not a system or KRI-based resource
* Avoids embedding pod-specific details, preventing metric churn in Kubernetes
* Cleanly separates inbound identity from system or KRI conventions

**Drawbacks:**

* Adds a third naming convention not originally covered in the [MADR-070 Resource Identifier](070-resource-identifier.md)
* Removes the ability to link the name directly to the originating `Dataplane`
* Adds a small layer of complexity to tooling that must now recognize this third category
* Might overlap with user-defined tags or labels in some environments

##### Suboption A: Use `self` as the keyword

Examples:

```
self_5050
self_httpport
```

**Extra benefits:**

* Clearly indicates that the resource belongs to the current `Dataplane`
* Familiar and widely used keyword in tools and programming languages to refer to the current context

##### Suboption B: Use `this` as the keyword

Examples:

```
this_5050  
this_httpport
```

**Extra benefits:**

* Clearly indicates that the resource belongs to this `Dataplane`
* Familiar and widely used keyword in tools and programming languages to refer to the current context

**Extra drawbacks:**

* Might be confused with general-purpose wording in documentation or examples
* Less commonly used than `self` in resource naming contexts

##### Suboption C: Use `local` as the keyword

Examples:

```
local_5050  
local_httpport
```

**Extra benefits:**

* Clearly indicates that the resource is local to the current `Dataplane`
* Common in networking contexts

**Extra drawbacks:**

* Less familiar as a naming prefix compared to `self` or `this`
* Might be confused with `localhost` or other network terms, suggesting a different meaning than intended (e.g., local to pod or machine instead of to the current `Dataplane`)

#### Option 2: Align resource names with existing `localhost_{port}` inbound clusters stat format

This option proposes using the already established stat format `localhost_{port}` (currently used for inbound Envoy cluster stats) as the unified name for all related xDS resources such as clusters and listeners.

Specifically:

* Change cluster names from `localhost:{port}` to `localhost_{sectionName}` to match stat format
* Change listener and other inbound-related resource names (e.g. from `inbound:10.42.0.83:5050`) to `localhost_{sectionName}` for consistency across resources

Unlike current formats, this allows `{sectionName}` to be either a port number or a named port (e.g. `httpport`).

**Examples:**

```
localhost_5050
localhost_httpport
```

**Benefits:**

* No increase in metric cardinality
* Keeps resource names in sync with existing `localhost_{port}` stat format for clusters, as long as no port name is used
* Dashboards and alerts using current cluster stat names will continue to work without changes when numeric ports are used
* Builds on a format already familiar to users

**Drawbacks:**

* Listener stat names will change, which may require updates to dashboards or alerts that rely on them
* The prefix `localhost_` may be misleading for listeners, which typically bind to pod IPs, not loopback
* Breaks the original [MADR-070 Resource Identifier](070-resource-identifier.md) model by introducing a third category of resources outside Kuma-based and system types

#### Option 3: Use modified KRI with placeholder in place of `Dataplane` name

This option preserves the structure of the original KRI format but modifies the `Dataplane`-related sections to avoid high cardinality. There are three suboptions for how to handle the `Dataplane` identity portion of the name:

##### Suboption A: Use a fixed keyword or empty value instead of the `Dataplane` name

Examples:

```
kri_dp_default_kuma-2_kuma-demo__5050
kri_dp_default_kuma-2_kuma-demo_self_5050
kri_dp_default_kuma-2_kuma-demo_this_5050
kri_dp_default_kuma-2_kuma-demo_local_5050
```

**Benefits:**

* Lower cardinality compared to full `Dataplane` name
* Retains full KRI structure with readable, recognizable identity
* Clear and straightforward for users and tools to recognize and match

**Drawbacks:**

* If a user names their `Dataplane` using the keyword chosen for this naming scheme (e.g., `self`, `this`, or `local`), the generated resource names will overlap with the special inbound format. This can cause confusion, as other `Dataplanes` may include identical names that refer to different resources
* Leaving the name section empty may look broken or incomplete even if technically valid
* Breaks the original [MADR-070 Resource Identifier](070-resource-identifier.md) model by introducing a special case for inbounds that does not align with the two existing categories
* Reduces consistency in the overall naming convention
* Still includes mesh, zone, and namespace, which are already present as metric labels

##### Suboption B: Replace `Dataplane` name with a `-` to indicate hidden value

Example:

```
kri_dp_default_kuma-2_kuma-demo_-_5050
```

The `-` acts as a placeholder to signal that the value exists but is intentionally omitted or obfuscated. This approach avoids using reserved or meaningful terms like `self`.

**Benefits:**

* Lower cardinality compared to full `Dataplane` name
* Retains full KRI structure with readable, recognizable identity
* Clear and straightforward for users and tools to recognize and match
* Avoids name collision with user-defined `Dataplane` names like `self`, `this` or `local`
* Opens the path to extending the KRI format definition to support `-` as a special reserved marker for any KRI section

**Drawbacks:**

* Still includes mesh, zone, and namespace, contributing to metric cardinality
* Requires updates to the [MADR-070 Resource Identifier](070-resource-identifier.md) to define the semantics of `-`
* Less intuitive than using keywords and may need extra explanation

##### Suboption C: Collapse all KRI sections before `sectionName` into `-`

Example:

```
kri_dp_-_-_-_-_5050
```

**Benefits:**

* Keeps KRI format while minimizing cardinality
* No risk of naming collisions or ambiguity
* Can be formally described in the KRI specification as a valid but anonymized identifier

**Drawbacks:**

* Looks awkward and artificial
* May confuse users and tooling - resource names like `kri_dp_-_-_-_-_5050` are syntactically valid but visually obscure
* Loses most of the traceability benefit of KRI
* Requires KRI specification to explicitly allow this pattern and define its meaning
* Still includes the `kri_` prefix, which might falsely suggest the name is fully qualified and traceable

#### Option 4: Treat inbounds as "system" resources and use `system_{prefix}_{sectionName}` format

Adjust the [MADR-076 Standardized Naming for internal xDS Resources](076-naming-internal-envoy-resources.md) and [MADR-070 Resource Identifier](070-resource-identifier.md) by:

* Including non-system inbounds in the `system` category
* Using names like `system_self_5050` or `system_this_5050`

**Benefits:**

* No increase in metric cardinality
* Small changes to tools (e.g., replace `localhost_*` with `system_self_*`)

**Drawbacks:**

* Strongly contradicts the purpose of the system resource category, which is meant for internal, non-user-facing entities. Inbounds are tied to user configuration and traffic, so including them breaks this separation
* Users and tools that ignore `system_*` resources may unintentionally miss inbound data
* Requires documentation and tooling exceptions
* Forces major changes to the [MADR-076 Standardized Naming for internal xDS Resources](076-naming-internal-envoy-resources.md)

#### Option 5: Use full KRI name for resources, simplified name for stats

In this option, the Envoy resource names (used in xDS configuration) would follow the full KRI format as originally defined:

```
kri_dp_default_kuma-2_kuma-demo_demo-app-ddd8546d5-vg5ql_5050
```

However, the stat names would use a simplified format as described in earlier options, for example:

```
self_5050
```

**Benefits:**

* No increase in metric cardinality
* Keeps the xDS resources traceable via full KRI names
* One side (stats or resources) remains compatible with the original KRI format

**Drawbacks:**

* Breaks the promise in the [MADR-070 Resource Identifier](070-resource-identifier.md) that stat names and resource names would be aligned and consistent
* Makes it harder to correlate stats with resources in tooling like the Kuma GUI, which would now need to handle both formats separately
* Requires dual parsing and matching logic in Kuma GUI (e.g., different strategies for xDS config vs. stats)
* Increases implementation complexity and potential for confusion when debugging or inspecting proxy behavior

### Inbound section naming strategy

This section describes the problem of how to choose the `<sectionName>` for inbound resource and stat names, and presents the options under consideration.

Kuma policies can target individual inbounds on a `Dataplane` by setting the `sectionName` field in the `targetRef`. This value must match the name or port used in the inbound configuration. This is helpful when a `Dataplane` exposes multiple inbounds and the policy should apply only to one of them.

For example, the following policy applies fault injection **only to the inbound named `backend-api`**:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshFaultInjection
metadata:
  name: only-backend-api-inbound
  namespace: kuma-demo
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: backend
    sectionName: backend-api
  from:
  - targetRef:
      kind: Mesh
    default:
      http:
      - abort:
          httpStatus: 500
          percentage: '2.5'
```

There are two options for choosing the section name format:

#### Option 1: Always use the port value

Example:

```
self_5050
```

**Benefits:**

* Simple and consistent naming
* All inbounds follow the same format regardless of configuration
* Easier to recognize and match across resources and stats

**Drawbacks:**

* Does not remove support for targeting inbounds in policies, but creates inconsistency between what values can be used in `sectionName` (name or port) and the actual resource and stat names, which always use the port value in this option

#### Option 2: Use port name if defined, otherwise fall back to port value

Examples:

```
self_5050
self_httpport
```

**Benefits:**

* Supports both named and unnamed inbounds while preserving the user-defined port name in resource and stat names
* Aligns with policy `sectionName` values, allowing consistent targeting using either port name or value
* Produces more meaningful names when users define port names in Universal mode or when taken from Kubernetes service port names

**Drawbacks:**

* Results in mixed naming styles (some with names, some with port numbers), reducing overall consistency
* Can make it harder to scan and compare resource or metric names across environments where naming practices vary

### Naming for transparent proxy passthrough resources

This decision area focuses on how to name the inbound and outbound Envoy resources and stats used for IPv4 and IPv6 transparent proxy passthrough. These resources are generated when a `Dataplane` is configured with:

```yaml
networking:
  transparentProxying:
    redirectPortInbound: <port>
    redirectPortOutbound: <port>
```

These passthrough resources are tied to the current `Dataplane` and play a role in redirecting traffic through the proxy.

Two options are considered:

#### Option 1: Use `system_` prefix

Treat passthrough resources as internal system components and name them using the `system` resources formatting described in [MADR-076 Standardized Naming for internal xDS Resources](076-naming-internal-envoy-resources.md).

**Benefits:**

* Makes passthrough resources clearly distinct from user-defined service-related resources
* Matches the behavior of other resources that may or may not be used, depending on the traffic path
* Reflects the initial perception that these resources are internal to proxy functionality
* Avoids adding more variation to `self_`-prefixed resources when their observability value is uncertain in some setups

**Drawbacks:**

* These resources may sit on the regular service-to-service traffic path
* The `system_` prefix can reduce visibility and observability of traffic that users care about
* Makes it harder to trace and analyze traffic flowing through these passthrough points

#### Option 2: Use `self_` prefix

Use the contextual naming format already applied to inbounds and treat passthrough resources as part of the `Dataplane`'s scoped configuration.

**Benefits:**

* Reflects that passthrough behavior is explicitly configured in the `Dataplane`
* Groups passthrough traffic with other contextual resources using the same naming convention
* Improves observability and traceability for users monitoring traffic through these paths

**Drawbacks:**

* Slightly expands the use of the `self_` prefix beyond direct inbounds
* Requires awareness in tooling to treat these passthrough resource types appropriately

### Impact on `MeshProxyPatch` policies

Enabling the new naming formats may break existing `MeshProxyPatch` policies that rely on old resource names. These patches often target specific cluster or listener names, and any mismatch caused by the updated naming will prevent the patch from applying.

Several ideas were considered to help detect or prevent breakage, but none were viable:

* Emit a warning when a `MeshProxyPatch` matches a proxy but none of its modifiers apply. However, patches like JSON Patch are complex and hard to analyze. Logging unmatched cases would likely produce too much noise and lead to false positives.

* Scan patches for the presence of used prefixes as an indicator of awareness. This would be fragile and could not reliably detect incompatible patches.

* Show a general warning in the GUI when KRI is enabled and any `MeshProxyPatch` is active. This might help initially but would quickly become repetitive and unhelpful once users confirm their setups.

Since there’s no clean way to catch or warn about this in code, the decision is to document the risk clearly. Users are expected to:

* Review and update their `MeshProxyPatch` policies before enabling the new naming formats

As part of updating the documentation, we will:

* Add a strong warning to the `MeshProxyPatch` section explaining that policies must be reviewed and updated when switching to the new naming formats
* Include a notice in the upgrade notes to alert users of this requirement

## Decision outcome

Introduce a data plane feature flag to enable the new naming scheme for Envoy resources and stats. This includes:

* KRI-based names for resources tied to distinct Kuma resources like `MeshService`
* `self_` format for inbound-related resources defined inside a `Dataplane`

The flag can be enabled per proxy by setting an environment variable:

* Directly in Universal mode
* Indirectly in Kubernetes via a pod annotation, which is converted by the sidecar injector into the same environment variable and applied to the `kuma-sidecar` container

In Kubernetes, the control plane can be configured to auto-inject the annotation into all workloads during sidecar injection, turning the feature on by default for all proxies in the zone.

### Inbound resource naming

Inbound-related Envoy resources and stats will adopt a dedicated naming scheme based on the `self_` prefix. This avoids the high cardinality of full KRI-formatted names (e.g., `kri_dp_default_..._5050`) and uses a more contextual format `self_<sectionName>`

Examples:

```
self_5050  
self_httpport
```

This scheme clearly identifies inbound resources as local to the `Dataplane`, avoids overloading existing `kri_` or `system_` prefixes, and minimizes impact on observability tooling. It strikes a balance between clarity, performance, and compatibility.

#### Chosen section naming strategy

We will use **port name if defined, otherwise fall back to port value** for the `<sectionName>` part of inbound resource and stat names.

This ensures alignment between policy `sectionName` references and actual resource names. It allows meaningful, user-defined names where available while still working with unnamed ports. Though it introduces some inconsistency in the format, the added clarity and policy compatibility are more valuable in practice.

#### Transparent proxy passthrough resource naming

We will use the `self_<descriptor>` format (defined in the [format definition](#format-self_descriptor)), where `<descriptor>` follows the [`passthrough_ipv<IPVersion>_<direction>` format](#format-self_passthrough_ipvipversion_direction), for naming IPv4 and IPv6 passthrough-related resources and stats.

Although these resources are not always active in all environments, they are explicitly tied to the current `Dataplane` through the `transparentProxying` configuration. In setups like `MeshPassthrough`, they are heavily used and form part of the main data path. Using the `self_` prefix clearly marks them as contextual to the `Dataplane` and aligns them with other scoped resources like inbounds.

This results in the following four possible final names:

```
self_passthrough_ipv4_inbound
self_passthrough_ipv4_outbound
self_passthrough_ipv6_inbound
self_passthrough_ipv6_outbound
```

### Formal format definitions

This section defines the format rules introduced by this decision to ensure consistency and compatibility across all affected components.

#### Format: `<sectionName>`

The `<sectionName>` is a placeholder used to identify a specific part or section within a given context (such as a `Dataplane`). It is used to label a resource or stat that originates from that section of the context. For example, when used with the `self_` prefix, it refers to an inbound or passthrough-related resource in the context of the current `Dataplane`.

To support future extensibility (such as upcoming support for `MeshPassthrough`), the format allows dots (`.`), which may be useful for representing DNS-like names.

**If numeric:**

* Must consist only of digits (`0`–`9`)
* Must represent a valid port number in the range **1 to 65535**
* Leading zeros are not allowed

Example:

```
5050
```

**If named:**

* Must be **1 to 63 characters** long
* Can contain:
   * Lowercase US-ASCII letters (`a`–`z`)
   * Digits (`0`–`9`)
   * Hyphens (`-`)
   * Dots (`.`)
* Must **start with a letter**
* Must **end with a letter or digit**
* Must **not contain** consecutive hyphens (`--`)
* Must **not contain** consecutive dots (`..`)
* Must **not start or end** with a hyphen or dot

These rules combine:

* [Kubernetes Service port name requirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#serviceport-v1-core), based on **DNS_LABEL** from [RFC1123, section 2.1](https://www.rfc-editor.org/rfc/rfc1123#section-2.1)
* [Kubernetes Container port name requirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#containerport-v1-core), based on **IANA_SVC_NAME** from [RFC6335, section 5.1](https://www.rfc-editor.org/rfc/rfc6335#section-5.1)
* Additional support for dots (`.`) to allow future use cases like `MeshPassthrough` where DNS-like names are expected

Example:

```
backend-kumahq.com
```

**Regular expression:**

```
(([1-9][0-9]{0,4})|([a-z](?!.*--)(?!.*\.\.)[a-z0-9.-]{0,61}[a-z0-9]))
```

#### Format: `<self_passthrough_ipv<IPVersion>_<direction>>`

This format is used for naming resources and stats related to transparent proxy passthrough traffic in the context of the current `Dataplane`.

**Structure:**

* Starts with the literal prefix `self_passthrough_`
* Followed by `ipv4` or `ipv6` to indicate the IP version
* Ends with either `inbound` or `outbound` to indicate the traffic direction

This format is reserved and fixed for passthrough-related resources.

Examples:

```
passthrough_ipv4_inbound  
passthrough_ipv6_outbound
```

These names are not user-configurable and are generated based on the `Dataplane`'s `transparentProxying` configuration.

##### Regular expression

```
passthrough_ipv(4|6)_(inbound|outbound)
```

#### Format: `self_<descriptor>`

This format is used for naming resources and stats that are specific to the context of the current `Dataplane`. The part after `self_` is called a `<descriptor>`, and it identifies the relevant section of the `Dataplane` configuration that the resource belongs to.

**Structure:**

* Starts with the literal prefix `self_`
* Followed by a `<descriptor>` value

A `<descriptor>` is any string that represents a specific part of the `Dataplane` context and allows policies, resources, and stats to be correlated. For example, in the case of non-system inbounds, the `<descriptor>` is a **section name**, which may be a port number or port name.

All descriptors must follow the common character set rules used for resource names in [MADR-070 Resource Identifier](https://github.com/kumahq/kuma/blob/df678629aa3bdce62a84c0e3752c7e5d45bf1e98/docs/madr/decisions/070-resource-identifier.md#url-query), which defines what is allowed in the URL query component. These characters include lowercase letters (`a–z`), digits (`0–9`), hyphens (`-`), underscores (`_`), and dots (`.`), depending on the specific format.

Examples:

```
self_5050  
self_http  
self_backend-kumahq.com
self_passthrough_ipv4_outbound
```

##### Valid formats for `<descriptor>`

The exact format of `<descriptor>` depends on the type of resource:

* **For non-system inbounds**: use a [`<sectionName>`](#format-sectionname), which can be either a port number or a port name
* **For transparent proxy passthrough resources**: use the [`<passthrough_ipv<IPVersion>_<direction>`](#format-self_passthrough_ipvipversion_direction) format, where `<IPVersion>` is `4` or `6` and `<direction>` is `inbound` or `outbound`

All valid `<descriptor>` values used in `self_<descriptor>` must match one of the formats listed above.

## Implications for Kong Mesh

The changes introduced by this document, along with those defined in [MADR-076 Standardized Naming for internal xDS Resources](076-naming-internal-envoy-resources.md), impact the Envoy resources generated by the `MeshGlobalRateLimit` policy. This policy currently adds a cluster with a static name `meshglobalratelimit:service`, which is then referenced in route-level [rate limit configuration](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-rate-limits). As a result of these changes, the cluster name will be updated to follow the standard system format, using the valid `kri` of the `MeshGlobalRateLimit` policy that is the source of this cluster:

```
system_kri_mglr___<namespace>_<policy_name>_
```

Example:

```
system_kri_mglr___kong-mesh-system_mesh-rate-limit_
```

The `MeshOPA` policy modifies existing non-system inbound listeners but does not rename any existing Envoy resources or create new ones. Therefore, it is not affected by the changes described in this document.

Other Kong Mesh–specific features or policies do not rely on or modify Envoy resource names or stat prefixes that fall within the scope of this document.

## Implementation

### Data plane feature flag and propagation

A new data plane feature flag will control the use of consistent naming for proxy resources and stats, including both KRI-based formats and the new `self_` format for inbounds. It will be optional at first, with a deprecation path: the new naming will become the default in the future, and the flag will eventually be removed. This gives users time to migrate without breaking existing setups.

The flag will be passed from the data plane proxy to the control plane via [xDS metadata](https://github.com/kumahq/kuma/blob/c61baf6110caa40eb3b69f7f635e9506389a7455/app/kuma-dp/cmd/run.go#L172) using:

```go
const FeatureUnifiedProxyResourcesAndStatsNaming string = "feature-unified-proxy-resources-and-stats-naming"
```

#### Per-proxy opt-in

Users can enable the feature for individual data plane proxies:

| Mode       | How to enable                                                                               |
|------------|---------------------------------------------------------------------------------------------|
| Universal  | Set env var: `KUMA_DATAPLANE_RUNTIME_UNIFIED_PROXY_RESOURCES_AND_STATS_NAMING_ENABLED=true` |
| Kubernetes | Add annotation: `features.kuma.io/unified-proxy-resources-and-stats-naming: "enabled"`      |

In Kubernetes, we cannot rely on setting the environment variable directly because environment variables are container-scoped, not pod-scoped. Since the `kuma-sidecar` container is injected into the pod by the sidecar injector, setting the environment variable on the user’s workload container would not affect the sidecar container. Therefore, the only viable and generic way to control this feature per pod is to use an annotation. The sidecar injector reads the annotation and converts it into the correct environment variable on the `kuma-sidecar` container. `kuma-dp` will then pick up the variable and include the feature flag in the xDS metadata sent to the control plane.

**This annotation follows a new convention for data plane feature flags: all new per-pod feature flags will use the `features.` prefix.** This aligns with Kubernetes annotation guidelines, which recommend that prefixes carry context and follow the pattern `subsystem.kubernetes.io/parameter` over flat or ambiguous keys. See [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#label-selector-and-annotation-conventions) for details.

#### Zone-wide opt-in for Kubernetes data planes

To enable the naming changes for all injected data plane proxies in a Kubernetes zone, users can configure the control plane to automatically inject the annotation during sidecar injection:

```go
UnifiedProxyResourcesAndStatsNamingEnabled bool `json:"unifiedProxyResourcesAndStatsNamingEnabled" envconfig:"KUMA_RUNTIME_KUBERNETES_INJECTOR_UNIFIED_PROXY_RESOURCES_AND_STATS_NAMING_ENABLED"`
```

#### ZoneIngress and ZoneEgress

`ZoneIngress` and `ZoneEgress` proxies must also include the feature flag to ensure consistent naming accross the mesh. In both Universal and Kubernetes, this is done by setting the following environment variable in their deployments:

```env
KUMA_DATAPLANE_RUNTIME_UNIFIED_PROXY_RESOURCES_AND_STATS_NAMING_ENABLED=true
```

#### Helper setting for Kubernetes installations

To simplify configuration, a new setting will be introduced under the `dataPlane.features` section for Helm and other Kubernetes-based install methods:

```yaml
dataPlane:
  features:
    unifiedProxyResourcesAndStatsNaming: true
```

When enabled, it will:

* Add `KUMA_DATAPLANE_RUNTIME_UNIFIED_PROXY_RESOURCES_AND_STATS_NAMING_ENABLED=true` to all `ZoneIngress` and `ZoneEgress` deployments
* Set `KUMA_RUNTIME_KUBERNETES_INJECTOR_UNIFIED_PROXY_RESOURCES_AND_STATS_NAMING_ENABLED=true` in the control plane deployment

When the new naming becomes the default, this setting will also default to `true`. Eventually, it will be removed when the feature can no longer be disabled.

### Updating ZoneIgress and ZoneEgress insight resources with feature flags

To support the Kuma GUI in adapting to KRI-based naming, we need to expose feature flag information in `ZoneIngressInsight` and `ZoneEgressInsight` resources, similar to how it's already done for `DataplaneInsight`. These insights resources are available via the control plane API and provide a summary of runtime state and metadata.

We will update both `ZoneIngressInsight` and `ZoneEgressInsight` to include metadata with active feature flags. This will allow the GUI to detect whether KRI naming is enabled for each proxy and adjust its behavior accordingly.

## Migration

### Kuma GUI

The Kuma GUI relies on Envoy stat names and xDS resource names to display inbound and outbound endpoint details for `Dataplane`, `ZoneIngress`, and `ZoneEgress` proxies. These names are used to associate metrics with specific resources and visualize traffic paths in the GUI.

To support the KRI naming format, the GUI parsers that extract resource information from stat names and xDS configuration (listeners, clusters, endpoints) must be updated. The new format is defined in the [MADR-070 Resource Identifier](070-resource-identifier.md):

```
kri_<resource-type>_<mesh>_<zone>_<namespace>_<resource-name>_<section-name>
```

When the `feature-unified-proxy-resources-and-stats-naming` flag is present in the metadata of `DataplaneInsight`, `ZoneIngressInsight`, and `ZoneEgressInsight`, the GUI must:

* Parse and interpret stat names and xDS resource names based on the KRI format
* Drop any assumptions about `inbound:` or `outbound:` prefixes, which are no longer used
* Match stat names directly to xDS resource names, which now correspond 1:1

**Examples of KRI-formatted resource names:**

* `kri_dp_mesh-1_us-east-2_kuma-demo_backend-app_8080`
* `kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport`
* `kri_mzsvc_mesh-1__kuma-system_backend-app_8080`
* `kri_mhttpr_mesh-1_us-east-2_kuma-demo_route-1_`
* `kri_extsvc_mesh-1__kuma-system_es1_`

**Examples of stats using KRI-formatted names:**

* `cluster.kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport.upstream_cx_active: 0`
* `cluster.kri_mzsvc_mesh-1__kuma-system_backend-app_8080.upstream_cx_active: 0`
* `cluster.kri_dp_mesh-1_us-east-2_kuma-demo_backend-app_8080.upstream_cx_active: 0`
* `http.kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport.downstream_cx_active: 0`
* `http.kri_mzsvc_mesh-1__kuma-system_backend-app_8080.downstream_cx_active: 0`
* `http.kri_dp_mesh-1_us-east-2_kuma-demo_backend-app_8080.downstream_cx_active: 0`
* `listener.kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport.downstream_cx_active: 0`
* `listener.kri_mzsvc_mesh-1__kuma-system_backend-app_8080.downstream_cx_active: 0`
* `listener.kri_dp_mesh-1_us-east-2_kuma-demo_backend-app_8080.downstream_cx_active: 0`

## Test scenarios required for completion

The following scenarios must be verified to consider the work complete. Each case ensures the correct generation and usage of KRI-based resource and stat names across various deployment modes.

### Single-zone deployments

* `MeshService` is targeted directly by a `Dataplane` in the same zone
* `MeshExternalService` is targeted by a `Dataplane` via the `ZoneEgress` in the same zone

### Multi-zone with ZoneEgress and ZoneIngress (Global + 2 zones)

* `Dataplane` in zone 1 targets a `MeshService` in zone 2
* `Dataplane` in zone 1 targets a `MeshExternalService` in zone 1
* `Dataplane` in zone 1 targets a `MeshMultiZoneService` in zone 2
* `Dataplane` in zone 1 targets a `MeshMultiZoneService` in zone 2 with locality awareness disabled via `MeshLoadBalancingStrategy`

### Multi-zone with only ZoneIngress (Global + 2 zones)

* `Dataplane` in zone 1 targets a `MeshService` in zone 2
* `Dataplane` in zone 1 targets a `MeshMultiZoneService` in zone 1
* `Dataplane` in zone 2 targets a `MeshMultiZoneService` in zone 2 with locality awareness disabled via `MeshLoadBalancingStrategy`

### Multi-zone with ZoneIngress in both zones and ZoneEgress in one zone (Global + 2 zones)

* `Dataplane` in zone 1 targets a `MeshService` in zone 2
* `Dataplane` in zone 2 targets a `MeshService` in zone 1
* `Dataplanes` in both zones target a `MeshMultiZoneService` in their respective zones
* `Dataplanes` in both zones target a `MeshMultiZoneService` with locality awareness disabled via `MeshLoadBalancingStrategy`
