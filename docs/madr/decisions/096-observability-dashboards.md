# Observability dashboards redesign

* Status: accepted

Technical Story: https://github.com/Kong/kong-mesh/issues/9162

## Context and problem statement

Kuma currently ships observability dashboards (Grafana) via `kumactl install observability`,
which bundles a full Prometheus + Grafana stack. This command is being removed in Kuma 3.0.

We need to:
1. Decide how to deliver dashboards without `kumactl install observability`
2. Design what the new dashboards should include

Current pain points:
- `kumactl install observability` installs an opinionated, hard-to-customize stack
- Users with existing Grafana/Prometheus setups cannot easily adopt Kuma dashboards
- Dashboard content is tightly coupled to the install method
- No clear separation between "how to install" and "what to show"

Use cases:
- UC1: User has an existing Grafana instance and wants to import Kuma dashboards
- UC2: User wants to deploy a fresh observability stack alongside Kuma
- UC3: User runs Kuma on Kubernetes with a Helm-based Grafana operator (e.g. grafana-operator)
- UC4: User runs Kuma in universal mode and wants dashboards without Kubernetes
- UC5: Platform team wants to manage dashboards-as-code via GitOps

## Design

### Dashboard set

Five dashboards replace the previous set. The Service Map dashboard is dropped; the maintenance burden wasn't worth it.

| Dashboard            | Focus                                                      |
|----------------------|------------------------------------------------------------|
| Control Plane        | xDS generation, store operations, API server health        |
| Service Health (RED) | Request rate, error rate, latency percentiles              |
| Workload             | Per-workload connections, resource usage, CP connectivity  |
| Multi-Zone           | Zone connectivity, KDS sync latency, version compatibility |
| mTLS & Certificates  | TLS handshakes, certificate expiry, CA health              |

### Format

1. Raw JSON — dashboards authored and shipped as plain Grafana JSON files.
2. Jsonnet/Grafonnet — dashboards templated in Jsonnet, compiled to JSON at build time.

|                                 | Raw JSON   | Jsonnet             |
|---------------------------------|------------|---------------------|
| Contributor barrier             | low        | high (new language) |
| Cross-dashboard reuse           | copy-paste | shared libs         |
| Tooling required                | none       | jsonnet binary      |
| Grafana import                  | direct     | compile step        |
| Scales well past ~10 dashboards | no         | yes                 |

Raw JSON is the right choice for the initial set of five dashboards. Revisit if the count grows past ten or bulk cross-dashboard changes become routine.

Datasource and namespace selection are handled via Grafana's native variable system (`$datasource`, `$namespace`) defined inside the JSON. No build-time preprocessing is needed.

### Shipping

Three approaches for distributing dashboards to users.

1. Release tarball — dashboard JSON files included in the GitHub release assets. Works for all users regardless of environment. Manual import required. Versioned with Kuma. (UC1, UC4)
2. Helm ConfigMaps — dashboards shipped as ConfigMaps in the Kuma Helm chart, labeled for Grafana sidecar auto-discovery. Kubernetes-only. GitOps-friendly. Requires Grafana sidecar or operator. (UC2, UC3, UC5)
3. grafana.com registry — dashboards published to the Grafana dashboard registry. Works anywhere, import by ID. Not versioned with Kuma releases — users must re-import on upgrades. (UC1, UC2)

#### Delivery method comparison

|                      | Release tarball | Helm ConfigMaps          | grafana.com |
|----------------------|-----------------|--------------------------|-------------|
| Works on universal   | yes             | no                       | yes         |
| Works on Kubernetes  | yes             | yes                      | yes         |
| Auto-provisioned     | no              | yes (via sidecar)        | no          |
| GitOps-friendly      | no              | yes                      | no          |
| Versioned with Kuma  | yes             | yes                      | no          |
| Requires extra setup | no              | Grafana sidecar/operator | no          |

### Repository structure

New `dashboards/grafana/` directory at repo root containing dashboard JSON files. Keeping files in `app/kumactl/data/` after removing the install command would be confusing.

### Label migration

Old dashboards use `kuma_io_service`/`kuma_io_services` labels. New dashboards use workload KRI labels (`kuma.workload`). 27+ query occurrences in the service dashboards alone make patching impractical. Rebuilding from scratch.

### Known gap: multi-zone labels

KDS metrics currently lack a `zone_name` label. The Multi-Zone dashboard requires per-zone breakdown, so `pkg/kds/status/status_tracker.go` needs a code change before that dashboard can ship.

## Implications for Kong Mesh

The same dashboard set applies to Kong Mesh without modifications. Enterprise-specific features don't add dashboard requirements:

- **Multitenancy** — internal implementation detail, not exposed in default dashboards
- **License** — not tracked via metrics, no dashboard panel needed
- **OPA** — being removed
- **FIPS** — build-time config, not a runtime metric

## Decision

Rebuild Grafana dashboards authored as raw JSON. Dashboards are stored under `dashboards/grafana/` at the repository root and included in the release tarball at `kuma-{VERSION}/dashboards/grafana/*.json`. Users download the tarball and import the JSON files manually into their existing Grafana instance.

## Notes <!-- optional -->

