# Contents

1. [Suite naming](#suite-naming)
2. [Suite directory layout](#suite-directory-layout)
3. [suite.md structure](#suitemd-structure)
4. [Baseline directory](#baseline-directory)
5. [Groups directory](#groups-directory)
6. [Group file structure](#group-file-structure)
7. [Standard group structure](#standard-group-structure)
8. [Manifest conventions](#manifest-conventions)
9. [Validation step patterns](#validation-step-patterns)
10. [Artifact capture patterns](#artifact-capture-patterns)
11. [Execution contract](#execution-contract)
12. [Reference](#reference)

---

# Suite structure

Format spec for test suites consumed by `kuma-manual-test`.

## Suite naming

Suite names follow the pattern `{feature}-{scope}` in kebab-case:

- `{feature}`: the primary resource or feature being tested (e.g., `meshmetric`, `meshtrace`, `motb`)
- `{scope}`: what aspect is covered (e.g., `core`, `pipe-mode`, `multizone`, `backendref-validation`)

Good: `meshmetric-core`, `motb-pipe-mode`, `meshtrace-otel-backends`, `delegated-gw-dataplane-targetref`
Bad: `test-suite-1`, `full`, `feature-branch`, `my-test`

For single-feature suites testing the full surface, use `{feature}-core`. For focused suites testing a specific aspect, name the aspect explicitly.

## Suite directory layout

```
${DATA_DIR}/suites/${SUITE_NAME}/
  suite.md                   # metadata, group table, execution contract (~80-130 lines)
  baseline/                  # shared manifests applied before G1
    namespace.yaml
    otel-collector.yaml
    demo-workload.yaml
  groups/                    # one file per group (or per range)
    g01-crud.md
    g02-validation.md
    ...
```

## suite.md structure

The entry point file (~80-130 lines) contains metadata and tables that reference the group files. Sections:

### TOC

Links to each section within suite.md.

### Suite metadata

```markdown
## Suite metadata

- suite id: <kebab-case-name>
- session_id: <session ID from kuma-suite-author, or "standalone">
- feature scope: <what this tests>
- target environments: single-zone / multi-zone / universal
- required dependencies: <workloads, collectors, etc.>
- skipped groups: <group IDs not included, with reason>
- keep_clusters: false (set true to skip cluster teardown after the run)
```

### Baseline manifests table

```markdown
## Baseline manifests

| File                       | Purpose                        |
| :------------------------- | :----------------------------- |
| baseline/namespace.yaml    | test namespace with mesh label |
| baseline/otel-collector.yaml | otel collector deployment    |
| baseline/demo-workload.yaml  | echo server + client pods    |
```

### Test groups table

```markdown
## Test groups

| Group   | File                       | Goal                    | Minimum artifacts              |
| :------ | :------------------------- | :---------------------- | :----------------------------- |
| G1      | groups/g01-crud.md         | Resource CRUD           | create/get/update/delete YAML  |
| G2      | groups/g02-validation.md   | Validation rejects      | admission errors               |
| ...     | ...                        | ...                     | ...                            |
```

### Execution contract

See [Execution contract](#execution-contract) below.

### Failure triage

Short section referencing `references/agent-contract.md` for the full procedure.

## Baseline directory

One `.yaml` file per shared resource applied before G1. These are manifests that multiple groups depend on (namespace setup, otel collector, demo workloads). Extract them from group steps to avoid duplication.

## Groups directory

One file per group. Naming convention: `g{NN}-{slug}.md` where NN is zero-padded and slug is kebab-case. Range groups use: `g17-g26-pipe-mode.md`.

## Group file structure

Each group file (~30-80 lines) contains:

### Heading

```markdown
# G{N} - {Goal}
```

### Signal tag (variant groups only)

```markdown
[S3 backend variant]
```

### Prerequisites (optional)

Any setup specific to this group beyond the baseline manifests.

### Steps

Each step includes:

- Inline YAML manifests in fenced code blocks. These are **authoritative** - `kuma-manual-test` will use them verbatim without modification. Make sure every field is correct (name, namespace, labels, spec) because the test runner is not allowed to silently fix them.
- Validation commands to run (kumactl inspect, curl, kubectl get, etc.)
- Expected results stated clearly
- Cleanup commands if the step requires deleting resources before the next step

### Artifacts

List of artifacts to capture for this group.

## Standard group structure

| Group | Purpose             | Typical contents                                              |
| :---- | :------------------ | :------------------------------------------------------------ |
| G1    | CRUD baseline       | create/get/update/delete the resource                         |
| G2    | Validation failures | invalid manifests that should be rejected (from validator.go) |
| G3    | Runtime config      | xDS inspection commands (from plugin.go understanding)        |
| G4    | End-to-end flow     | traffic generation + expected behavior                        |
| G5    | Edge cases          | dangling refs, missing deps, bad combinations                 |
| G6    | Multi-zone          | KDS sync, cross-zone, cross-mesh isolation                    |
| G7    | Backward compat     | legacy paths, deprecated fields, migration behavior           |

Not all groups apply to every feature. Skip groups that don't make sense, but document why in the suite metadata.

## Manifest conventions

- `apiVersion`: use the correct group/version from the CRD (e.g., `kuma.io/v1alpha1`)
- `metadata.namespace`: `kuma-system` for mesh-scoped resources, workload namespace for namespace-scoped
- `metadata.labels`: include `kuma.io/mesh: <mesh-name>` where required
- `metadata.annotations`: include `kuma.io/mesh: <mesh-name>` for universal resources
- Use realistic but minimal manifests - enough to trigger the behavior, no extras

## Domain knowledge

### Delegated gateways

A "delegated gateway" in Kuma is a standalone gateway proxy (not managed by Kuma's builtin gateway) that Kuma treats as a gateway dataplane. In practice this means Kong Gateway. When generating test suites that involve delegated gateways:

- Use Kong Gateway (image `kong:3.9` or similar) as the delegated gateway workload, not nginx or a generic proxy
- Deploy Kong in its own namespace (`kong`) with `kuma.io/sidecar-injection: enabled`
- Annotate the pod with `kuma.io/gateway: enabled` so the injector treats it as a delegated gateway
- Label the pod `app: kong-gateway` to match the convention used in unit test fixtures
- Configure Kong in DB-less mode (`KONG_DATABASE=off`) with declarative config routing to backend services
- The resulting Dataplane resource will have `networking.gateway.type: DELEGATED`
- Policies target delegated gateways via `kind: Dataplane` with label selectors (not `kind: MeshGateway` which is for builtin gateways)

### Builtin vs delegated

| Aspect | Builtin gateway | Delegated gateway (Kong) |
| :--- | :--- | :--- |
| Managed by | Kuma (MeshGateway + GatewayClass) | External (Kong, deployed by user) |
| Dataplane type | `BUILTIN` | `DELEGATED` |
| Policy targeting | `kind: MeshGateway` | `kind: Dataplane` with labels |
| Pod annotation | none (auto-created) | `kuma.io/gateway: enabled` |
| Transparent proxy | disabled | disabled |

## Validation step patterns

Commands to verify expected state after applying manifests:

```bash
# Resource exists
kubectl get <resource-type> <name> -n <namespace> -o yaml

# kumactl inspection
"${KUMACTL}" inspect dataplanes --mesh default

# Envoy config dump (specific section)
kubectl exec deploy/<name> -c kuma-sidecar -- \
  wget -qO- localhost:9901/config_dump | \
  jq '.configs[] | select(."@type" | contains("<Section>"))'

# Control plane logs
kubectl logs -n kuma-system deploy/kuma-control-plane --tail=50
```

## Artifact capture patterns

| Group type      | What to capture                                 |
| :-------------- | :---------------------------------------------- |
| CRUD            | resource YAML before/after, kubectl output      |
| Validation      | admission error messages                        |
| Runtime config  | config dump snippets for relevant xDS sections  |
| E2E flow        | traffic tool output, collector/backend logs     |
| Edge cases      | CP logs, resource status, error messages        |
| Multi-zone      | resource presence on global and zone CPs        |
| Backward compat | deprecation warnings, runtime config comparison |

## Execution contract

Every suite must include this checklist in suite.md:

- all manifests applied through `"${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh"`
- all commands (inspect, curl, delete, kubectl get, etc.) recorded via `"${CLAUDE_SKILL_DIR}/scripts/record-command.sh"`
- cluster state captured after each completed group via `"${CLAUDE_SKILL_DIR}/scripts/capture-state.sh"`
- `run-status.yaml` updated after each group with counts and last_completed_group
- all failures trigger immediate triage before next group
- all pass/fail decisions include artifact pointers to existing files
- deviations from suite definitions require user approval and are recorded in the report
- inline manifests in group files are authoritative - the test runner must use them verbatim
- edge cases from `references/mesh-policies.md` included when suite tests Mesh\* policies

## Reference

- Suite directory format: described in this file
- Example suite: `examples/example-motb-core-suite.md`
- Edge case matrix: `kuma-manual-test` skill's `references/mesh-policies.md`
