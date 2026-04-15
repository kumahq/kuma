# MeshPassthrough: Migrate filter chain selection to Envoy Matching API

* Status: accepted

Technical Story: <!-- TODO: link to github issue -->

## Context and Problem Statement

MeshPassthrough generates Envoy filter chains with `FilterChainMatch` fields
(transport protocol, application protocols, server names, destination port,
prefix ranges). This is Envoy's legacy linear-list matching approach — Envoy
evaluates each filter chain match sequentially until one matches.

The [Envoy Matching API](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching_api)
provides a tree-structured alternative via `Listener.filter_chain_matcher`
(`xds.type.matcher.v3.Matcher`). It offers:

- **Sublinear matching** — exact match maps (O(1) HashMap), prefix match maps
  (Trie), IP range matchers (Level-Compressed trie) instead of linear scan
- **Composable tree structure** — hierarchical decisions (transport protocol →
  SNI → port → IP) instead of flat list with implicit ordering
- **Eliminates ordering sensitivity** — the current implementation has complex
  ordering logic (`order.go`, 284 lines) to ensure Envoy picks the right filter
  chain. The matcher tree encodes precedence structurally

### Current pain points

1. **Ordering complexity** — `orderMatchers()` sorts by protocol, match type,
   port, domain length, CIDR prefix length, IP value, and route count. Bugs in
   ordering cause silent misrouting
2. **Port duplication** — rules matching port 0 (all ports) must be duplicated
   for every specific port in the config to avoid Envoy selecting the wrong
   filter chain
3. **Linear scan at runtime** — with many passthrough entries, Envoy evaluates
   filter chains one by one on every new connection
4. **FilterChainMatch limitations** — cannot express complex conditions like
   "match this IP range AND this port AND this protocol" without creating
   separate filter chains for each combination

### Scope

This MADR covers MeshPassthrough only. Other policies that generate filter
chains (e.g. transparent proxy generator) are out of scope.

## Design

### How the Envoy Matching API works for filter chain selection

1. `Listener.FilterChainMatcher` is set to a `xds.type.matcher.v3.Matcher`
2. Each filter chain has a unique `name` and **no** `FilterChainMatch`
3. The matcher tree evaluates network inputs and produces an `OnMatch` action
   containing `google.protobuf.StringValue` with the filter chain name
4. Filter chains, filters, and clusters remain unchanged — only selection changes

#### Available network inputs

All types exist in go-control-plane v1.36.0 (already a dependency) and are
registered in `pkg/xds/envoy/imports.go`.

| Input | Go type | Envoy name |
|-------|---------|------------|
| Destination IP | `network.v3.DestinationIPInput` | `envoy.matching.inputs.destination_ip` |
| Destination Port | `network.v3.DestinationPortInput` | `envoy.matching.inputs.destination_port` |
| Server Name (SNI) | `network.v3.ServerNameInput` | `envoy.matching.inputs.server_name` |
| Transport Protocol | `network.v3.TransportProtocolInput` | `envoy.matching.inputs.transport_protocol` |
| Application Protocol | `network.v3.ApplicationProtocolInput` | `envoy.matching.inputs.application_protocol` |

Package: `envoy/extensions/matching/common_inputs/network/v3`

#### Available matchers

| Matcher | Go type | Purpose |
|---------|---------|---------|
| IP/CIDR range | `ip.v3.Ip` with `cidr_ranges` | Matches destination IP against CIDR list (Level-Compressed trie) |
| String | `matcher.v3.StringMatcher` | Exact, prefix, suffix, regex matching |

#### Matcher tree structure

```
Matcher
├── MatcherType (one of):
│   ├── MatcherList    — linear list with predicates (for IP/CIDR custom matchers)
│   └── MatcherTree    — sublinear tree with input + map
│       ├── Input: TypedExtensionConfig (network input)
│       └── TreeType (one of):
│           ├── ExactMatchMap   — map[string]*OnMatch (O(1) lookup)
│           ├── PrefixMatchMap  — map[string]*OnMatch (trie lookup)
│           └── CustomMatch     — TypedExtensionConfig (e.g. IP matcher)
├── OnNoMatch: *OnMatch  — fallback when nothing matches
```

`OnMatch` is either:
- `Action` → `TypedExtensionConfig` wrapping `google.protobuf.StringValue` (the filter chain name)
- `Matcher` → nested `Matcher` for hierarchical decisions

### Proposed matcher tree

