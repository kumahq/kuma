# Example suite - MOTB core (directory format)

Reference example showing the directory-format suite produced by kuma-suite-author v2.

## Directory layout

```
suites/motb-core/
├── suite.md
├── baseline/
│   ├── namespace.yaml
│   ├── otel-collector.yaml
│   └── demo-workload.yaml
└── groups/
    ├── g01-crud.md
    ├── g02-validation.md
    ├── g03-backendref.md
    ├── g04-xds.md
    ├── g05-signal-flow.md
    └── g06-http-protocol.md
```

## suite.md content

```markdown
# MOTB core - manual test suite

## Suite metadata

- suite id: motb-core
- feature scope: MeshMetric, MeshTrace, MeshAccessLog (unified observability)
- target environments: single-zone, multi-zone
- kuma-suite-author version: 2.0.0

## Baseline manifests

| File                          | Purpose                                           |
| ----------------------------- | ------------------------------------------------- |
| baseline/namespace.yaml       | Create kuma-demo namespace with sidecar injection |
| baseline/otel-collector.yaml  | Deploy OpenTelemetry collector backend            |
| baseline/demo-workload.yaml   | Deploy demo-app client and server pods            |

## Group structure

| Group | File               | Goal                                | Minimum artifacts                           |
| ----- | ------------------ | ----------------------------------- | ------------------------------------------- |
| G1    | groups/g01-crud.md | Resource CRUD                       | create/get/update/delete outputs            |
| G2    | groups/g02-validation.md | Validation rejects invalid specs | admission errors for invalid inputs         |
| G3    | groups/g03-backendref.md | backendRef acceptance and mutual exclusion | accepted and rejected applies     |
| G4    | groups/g04-xds.md  | xDS correctness                     | config dump snippets                        |
| G5    | groups/g05-signal-flow.md | Signal flow (metrics/traces/logs) | collector logs with all signal types      |
| G6    | groups/g06-http-protocol.md | HTTP protocol behavior          | no forced HTTP/2, URI path artifacts        |

## Execution contract

- All manifests applied through `"$SKILL_DIR/scripts/apply-tracked-manifest.sh"`
- All failures trigger immediate triage before next group
- All pass/fail decisions include artifact pointers
```

## Group file content (g01-crud.md)

```markdown
# G1 - Resource CRUD

## Prerequisites

- Baseline manifests applied
- demo-app pods Running/Ready

## Steps

### S1.1 - Create MeshMetric

- manifest: inline MeshMetric targeting demo-app
- command: apply through tracked script
- expected: resource accepted, status shows `Accepted: true`
- artifacts: apply output, resource YAML

### S1.2 - Get and inspect

- command: `kubectl get meshmetric -n kuma-demo -o yaml`
- expected: resource present with correct spec
- artifacts: get output

### S1.3 - Update spec

- manifest: modified MeshMetric with changed backend
- command: apply through tracked script
- expected: resource updated, generation incremented
- artifacts: apply output, diff from previous

### S1.4 - Delete

- command: `kubectl delete meshmetric -n kuma-demo <name>`
- expected: resource removed
- artifacts: delete output, get confirming 404
```

## Baseline manifest (namespace.yaml)

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: kuma-demo
  labels:
    kuma.io/sidecar-injection: enabled
```

## How the runner consumes this

1. Phase 1 reads `suite.md` (~40 lines of metadata, tables, contract).
2. Phase 3 baseline applies `namespace.yaml`, `otel-collector.yaml`, `demo-workload.yaml`.
3. Phase 3 G1 reads `groups/g01-crud.md` (~30 lines), executes steps, drops from context.
4. Phase 3 G2 reads `groups/g02-validation.md`, and so on per group.

Peak context per group: ~140 lines (suite.md + one group file) instead of 1300+ for a monolithic suite.
