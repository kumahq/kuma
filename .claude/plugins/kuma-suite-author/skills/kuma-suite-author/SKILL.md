---
name: kuma-suite-author
description: >-
  Generate test suites for kuma-manual-test by reading Kuma source code.
  Produces ready-to-run suites with manifests, validation steps, and expected outcomes.
argument-hint: "<feature-name> [--repo /path/to/kuma] [--mode generate|wizard] [--from-pr PR_URL] [--from-branch BRANCH]"
allowed-tools: Read, Write, Edit, Bash, Glob, Grep, AskUserQuestion
user-invocable: true
---

# Kuma suite author

Generate test suites for `kuma-manual-test` by reading Kuma source code. Produces ready-to-run suites with inline manifests, validation steps, and expected outcomes. Detects feature variants automatically and confirms the full suite before saving.

## Arguments

Parse from `$ARGUMENTS`:

| Argument        | Default              | Purpose                                                     |
| :-------------- | :------------------- | :---------------------------------------------------------- |
| (positional)    | -                    | Feature or policy name (e.g., `meshretry`, `meshtrace`)     |
| `--repo`        | auto-detect cwd      | Path to Kuma repo checkout                                  |
| `--mode`        | `generate`           | `generate` (full AI) or `wizard` (interactive step-by-step) |
| `--from-pr`     | -                    | GitHub PR URL to scope the feature from                     |
| `--from-branch` | -                    | Git branch to diff against master for scope                 |
| `--suite-name`  | derived from feature | Override output filename                                    |

## Workflow - generate mode (default)

### Step 1: Resolve paths

```bash
DATA_DIR="$(echo "${XDG_DATA_HOME:-$HOME/.local/share}/sai/kuma-manual-test")"
mkdir -p "${DATA_DIR}/suites" "${DATA_DIR}/runs"
```

Resolve `REPO_ROOT`: `--repo` flag > check if cwd has `go.mod` with `kumahq/kuma` > fail with message.

### Step 2: Check worktree and branch

Before scoping the feature, verify the user is in the right repo location.

Run `git rev-parse --show-toplevel` and `git branch --show-current` in `REPO_ROOT` to detect the current repo root and branch.

**First question** - ask if current location is correct. Use AskUserQuestion with options:

- "Yes, correct" - show current path and branch in the description (e.g., `~/Projects/kuma on feat/meshretry`)
- "No, wrong location" - need to switch worktree or branch

If user confirms, continue to step 3.

**If wrong location** - ask what to switch. Use AskUserQuestion with options:

- "Switch worktree"
- "Switch branch"

**Worktree flow**: if the user provided a worktree name in their "Other" response, use it directly. Otherwise run `git worktree list`, parse the output, and present available worktrees via AskUserQuestion (each option shows path and branch). After selection, update `REPO_ROOT` to the chosen worktree path.

**Branch flow**: if the user provided a branch name in their "Other" response, use it directly. Otherwise run `git branch --list`, present local branches via AskUserQuestion. After selection, run `git checkout <branch>` in `REPO_ROOT`.

### Step 3: Scope the feature

Identify what code to read based on the input:

- **From feature name**: find policy dir in `pkg/plugins/policies/`, API spec, plugin.go, tests.
- **From PR URL**: run `gh pr diff <number> --repo kumahq/kuma` to identify changed files.
- **From branch**: run `git diff master...<branch> --name-only` to identify changed files.

Handle ambiguity with AskUserQuestion:

- Multiple policy dirs match the feature name - ask which one to use.
- Feature type unclear (policy vs non-policy) - ask the user.
- PR diff touches files outside the expected scope - ask whether to include them.

### Step 4: Read code

Read `references/code-reading-guide.md` for where to look in the Kuma repo.
Read `references/variant-detection.md` for variant signal patterns.

For each identified file, read and extract two kinds of data:

**Group material** (feeds G1-G7):

- **Policy API spec** (`api/v1alpha1/<policy>.go`): struct fields, markers, validation constraints.
- **Plugin implementation** (`plugin/v1alpha1/plugin.go`): xDS generation logic, which resource types are affected.
- **Existing tests** (`plugin/v1alpha1/testdata/`): golden files show expected Envoy configs.
- **Validator** (`api/v1alpha1/validator.go`): what inputs are rejected and why.
- **Non-policy features**: read relevant `pkg/` code based on changed files list.

**Variant signals** (patterns from variant-detection.md):

- S1: KDS markers, resource registration for deployment topology
- S2: Enum/string fields, switch/case blocks for feature modes
- S3: Multiple backend types, backendRef kinds for backend variants
- S4: Conditional Apply() branches for feature flags
- S5: targetRef section, producer/consumer markers for policy roles
- S6: HTTP/TCP/gRPC branching for protocol variants
- S7: deprecated.go, old field names for backward compat paths

Collect each signal with its source file, evidence, and estimated strength (strong/moderate/weak).

### Step 5: Detect and confirm variants

Build a variant list from collected signals. Each entry has: id, signal type, source file, description, suggested group count, strength.

Use AskUserQuestion with multiSelect to present detected variants. Group by strength:

- **Strong signals** (pre-selected): distinct code paths, different xDS output, separate golden files. Recommend including these.
- **Moderate signals** (tagged `[uncertain]`): present separately, explain the evidence, let user decide.
- **Weak signals**: mention in the question description but don't offer as selectable options.

If no variants detected, note it and continue to step 6 with G1-G7 only.

### Step 6: Generate suite

Read `references/suite-structure.md` for the format spec.

Build the suite with base groups (skip groups that don't apply, document why):

| Group              | Source                             | Contents                                                  |
| :----------------- | :--------------------------------- | :-------------------------------------------------------- |
| G1 CRUD            | API spec struct                    | create/read/update/delete manifests with realistic values |
| G2 Validation      | validator.go                       | invalid manifests that trigger each rejection path        |
| G3 Runtime config  | plugin.go                          | xDS inspection commands based on what Apply() generates   |
| G4 E2E flow        | golden files + plugin logic        | traffic generation + expected behavior                    |
| G5 Edge cases      | validator edge cases, nil handling | dangling refs, missing deps, boundary values              |
| G6 Multi-zone      | KDS markers, sync config           | global-to-zone presence checks                            |
| G7 Backward compat | deprecated.go, old fields          | legacy path behavior, deprecation warnings                |

Then add G8+ for selected variants. Variant groups number sequentially from G8. Use range notation for multi-group variants (e.g., `G17-G26 Pipe mode`).

For G6 multi-zone: if no KDS markers found and no multi-zone variant was selected in step 5, use AskUserQuestion to confirm skipping multi-zone groups.

The suite is split into three buckets:

- `suite.md` - metadata, baseline table, group table, execution contract
- `baseline/*.yaml` - shared manifests (namespace setup, otel collector, demo workloads) extracted from groups
- `groups/g{NN}-*.md` - one file per group with steps, manifests, validation commands, artifacts

For each group (base and variant):

- Generate actual YAML manifests inline in the group file.
- Include specific validation commands (kubectl, kumactl, config_dump).
- State expected outcomes clearly.
- List artifacts to capture.
- Tag variant groups with their source signal (e.g., `[S3 backend variant]`).

### Step 7: Confirmation wizard

Before saving, present the full suite summary via AskUserQuestion.

Show in the question description:

- Suite metadata: id, scope, environments, dependencies
- All groups with one-line descriptions and artifact counts
- Variant tags on applicable groups (G8+)
- Total group count

Options:

- "Confirm and save"
- "Add a group"
- "Remove a group"
- "Edit a group"

If user picks add/remove/edit: handle the change, then present the summary again. Loop until user confirms.

### Step 8: Save suite

```bash
SUITE_NAME="${SUITE_NAME:-<derived-from-feature>}"
SUITE_DIR="${DATA_DIR}/suites/${SUITE_NAME}"
mkdir -p "${SUITE_DIR}/baseline" "${SUITE_DIR}/groups"
```

Write each part separately:

- `${SUITE_DIR}/suite.md` - metadata, baseline table, group table, execution contract
- `${SUITE_DIR}/baseline/*.yaml` - one file per shared manifest (namespace, otel collector, demo workloads)
- `${SUITE_DIR}/groups/g{NN}-{slug}.md` - one file per group (or per range, e.g., `g17-g26-pipe-mode.md`)

### Step 9: Report

Print the saved path and suggest how to run it:

```
Suite saved to: ${SUITE_DIR}/
Run with: /kuma-manual-test ${SUITE_NAME} --repo ${REPO_ROOT}
```

## Workflow - wizard mode

Interactive step-by-step suite generation:

1. Same path resolution as generate mode (step 1).
2. Check worktree and branch (step 2) - same verification flow.
3. Ask feature name, target environment, scope using AskUserQuestion.
4. Read code and collect variant signals (step 4).
5. Detect and confirm variants (step 5) - present each signal individually for review.
6. Show the group structure from `references/suite-structure.md`, ask which base groups (G1-G7) to include.
7. For each selected group: ask what to test, generate manifests, show for review.
8. User edits/approves each group before moving to next.
9. Confirmation wizard (step 7) - same full summary review before saving.
10. Save and report same as generate mode.

## Bundled resources

- `references/code-reading-guide.md` - where to find policy specs, xDS generators, tests in a Kuma repo
- `references/variant-detection.md` - variant signal catalog, strength classification, group mapping
- `references/suite-structure.md` - suite format spec, group structure, manifest conventions
- `examples/example-motb-core-suite.md` - worked example of a complete test suite

## Example invocations

```bash
# Generate suite for MeshRetry policy
/kuma-suite-author meshretry --repo ~/Projects/kuma

# Generate from a PR
/kuma-suite-author meshexternalservice --from-pr https://github.com/kumahq/kuma/pull/15571

# Generate from a branch
/kuma-suite-author motb --from-branch feat/implement-motb --repo ~/Projects/kuma

# Interactive wizard mode
/kuma-suite-author meshtrace --mode wizard --repo ~/Projects/kuma

# Custom suite name
/kuma-suite-author meshretry --suite-name meshretry-timeout-edge-cases
```