```
Listener.FilterChainMatcher = MatcherTree(TransportProtocolInput)
├── exact_match_map:
│   ├── "tls" → MatcherTree(ServerNameInput)
│   │   ├── exact_match_map:
│   │   │   └── "api.example.com" → MatcherTree(DestinationPortInput)
│   │   │       ├── exact_match_map:
│   │   │       │   └── "443" → Action("meshpassthrough_tls_api.example.com_443")
│   │   │       └── on_no_match → Action("meshpassthrough_tls_api.example.com_*")
│   │   ├── prefix_match_map:  (for wildcard domains like *.example.com)
│   │   │   └── ".example.com" → ...
│   │   └── on_no_match → MatcherList [IP/CIDR TLS matchers]
│   │
│   └── "raw_buffer" → MatcherTree(ApplicationProtocolInput)
│       ├── exact_match_map:
│       │   └── "http/1.1" → MatcherTree(DestinationPortInput)
│       │       └── exact_match_map → per-port → IP/domain matchers → Action(name)
│       └── on_no_match → [TCP filter chains]
│           MatcherList with IP/CIDR predicates → Action(name)
```

Key decisions:
- **Top-level split**: transport protocol (tls vs raw_buffer) — O(1)
- **TLS branch**: ServerNameInput with exact/prefix maps for domains, nested
  DestinationPortInput for port. IP/CIDR via MatcherList with CustomMatch
- **Raw buffer branch**: ApplicationProtocolInput separates HTTP from TCP.
  HTTP uses port + VirtualHost routing (unchanged). TCP uses IP/CIDR matching
- **Filter chains keep names** but lose `FilterChainMatch`
- **IPv4/IPv6**: separate listeners, separate matcher trees (as today)
- **Port 0 (wildcard)**: `on_no_match` fallback in port trees
- **MySQL**: listener filter port exclusions are orthogonal, unchanged

### Feature flag approach

The new matcher path is gated behind a CP-level experimental flag. Old code is
preserved intact. Both code paths coexist:

**CP config** (`pkg/config/app/kuma-cp/config.go`):
```go
type ExperimentalConfig struct {
    // ...existing fields...
    // If true, MeshPassthrough uses the Envoy Matching API (filter_chain_matcher)
    // instead of FilterChainMatch for filter chain selection.
    // This provides sublinear matching performance for large passthrough configs.
    MeshPassthroughMatcherAPI bool `json:"meshPassthroughMatcherAPI" envconfig:"KUMA_EXPERIMENTAL_MESH_PASSTHROUGH_MATCHER_API"`
}
```

**Plumbing**: `ExperimentalConfig` → `ControlPlaneContext` → policy plugin → `xds.Configurer`

Since `xds_context.Context` has `ControlPlane *ControlPlaneContext`, the
cleanest path is to add the flag to `ControlPlaneContext`:

```go
// pkg/xds/context/context.go
type ControlPlaneContext struct {
    // ...existing fields...
    MeshPassthroughMatcherAPI bool
}
```

Set it during context build from `rt.Config().Experimental.MeshPassthroughMatcherAPI`.

**In the configurer**:
```go
// pkg/plugins/policies/meshpassthrough/plugin/xds/configurer.go
type Configurer struct {
    // ...existing fields...
    UseMatcherAPI bool
}

func (c Configurer) configureListener(...) error {
    // build filter chains (shared code)
    for _, matcher := range orderedFilterChainMatches {
        configurer := FilterChainConfigurer{
            // ...existing fields...
            SkipFilterChainMatch: c.UseMatcherAPI,
        }
        // ...
    }
    // new code path: build and set matcher tree
    if c.UseMatcherAPI {
        matcherTree, err := BuildFilterChainMatcher(orderedFilterChainMatches, isIPv6)
        if err != nil {
            return err
        }
        listener.FilterChainMatcher = matcherTree
    }
    // ...
}
```

**In the plugin**:
```go
// pkg/plugins/policies/meshpassthrough/plugin/v1alpha1/plugin.go
configurer := xds.Configurer{
    // ...existing fields...
    UseMatcherAPI: ctx.ControlPlane.MeshPassthroughMatcherAPI,
}
```

When the flag is off (default), behavior is identical to today. When enabled,
filter chains lose their `FilterChainMatch` and the listener gets a
`FilterChainMatcher` tree. Tests are duplicated to verify both paths.

### Existing code reuse

The `pkg/envoy/builders/xds/matchers/` package already provides:
- Generic `Builder[R]` / `Configurer[R]` pattern (`pkg/envoy/builders/common/builder.go`)
- `NewMatcherBuilder()`, `MatchersList()`, `FieldMatcher()`, `NewPredicate()`
- `NewOnMatch()`, `SkipFilterAction()`, `RbacAction()`
- `NewExtensionWithMatcher()`, `Matcher()`, `Filter()`

These are oriented toward `MatcherList` with predicates (used by
MeshFaultInjection and MeshTrafficPermission). New utilities are needed for
`MatcherTree` with exact/prefix/custom match maps.

## Implementation steps

### Step 1: Add experimental flag to CP config

