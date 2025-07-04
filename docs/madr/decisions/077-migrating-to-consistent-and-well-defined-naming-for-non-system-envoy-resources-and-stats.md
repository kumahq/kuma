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

The `Dataplane` view in the GUI shows inbound and outbound endpoints by parsing Envoy stat names such as clusters and listeners, and matching them to the config. Changes to how names are built will affect this matching. The GUI must be updated to support both old and new formats during the migration.

## Scope

The original goal of this MADR was to describe how to switch all non-system Envoy resources to use the KRI naming format. But during the investigation, it became clear that using full KRI names for some resource types, such as inbounds and transparent proxy passthrough resources, would lead to a sharp increase in metrics cardinality without providing clear value. This would negatively affect observability systems and increase cost. As a result, the scope of the work changed.

This MADR now:

* Defines how non-system Envoy resources and stat names should be structured
* Introduces and formally defines the `self_<descriptor>` naming format for resources that exist only in the context of the current `Dataplane`
* Defines valid formats for `<descriptor>`, including:
   * Port names or values for non-system inbounds
   * Predefined values for passthrough traffic (e.g., `passthrough-ipv4-inbound`)
* Specifies which resource types must use the `self_<descriptor>` format
* Describes when KRI format (defined in [MADR-070](070-resource-identifier.md)) is used for resources tied to distinct Kuma objects
* Provides a migration path to switch to the new formats safely and gradually

The decision applies only to environments using the new service discovery model based on the following resources:

* `MeshService`
* `MeshExternalService`
* `MeshMultiZoneService`

These resources replace the legacy `kuma.io/service` tag, which is being phased out and will eventually be removed. The new naming formats apply only to resources generated from this new model. Environments still using the legacy tag are out of scope.

