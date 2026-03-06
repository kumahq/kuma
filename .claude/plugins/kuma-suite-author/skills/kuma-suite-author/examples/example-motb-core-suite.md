# Example suite - MOTB core (directory format)

Reference example showing the directory-based suite layout for MOTB testing.

## Directory tree

```
motb-core/
  suite.md
  baseline/
    namespace.yaml
    otel-collector.yaml
    demo-workload.yaml
  groups/
    g01-crud.md
    g02-validation.md
    g03-backendref.md
    g04-xds.md
    g05-signal-flow.md
    g06-http-protocol.md
    g07-dangling-ref.md
    g08-backward-compat.md
    g09-kds-sync.md
    g10-mixed-backend.md
    g11-path-suffix.md
    g12-unified-naming.md
    g13-gap-analysis.md
    g14-endpoint-optionality.md
    g15-mesh-isolation.md
    g16-node-endpoint.md
    g17-g26-pipe-mode.md
    g27-g39-universal-multizone.md
    g40-g53-unified-pipe-mode.md
```

## Example suite.md

```markdown
# MOTB core manual tests

## Contents

1. [Suite metadata](#suite-metadata)
2. [Baseline manifests](#baseline-manifests)
3. [Test groups](#test-groups)
4. [Execution contract](#execution-contract)
5. [Failure triage](#failure-triage)

## Suite metadata

- suite id: motb-core
- feature scope: MeshMetric, MeshTrace, MeshAccessLog observability backends
- target environments: single-zone, multi-zone, universal
- required dependencies: otel-collector, demo-workload, prometheus (for G10)
- skipped groups: none

## Baseline manifests

| File                         | Purpose                                     |
| :--------------------------- | :------------------------------------------ |
| baseline/namespace.yaml      | kuma-demo namespace with kuma.io/sidecar-injection |
| baseline/otel-collector.yaml | OTel collector deployment in kuma-system     |
| baseline/demo-workload.yaml  | echo-server + demo-client pods              |

## Test groups

| Group   | File                              | Goal                                  | Minimum artifacts                          |
| :------ | :-------------------------------- | :------------------------------------ | :----------------------------------------- |
| G1      | groups/g01-crud.md                | Resource CRUD                         | create/get/update/delete outputs + YAML    |
| G2      | groups/g02-validation.md          | Validation rejects invalid specs      | admission errors for invalid inputs        |
| G3      | groups/g03-backendref.md          | backendRef acceptance + exclusion     | accepted and rejected applies              |
| G4      | groups/g04-xds.md                 | xDS correctness                       | config dump snippets                       |
| G5      | groups/g05-signal-flow.md         | Signal flow (metrics/traces/logs)     | collector logs with all signal types       |
| G6      | groups/g06-http-protocol.md       | HTTP protocol behavior                | URI path artifacts                         |
| G7      | groups/g07-dangling-ref.md        | Dangling reference behavior           | no crash, info log artifacts               |
| G8      | groups/g08-backward-compat.md     | Backward compat (inline endpoint)     | deprecation warning + runtime config       |
| G9      | groups/g09-kds-sync.md            | KDS sync in multi-zone                | global to zone presence artifacts          |
| G10     | groups/g10-mixed-backend.md       | Mixed backend usage                   | OTel and Prometheus artifacts              |
| G11     | groups/g11-path-suffix.md         | Path suffix semantics                 | URI with/without base path                 |
| G12     | groups/g12-unified-naming.md      | Unified naming mode                   | listener/cluster naming artifacts          |
| G13     | groups/g13-gap-analysis.md        | Gap analysis and edge semantics       | expected limitations confirmed             |
| G14     | groups/g14-endpoint-optionality.md | Endpoint optionality + schema parity | backendRef-only acceptance artifacts       |
| G15     | groups/g15-mesh-isolation.md      | Mesh isolation                        | cross-mesh dangling behavior               |
| G16     | groups/g16-node-endpoint.md       | nodeEndpoint behavior                 | HOST_IP + STATIC cluster artifacts         |
| G17-G26 | groups/g17-g26-pipe-mode.md       | Pipe mode pre-unified                 | per-signal sockets, dynconf, E2E           |
| G27-G39 | groups/g27-g39-universal-multizone.md | Universal multi-zone              | k8s and universal zone parity              |
| G40-G53 | groups/g40-g53-unified-pipe-mode.md   | Unified pipe mode                 | shared socket, /otel route, opt-out        |

## Execution contract

- all manifests applied through `"$SKILL_DIR/scripts/apply-tracked-manifest.sh"`
- all failures trigger immediate triage before next group
- all pass/fail decisions include artifact pointers
- edge cases from `references/mesh-policies.md` included

## Failure triage

See `references/agent-contract.md` (failure policy and bug triage protocol) for the full procedure.
```

## Example group file (g01-crud.md)

```markdown
# G1 - Resource CRUD

## Steps

### 1. Create MeshMetric

Apply the resource:

~~~yaml
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: basic-metrics
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: Mesh
  default:
    backends:
      - type: OpenTelemetry
        openTelemetry:
          endpoint: otel-collector.kuma-system:4317
~~~

Verify:

~~~bash
kubectl get meshmetric basic-metrics -n kuma-system -o yaml
~~~

Expected: resource created with status conditions populated.

### 2. Update MeshMetric

Change the endpoint port to 4318 and reapply. Verify the resource reflects the update.

### 3. Delete MeshMetric

~~~bash
kubectl delete meshmetric basic-metrics -n kuma-system
~~~

Verify the resource is gone and no orphaned xDS config remains.

## Artifacts

- resource YAML before/after update
- kubectl output for create/get/update/delete
```

## Example baseline manifest (namespace.yaml)

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: kuma-demo
  labels:
    kuma.io/sidecar-injection: enabled
```
