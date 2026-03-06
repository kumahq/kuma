# Contents

1. [Overview](#overview)
2. [Signal catalog](#signal-catalog)
3. [Signal strength](#signal-strength)
4. [Group mapping](#group-mapping)
5. [Worked example](#worked-example)

---

# Variant detection

How to identify feature variants that expand a test suite beyond the standard G1-G7 groups.

## Overview

A **variant** is a distinct mode, configuration, or topology that changes runtime behavior enough to need its own test groups. A single Kuma feature can have many variants - the MOTB example expands from 7 base groups to 53 total because of deployment topologies, backend types, naming modes, and protocol paths.

Variants are not edge cases (those go in G5). A variant produces different xDS output, uses different code paths in `plugin.go`, or requires a different cluster topology to test.

## Signal catalog

Scan source code for these patterns. Each signal type maps to a specific kind of variant.

### S1 Deployment topology

**Where to look**: resource type registration, KDS markers, `pkg/kds/` sync config.

**What it means**: the feature behaves differently in multi-zone deployments. Resources sync from global to zone CPs, zone-local resources may conflict, universal and Kubernetes zones may diverge.

**Produces**: multi-zone test groups covering global-to-zone sync, zone-local override, universal zone parity.

### S2 Feature modes

**Where to look**: enum or string fields in the API spec struct (e.g., `NamingMode`, `Mode`), switch/case blocks in `plugin.go`.

**What it means**: the feature has named modes that change its output. Each mode generates different xDS configuration.

**Produces**: one group per mode covering mode-specific xDS output and signal flow.

### S3 Backend/endpoint variants

**Where to look**: multiple backend types in the spec struct (e.g., `OtelBackend`, `PrometheusBackend`), `backendRef` with different target kinds, endpoint configuration alternatives.

**What it means**: the feature supports multiple backend types with different connection semantics, protocol handling, or xDS cluster types.

**Produces**: per-backend groups covering connection setup, signal flow, and protocol behavior.

### S4 Feature flags

**Where to look**: conditional branches in `Apply()` that check feature flags, environment variables, or control plane config.

**What it means**: the feature has opt-in/opt-out paths that produce different xDS output.

**Produces**: per-flag groups covering both enabled and disabled paths.

### S5 Policy role variants

**Where to look**: `targetRef` section of the API spec, `+kuma:policy:is_target_ref` markers, producer/consumer/workload-owner annotations.

**What it means**: the policy behaves differently depending on whether it targets the producer, consumer, or workload owner side of a connection.

**Produces**: role-specific groups covering targeting behavior, precedence, and conflict resolution.

### S6 Protocol variants

**Where to look**: HTTP/TCP/gRPC branching in `plugin.go`, protocol-specific listener/cluster generation, `MatchedPolicies` filtering by protocol.

**What it means**: the feature generates different xDS for different protocols. HTTP may get route-level config while TCP gets listener-level.

**Produces**: per-protocol groups covering protocol-specific xDS and traffic behavior.

### S7 Backward compatibility paths

**Where to look**: `deprecated.go`, old field names in the API spec, migration logic in `plugin.go`, version-gated behavior.

**What it means**: the feature has legacy code paths that must keep working during migration. New and old configs may coexist.

**Produces**: legacy path groups covering deprecated field behavior, migration warnings, and old-to-new transition.

## Signal strength

Not all signals warrant variant groups. Classify each detected signal:

| Strength | Criteria                                                                            | Action                                           |
| :------- | :---------------------------------------------------------------------------------- | :----------------------------------------------- |
| Strong   | Distinct code path in plugin.go, different xDS output, separate test golden files   | Include in variant list, recommend selecting     |
| Moderate | Mentioned in spec but no separate code path yet, or code path exists but is trivial | Present with `[uncertain]` tag, explain evidence |
| Weak     | Hinted at in comments or docs but no implementation                                 | Mention in notes, don't pre-select               |

Strong signals: the code does something measurably different. Moderate signals: the code might do something different but runtime testing has not confirmed the behavior. Weak signals: a pattern that could become a variant but is not one yet.

## Group mapping

Variant groups start at G8 and number sequentially. When a variant produces multiple related groups, use range notation.

Rules:

- One variant can produce 1-N groups depending on complexity
- Range notation: `G17-G26 Pipe mode pre-unified` means groups 17 through 26 all belong to the "pipe mode pre-unified" variant
- Each variant group follows the same format as G1-G7 (manifests, commands, expected outcomes, artifacts)
- Document which signal produced each variant group

## Worked example

MOTB (MeshObservabilityTelemetryBackend) breakdown:

**Base groups (G1-G7):**

- G1: CRUD for the MOTB resource
- G2: Validation - invalid specs, missing fields
- G3: backendRef acceptance and mutual exclusion
- G4: xDS correctness for metrics/traces/logs clusters and listeners
- G5: Signal flow - data reaches collectors
- G6: HTTP protocol behavior (no forced HTTP/2)
- G7: Dangling reference handling

**Detected variants and resulting groups:**

| Signal | Variant                                 | Groups  | Count |
| :----- | :-------------------------------------- | :------ | :---- |
| S7     | Backward compat (inline endpoint)       | G8      | 1     |
| S1     | KDS sync in multi-zone                  | G9      | 1     |
| S3     | Mixed backend usage (OTel + Prometheus) | G10     | 1     |
| S3     | Path suffix semantics                   | G11     | 1     |
| S2     | Unified naming mode                     | G12     | 1     |
| S5     | Gap analysis and edge semantics         | G13     | 1     |
| S3     | Endpoint optionality and schema parity  | G14     | 1     |
| S1     | Mesh isolation (cross-mesh)             | G15     | 1     |
| S3     | nodeEndpoint behavior                   | G16     | 1     |
| S6+S4  | Pipe mode pre-unified (per-signal)      | G17-G26 | 10    |
| S1     | Universal multi-zone parity             | G27-G39 | 13    |
| S2+S6  | Unified pipe mode                       | G40-G53 | 14    |

**Total**: 7 base + 46 variant = 53 groups.

The variant signals came from: 3 backend types (S3), 2 naming modes (S2), KDS markers (S1), protocol branching for pipe vs network (S6), deprecated inline endpoint (S7), and feature flag for unified pipe mode (S4).