This decision also includes updates to the Grafana dashboards installed via `kumactl install observability`. While the long-term future of these observability components is still being discussed (see [kumahq/kuma#11693](https://github.com/kumahq/kuma/issues/11693)), they are not yet deprecated or removed. Given the limited scope of changes and the relatively low effort involved, the included dashboards will be updated to support the new naming formats.

## Out of scope

The changes described in this document do not apply to all types of Envoy resources. The following are explicitly out of scope:

* **Legacy service discovery model**: Resources and stat names generated using the `kuma.io/service` tag are excluded. The new naming formats only apply to environments using the new service discovery model (`MeshService`, `MeshExternalService`, `MeshMultiZoneService`). Even if the related data plane proxy feature flag is enabled, legacy-tagged resources will continue using their existing names.

* **System-generated resources**: These are not renamed here. Some are already covered by [MADR-076 Standardized Naming for internal xDS Resources](076-naming-internal-envoy-resources.md), including:

   * Secrets
   * System listeners and clusters

  Others will be addressed in a separate MADR, including:

   * Default routes when no `MeshHTTPRoute` or `MeshTCPRoute` is defined

* **MeshPassthrough**: Resources related to the `MeshPassthrough` policy are not included in this decision and will be addressed separately.

* **Built-in gateway**: Gateway-related resources are excluded from this effort and may be handled independently.

## Design areas and decisions

### Naming format for contextual resources

The KRI format from [MADR-070 Resource Identifier](070-resource-identifier.md) is used for resources tied directly to distinct Kuma resources, like `MeshService`. It includes the mesh, zone, namespace, and resource name. While some resources could technically be named using the `KRI` format, this document defines that they should not. Using `KRI` for these context-specific resources leads to unnecessary metric cardinality, longer and less readable names, and extra complexity without real value.

Some Envoy resources are defined and used only in the context of a specific proxy. In this document, the context is the `Dataplane`, but the format is intentionally more general and not strictly tied to that resource type. These resources may be referenced in policies, metrics, and other tooling, so a naming format is needed that reflects their scoped, contextual nature without pretending they are globally distinct Kuma resources.

To solve this, we define a contextual format that uses a `<keyword>_` prefix followed by a descriptor. The descriptor identifies the relevant section of the current proxy context the resource belongs to.

#### Format: `<keyword>_<descriptor>`

* Begins with a keyword that indicates the resource is local to the current proxy context
* Followed by a descriptor string that identifies the specific part of the proxy context

The descriptor must use only characters allowed in URL queries, as defined in [MADR-070](070-resource-identifier.md#url-query):

* Lowercase letters (`a` to `z`)
* Numbers (`0` to `9`)
* Dashes (`-`)
* Underscores (`_`)
* Dots (`.`) for future use cases like domain-style or MeshPassthrough naming

#### Keyword options

##### Option 1: Use `self` as the keyword

**Benefits:**

* Clearly indicates the resource belongs to the current proxy context
* Familiar and widely used in tools and programming languages to refer to the current object or scope

##### Option 2: Use `this` as the keyword

**Benefits:**

* Clearly signals that the resource is tied to this specific proxy context
* Common in many programming languages for referencing current object or scope

**Drawbacks:**

* Might be confused with general-purpose language in documentation or examples
* Less commonly used in naming schemes compared to `self`

##### Option 3: Use `local` as the keyword

**Benefits:**

* Suggests the resource is local to the current proxy
* Has precedent in networking terminology

**Drawbacks:**

* Less familiar than `self` or `this` in naming formats
* May be confused with other networking terms like `localhost`, implying different meaning (e.g. local to host or machine rather than the proxy)

##### Chosen keyword: `self`

We choose `self` as the keyword in the contextual naming format `<keyword>_<descriptor>`. This makes the final format `self_<descriptor>`.

### Format definition: `sectionName`

The `sectionName` identifies a specific unit within the current context (in this document, the `Dataplane`) that can be referred to by Kuma policies. Currently, this applies to ports (such as outbounds and non-system inbounds), but the format is designed to support future extensions, such as DNS-like labels for `MeshPassthrough`.

This format is the same as the `sectionName` used in the KRI naming scheme and is used in contextual naming to label resources tied to a particular part of the proxy's configuration.

**Requirements:**

* Must be **1 to 63 characters** long
* Can contain:
  * Lowercase letters (`a`–`z`)
  * Digits (`0`–`9`)
  * Hyphens (`-`)
  * Dots (`.`)
* Must **start with a letter or digit**
* Must **end with a letter or digit**
* Must **not contain** consecutive hyphens (`--`)
* Must **not contain** consecutive dots (`..`)
* Must **not start or end** with a hyphen or dot

These rules combine:

* [Kubernetes Service port name requirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#serviceport-v1-core) (based on **DNS_LABEL** from [RFC1123 section 2.1](https://www.rfc-editor.org/rfc/rfc1123#section-2.1))
* [Kubernetes Container port name requirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#containerport-v1-core) (based on **IANA_SVC_NAME** from [RFC6335 section 5.1](https://www.rfc-editor.org/rfc/rfc6335.html#section-5.1))
* Added support for `.` to allow future formats (e.g. for `MeshPassthrough`)

**Regular expression:**

```
(([1-9][0-9]{0,4})|([a-z0-9](?!.*--)(?!.*\.\.)[a-z0-9.-]{0,61}[a-z0-9]))
```

### Naming for non-system inbound-related resources

The current KRI naming format, proposes that non-system inbound resources (such as listeners and clusters) use the full KRI name of the originating `Dataplane` plus the inbound port or name as the `sectionName`. This would result in names like:

```
kri_dp_default_kuma-2_kuma-demo_demo-app-ddd8546d5-vg5ql_5050
```

While technically correct and traceable, this significantly increases metrics cardinality compared to current formats, such as `localhost_<port>` for clusters and `<address>_<port>` for listeners, which use Envoy’s default format. This would lead to regressions in performance, cost, and compatibility with observability tooling. The following alternatives are considered:

#### Option 1: Use the contextual format defined for proxy-local resources

This option proposes that non-system inbounds use the [contextual format](#context-and-problem-statement) with the `inbound_<sectionName>` descriptor:

```
self_inbound_<sectionName>
```

The `self` keyword refers to the current proxy context. The `<sectionName>` follows the format defined in [section name](#format-sectionname): it uses the port name if present, otherwise the port number.

**Benefits:**

* Avoids metric cardinality increase by skipping global identifiers like mesh, zone, or proxy name
* Keeps names short and focused on local context
* Prevents metric churn in Kubernetes environments
* Clearly separates inbounds from KRI or system resources
* Aligns with how policies reference inbounds via `sectionName`

**Drawbacks:**

* No direct reference to the originating `Dataplane`
* Requires tooling to recognize the contextual format category

#### Option 2: Align resource names with existing `localhost_<port>` inbound clusters stat format

This option proposes using the already established stat format `localhost_<port>` (currently used for inbound Envoy cluster stats) as the unified name for all related xDS resources such as clusters and listeners.

Specifically:

* Change cluster names from `localhost:<port>` to `localhost_<sectionName>` to match stat format
* Change listener and other inbound-related resource names (e.g. from `inbound:10.42.0.83:5050`) to `localhost_<sectionName>` for consistency across resources

Unlike current formats, this allows `<sectionName>` to be either a port number or a named port (e.g. `httpport`).

**Examples:**

```
localhost_5050
localhost_httpport
```

**Benefits:**

* No increase in metric cardinality
* Keeps resource names in sync with existing `localhost_<port>` stat format for clusters, as long as no port name is used
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

#### Option 4: Treat inbounds as "system" resources and use `system_<prefix>_<sectionName>` format

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

#### Decision: Naming for non-system inbound-related resources

Non-system inbound-related Envoy resources and stats will use the [`self_inbound_<sectionName>`](#format-self_descriptor) format. The `self` keyword marks the resource as scoped to the current proxy context. The `inbound_<sectionName>` descriptor uses the port name if defined, otherwise the port number, following the [section name format](#format-sectionname).

Examples:

```
self_inbound_httpport  
self_inbound_5050
```

This format avoids high cardinality from full KRI names, ensures clarity by keeping the naming local to the proxy, and aligns with how policies reference inbounds using `sectionName`. It also keeps the naming distinct from `kri_` and `system_` resources, reducing tooling complexity.

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

#### Option 1: Use system resource naming format

Treat passthrough resources as internal system components and name them using the `system_<..>` format described in [MADR-076 Standardized Naming for internal xDS Resources](076-naming-internal-envoy-resources.md). This approach follows the conventions used for system-generated resources not tied to specific Kuma resources.

**Benefits:**

* Clearly separates passthrough resources from user-defined service-related resources
* Matches the assumption that these resources may be internal to proxy behavior in some setups
* Avoids expanding contextual naming to cases where observability value may vary

**Drawbacks:**

* These resources may be part of the regular service-to-service traffic path, where treating them as system components would be misleading and using the system format would not accurately reflect their role
* Using the system format may reduce visibility into traffic users expect to observe
* Makes it harder to trace and correlate passthrough traffic with `Dataplane` configuration and metrics when relevant

#### Option 2: Use contextual `self_<descriptor>` format

Apply the same contextual naming structure used for inbounds to passthrough resources, using the `self_<descriptor>` format. Treat passthrough resources as part of the scoped configuration defined within the `Dataplane`.

**Benefits:**

* Reflects that passthrough traffic is explicitly configured in the `Dataplane`
* Reuses the same naming logic as other contextual resources, keeping naming consistent
* Helps users trace and monitor passthrough traffic with familiar patterns

**Drawbacks:**

* Extends use of the contextual format to resources that are not bound to a specific port or service
* Requires tooling to recognize and handle these new descriptors correctly
* Introduces new descriptor values for passthrough traffic that must be formally defined and validated

#### Decision: Naming for transparent proxy passthrough resources

Transparent proxy passthrough-related Envoy resources and stats will use the contextual format `self_<descriptor>`, where the `<descriptor>` is:

```
transparentproxy_passthrough_<direction>_<port>_ipv<IPVersion>
```

* `<direction>` is `inbound` or `outbound`
* `<port>` is the port number from the `Dataplane`'s `transparentProxying` config
* `<IPVersion>` is `4` or `6`

We don’t start the descriptor with `<direction>` here because the config source is already under `networking.transparentProxying.<direction>` in the `Dataplane`. Instead, we use a fixed prefix and then the direction. This is different from non-system inbounds, where the descriptor starts with `inbound` because their config comes from `networking.inbound`.

This format is for passthrough traffic set up in the current `Dataplane`. It’s not always used, but in setups like `MeshPassthrough`, these resources are important and on the main traffic path. The consistent format helps make them easier to understand and observe.

### Enabling the feature per proxy

This decision introduces a breaking change to resource and stat naming. To support a safe and gradual migration, the feature will be controlled by a data plane feature flag. This flag is enabled by setting the appropriate environment variable in the context where `kuma-dp` runs.

In **Universal mode**, this is straightforward. Users are responsible for running `kuma-dp` directly, so setting the environment variable is straightforward and fully under user control.

In **Kubernetes**, the situation is more complex. The `kuma-sidecar` container, which runs `kuma-dp`, is injected by the control plane. Users cannot directly configure its environment. We considered two options for enabling this feature per workload:

#### Option 1: Dedicated annotation per workload

Users add a dedicated annotation to each pod where the feature should be enabled. The sidecar injector detects this annotation and injects the corresponding environment variable into the `kuma-sidecar` container.

**Benefits:**

* Familiar and intuitive UX for Kubernetes users
* Aligns with common patterns used by other tools
* Keeps control of the feature close to the workload definition

**Drawbacks:**

* Requires adding new annotation logic in the injector
* The annotation would become obsolete once the feature becomes the only supported naming scheme and can no longer be disabled
* No existing feature flags are currently exposed this way

#### Option 2: ContainerPatch-based opt-in

Users apply a `ContainerPatch` in each zone where they want to enable per-proxy opt-in. The patch injects the required environment variable into the `kuma-sidecar` container. Once applied, the feature can be turned on per workload using the standard container patch annotation.

**Benefits:**

* Avoids introducing and maintaining a temporary annotation
* Uses existing, well-defined `ContainerPatch` mechanism
* Matches the idea that per-proxy opt-in is a transitional state

**Drawbacks:**

* Users must create and maintain additional resources to enable the feature
* Slightly more complex and less discoverable for users unfamiliar with container patches

#### Decision: Enabling the feature per proxy

To enable the new naming format on a per-proxy basis, we will use [**Option 2: ContainerPatch-based opt-in**](#option-2-containerpatch-based-opt-in). This avoids introducing a dedicated annotation that would become obsolete once the feature becomes the only supported naming scheme.

Users can enable the feature for specific workloads by applying a `ContainerPatch` resource per zone. Then, they can set the annotation `kuma.io/container-patches: <name>` on selected workloads. This provides a clear migration path without requiring permanent API additions.

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

## Implications for Kong Mesh

The changes introduced by this document, along with those defined in [MADR-076 Standardized Naming for internal xDS Resources](076-naming-internal-envoy-resources.md), impact the Envoy resources generated by the `MeshGlobalRateLimit` policy. This policy currently adds a cluster with a static name `meshglobalratelimit:service`, which is then referenced in route-level [rate limit configuration](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-rate-limits). As a result of these changes, the cluster name will be updated to follow the standard system format, using the valid `kri` of the `MeshGlobalRateLimit` policy that is the source of this cluster:

```
system_kri_mgrl___<namespace>_<policyName>_
```

Example:

```
system_kri_mgrl___kong-mesh-system_mesh-rate-limit_
```

The `MeshOPA` policy modifies existing non-system inbound listeners but does not rename any existing Envoy resources or create new ones. Therefore, it is not affected by the changes described in this document.

Other Kong Mesh–specific features or policies do not rely on or modify Envoy resource names or stat prefixes that fall within the scope of this document.

## Implementation details

### Data plane feature flag and propagation

A new data plane feature flag will control the use of consistent naming for proxy resources and stats, including both KRI-based formats and the new `self_` format for inbounds. It will be optional at first, with a deprecation path: the new naming will become the default in the future, and the flag will eventually be removed. This gives users time to migrate without breaking existing setups.

The flag will be passed from the data plane proxy to the control plane via [xDS metadata](https://github.com/kumahq/kuma/blob/c61baf6110caa40eb3b69f7f635e9506389a7455/app/kuma-dp/cmd/run.go#L172) using:

```go
const FeatureUnifiedProxyResourcesAndStatsNaming string = "feature-unified-resource-naming"
```

#### Per-proxy opt-in

Users can enable the feature for individual data plane proxies:

| Mode       | How to enable                                                                  |
|------------|--------------------------------------------------------------------------------|
| Universal  | Set env var: `KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED=true`     |
| Kubernetes | Apply `ContainerPatch` and set annotation: `kuma.io/container-patches: <name>` |

In Kubernetes, setting the environment variable directly on the workload container does not affect the `kuma-sidecar` container, which runs `kuma-dp`. Instead, users must apply a `ContainerPatch` to inject the environment variable into the sidecar, and enable the patch with a pod annotation. This approach avoids introducing a dedicated annotation and supports gradual migration.

Example `ContainerPatch`:

```yaml
apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  name: enable-feature-unified-resource-naming
  namespace: kuma-system
spec:
  sidecarPatch:
  - op: add
    path: /env/-
    value: '{
      "name": "KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED",
      "value": "true"
    }'
```

#### Zone-wide opt-in for Kubernetes data planes

To enable the naming changes for all injected data plane proxies in a Kubernetes zone, users can configure the control plane with:

```go
UnifiedResourceNamingEnabled bool `json:"unifiedResourceNamingEnabled" envconfig:"KUMA_RUNTIME_KUBERNETES_INJECTOR_UNIFIED_RESOURCE_NAMING_ENABLED"`
```

When this setting is enabled, the sidecar injector automatically adds the environment variable to the injected `kuma-sidecar` containers. It does not rely on annotations or `ContainerPatch` resources.

#### ZoneIngress and ZoneEgress

`ZoneIngress` and `ZoneEgress` proxies must also include the feature flag to ensure consistent naming across the mesh. In both Universal and Kubernetes, this is done by setting the following environment variable in their deployments:

```env
KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED=true
```

#### Helper setting for Kubernetes installations

To simplify configuration, a new setting will be introduced under the `dataPlane.features` section for Helm and other Kubernetes-based install methods:

```yaml
dataPlane:
  features:
    unifiedResourceNaming: true
```

When enabled, it will:

* Add `KUMA_DATAPLANE_RUNTIME_UNIFIED_RESOURCE_NAMING_ENABLED=true` to all `ZoneIngress` and `ZoneEgress` deployments
* Set `KUMA_RUNTIME_KUBERNETES_INJECTOR_UNIFIED_RESOURCE_NAMING_ENABLED=true` in the control plane deployment

When the new naming becomes the default, this setting will also default to `true`. Eventually, it will be removed when the feature can no longer be disabled.

### Updating `ZoneIngressInsight` and `ZoneEgressInsight` with feature flags

To support the Kuma GUI in adapting to the updated naming formats, `ZoneIngressInsight` and `ZoneEgressInsight` must expose the same feature flag metadata already present in `DataplaneInsight`. These insight resources are exposed through the control plane API and provide metadata and status about each proxy.

We will extend both `ZoneIngressInsight` and `ZoneEgressInsight` to include active feature flags in their metadata section. This enables the GUI to detect if the `feature-unified-resource-naming` flag is active for each proxy and adjust its behavior accordingly, including how it parses and renders resource and stat names.

This work is tracked in [kumahq/kuma#13788](https://github.com/kumahq/kuma/issues/13788).

## Migration

### Kuma GUI

The Kuma GUI relies on Envoy xDS resource names and stat names to display inbound and outbound traffic details for `Dataplane`, `ZoneIngress`, and `ZoneEgress` proxies. These names are used to associate metrics with specific resources and visualize traffic paths in the interface.

To support the updated naming formats introduced in this decision, including both KRI-based and contextual `self_` names, the GUI parsers must be updated. Specifically, when the `feature-unified-resource-naming` flag is present in the metadata of `DataplaneInsight`, `ZoneIngressInsight`, or `ZoneEgressInsight`, the GUI must:

* Parse xDS resource names and stat names using:
   * The KRI format for resources tied to distinct Kuma objects (as defined in [MADR-070 Resource Identifier](070-resource-identifier.md))
   * The `self_<descriptor>` format for contextual `Dataplane`-scoped resources such as inbounds and transparent proxy passthrough
* Drop legacy assumptions like the presence of `inbound:` or `outbound:` prefixes
* Match stat names directly to xDS resource names, which now have a 1:1 correspondence
* Use the new Inspect API (described in [MADR-075](075-inspect-api-redesign.md)) to fetch resource and stat names, instead of parsing raw metrics or xDS config

### Examples

**KRI-formatted resource names:**

```
kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport
kri_mzsvc_mesh-1__kuma-system_backend-app_8080
kri_mhttpr_mesh-1_us-east-2_kuma-demo_route-1_
kri_extsvc_mesh-1__kuma-system_es1_
```

**Contextual `self_` resource names:**

```
self_8080
self_http
self_passthrough-ipv4-inbound
self_passthrough-ipv6-outbound
```

**Corresponding stat names:**

```
cluster.self_8080.upstream_cx_active: 0
cluster.self_passthrough-ipv4-inbound.upstream_cx_active: 0
cluster.kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport.upstream_cx_active: 0

http.self_8080.downstream_cx_active: 0
http.kri_mzsvc_mesh-1__kuma-system_backend-app_8080.downstream_cx_active: 0

listener.self_8080.downstream_cx_active: 0
listener.kri_msvc_mesh-1_us-east-2_kuma-demo_backend_httpport.downstream_cx_active: 0
```
