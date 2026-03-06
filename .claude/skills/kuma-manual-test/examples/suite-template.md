# Suite template - directory format

Template for `suite.md` in a directory-format suite. Copy the directory structure and fill in the blanks.

## Directory structure

```
suites/<suite-name>/
├── suite.md           # this file
├── baseline/
│   └── *.yaml         # shared manifests applied before G1
└── groups/
    └── g{NN}-*.md     # one file per group
```

## suite.md template

### Suite metadata

- suite id:
- feature scope:
- target environments: single-zone / multi-zone / universal
- required dependencies:

### Baseline manifests

| File                  | Purpose                       |
| --------------------- | ----------------------------- |
| baseline/namespace.yaml | Create namespace with sidecar |
| (add more rows)       | (describe purpose)            |

### Group structure

| Group | File                 | Goal                          | Minimum artifacts          |
| ----- | -------------------- | ----------------------------- | -------------------------- |
| G1    | groups/g01-crud.md   | CRUD baseline                 |                            |
| G2    | groups/g02-valid.md  | Validation failures           |                            |
| G3    | groups/g03-xds.md    | Runtime config verification   |                            |
| G4    | groups/g04-e2e.md    | End-to-end flow               |                            |
| G5    | groups/g05-edge.md   | Edge cases and negative paths |                            |
| G6    | groups/g06-mz.md     | Multi-zone and isolation      |                            |
| G7    | groups/g07-compat.md | Backward compatibility        |                            |

### Execution contract

- All manifests applied through `"${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh"`
- All failures trigger immediate triage before next group
- All pass/fail decisions include artifact pointers
- Include edge cases from `references/mesh-policies.md` when suite uses Mesh\* policies

### Failure triage

See `references/agent-contract.md` (failure policy and bug triage protocol) for the full procedure.

## Group file template

Each group file follows this structure:

```markdown
# G{N} - Group name

## Prerequisites

- (what must be true before this group starts)

## Steps

### S{N}.1 - Step name

- manifest: (inline YAML or path)
- command: (kubectl / kumactl command)
- expected: (what the output should contain)
- artifacts: (what to save)

### S{N}.2 - Step name

- manifest:
- command:
- expected:
- artifacts:
```

## Legacy single-file format

For simple suites that don't need progressive loading, a single `<suite-name>.md` file in `suites/` still works. Put all group details inline. The runner detects the format automatically based on whether the path resolves to a directory or a file.