**File**: `pkg/config/app/kuma-cp/config.go`
- Add `MeshPassthroughMatcherAPI bool` to `ExperimentalConfig`
- Env var: `KUMA_EXPERIMENTAL_MESH_PASSTHROUGH_MATCHER_API`
- Default: `false`

**File**: `pkg/xds/context/context.go`
- Add `MeshPassthroughMatcherAPI bool` to `ControlPlaneContext`

**File**: `pkg/xds/sync/components.go` (or wherever context is built)
- Wire `rt.Config().Experimental.MeshPassthroughMatcherAPI` into the context

### Step 2: Add MatcherTree builder utilities to shared package

**File**: `pkg/envoy/builders/xds/matchers/matchers.go`

New configurers:
- `MatcherTree(input, treeType)` — creates `Matcher` with `MatcherTree`
- `ExactMatchMap(entries map[string]*OnMatch)` — tree type
- `PrefixMatchMap(entries map[string]*OnMatch)` — tree type
- `CustomMatch(name string, config proto.Message)` — tree type (for IP matcher)

New input helpers:
- `TransportProtocolInput()`, `DestinationPortInput()`, `ServerNameInput()`,
  `DestinationIPInput()`, `ApplicationProtocolInput()`

New action:
- `FilterChainNameAction(name string)` — wraps name in `google.protobuf.StringValue`

New matcher helper:
- `IPRangeMatcher(cidrRanges []*core.CidrRange, statPrefix string)` — wraps
  `envoy.extensions.matching.input_matchers.ip.v3.Ip`

### Step 3: New filter chain matcher builder for MeshPassthrough

**New file**: `pkg/plugins/policies/meshpassthrough/plugin/xds/filter_chain_matcher.go`

- `BuildFilterChainMatcher(matches []FilterChainMatch, isIPv6 bool) (*matcher_config.Matcher, error)`
- Internal helpers for each tree layer

### Step 4: Add feature flag to configurer and filter chain builder

**File**: `pkg/plugins/policies/meshpassthrough/plugin/xds/configurer.go`

- Add `UseMatcherAPI bool` field to `Configurer`
- When true: build filter chains without `FilterChainMatch`, then build and set
  `listener.FilterChainMatcher`
- When false: existing behavior unchanged

**File**: `pkg/plugins/policies/meshpassthrough/plugin/xds/listeners_filter_chain.go`

- Add `SkipFilterChainMatch bool` to `FilterChainConfigurer`
- When true: skip `MatchTransportProtocol`, `MatchApplicationProtocols`,
  `MatchServerNames`, `MatchDestiantionAddress`, `MatchDestiantionPort` calls
- Filter chain name, network filters, route config unchanged

### Step 5: Wire feature flag from plugin

**File**: `pkg/plugins/policies/meshpassthrough/plugin/v1alpha1/plugin.go`

- Read `ctx.ControlPlane.MeshPassthroughMatcherAPI` and pass to `xds.Configurer`

### Step 6: Duplicate test cases

**File**: `pkg/plugins/policies/meshpassthrough/plugin/v1alpha1/plugin_test.go`

- Copy existing test table entries for the matcher API path
- New golden files in `testdata/` with `_matcher_api` suffix
- Same inputs, different expected output (filter chains without
  `filterChainMatch`, listener with `filterChainMatcher`)

### Step 7: Validate

```bash
make format && make check
make test TEST_PKG_LIST=./pkg/plugins/policies/meshpassthrough/...
```

## Open questions

1. **Application protocol matching**: `ApplicationProtocolInput` returns a
   single value per connection. HTTP filter chains currently match on
   `["http/1.1", "h2c"]`. Need to verify if both values need separate entries
   in the exact match map or if Envoy handles this differently
2. **Envoy version floor**: `filter_chain_matcher` is available since Envoy
   1.22+. Kuma's minimum supported Envoy version should be verified
3. **Interaction with transparent_proxy_generator**: passthrough listener is
   also configured by other generators. Need to ensure `filter_chain_matcher`
   does not conflict with filter chains added elsewhere

## Security implications and review

No security impact. Filter chain selection mechanism changes, but the actual
filters, RBAC, and mTLS configuration remain identical.

## Reliability implications

- Feature-flagged: zero risk to existing deployments
- Matcher tree is more deterministic than ordering-dependent linear matching
- Sublinear matching reduces CPU on connection setup for large passthrough configs

## Implications for Kong Mesh

None. MeshPassthrough is an upstream-only policy. The feature flag ensures no
behavior change until explicitly enabled.

## Decision

1. Migrate MeshPassthrough filter chain selection from `FilterChainMatch` to the
   Envoy Matching API (`Listener.filter_chain_matcher`)
2. Gate behind a feature flag; preserve existing code path
3. Add `MatcherTree` builder utilities to `pkg/envoy/builders/xds/matchers/`
   for reuse by other policies
4. Duplicate tests to verify both paths
