# Contents

1. [Suite directory layout](#suite-directory-layout)
2. [suite.md structure](#suitemd-structure)
3. [Baseline directory](#baseline-directory)
4. [Groups directory](#groups-directory)
5. [Group file structure](#group-file-structure)
6. [Standard group structure](#standard-group-structure)
7. [Manifest conventions](#manifest-conventions)
8. [Validation step patterns](#validation-step-patterns)
9. [Artifact capture patterns](#artifact-capture-patterns)
10. [Execution contract](#execution-contract)
11. [Reference](#reference)

---

# Suite structure

Format spec for test suites consumed by `kuma-manual-test`.

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
- feature scope: <what this tests>
- target environments: single-zone / multi-zone / universal
- required dependencies: <workloads, collectors, etc.>
- skipped groups: <group IDs not included, with reason>
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

- Inline YAML manifests in fenced code blocks
- Validation commands to run
- Expected results stated clearly

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

- all manifests applied through `"$SKILL_DIR/scripts/apply-tracked-manifest.sh"`
- all failures trigger immediate triage before next group
- all pass/fail decisions include artifact pointers
- edge cases from `references/mesh-policies.md` included when suite tests Mesh\* policies

## Reference

- Suite directory format: described in this file
- Example suite: `examples/example-motb-core-suite.md`
- Edge case matrix: `kuma-manual-test` skill's `references/mesh-policies.md`
